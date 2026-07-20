package handlers

import (
	"net/http"
	"time"

	"masterfabric-backend/internal/config"
	"masterfabric-backend/internal/models"
	"masterfabric-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{DB: db, Cfg: cfg}
}

// ---------- 1) POST /api/auth/register ----------

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "bu email zaten kayıtlı"})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "şifre işlenemedi"})
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		VerifyToken:  uuid.NewString(),
	}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kullanıcı oluşturulamadı"})
		return
	}

	// NOT: Gerçek projede burada email gönderim servisi (SES/SendGrid vb.)
	// tetiklenir. Demo günü için verify token'ı response'ta dönmek yeterli.
	c.JSON(http.StatusCreated, gin.H{
		"message":      "kayıt başarılı, email doğrulaması gerekli",
		"user_id":      user.ID,
		"verify_token": user.VerifyToken,
	})
}

// ---------- 2) POST /api/auth/login ----------

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email veya şifre hatalı"})
		return
	}
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email veya şifre hatalı"})
		return
	}

	access, refresh, err := h.issueTokenPair(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token üretilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"user":          user,
	})
}

// ---------- 3) POST /api/auth/logout ----------

func (h *AuthHandler) Logout(c *gin.Context) {
	// NOT: Stateless JWT kullanıldığı için gerçek bir "iptal" işlemi yapmak
	// istersen refresh token'ları DB'de bir blacklist tablosunda tutup
	// burada işaretlemen gerekir. Demo kapsamı için client'ın token'ı
	// silmesi yeterli; burada 200 dönüyoruz.
	c.JSON(http.StatusOK, gin.H{"message": "çıkış yapıldı"})
}

// ---------- 4) POST /api/auth/refresh ----------

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ParseToken(req.RefreshToken, h.Cfg.JWTSecret)
	if err != nil || claims.Type != utils.RefreshToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "geçersiz refresh token"})
		return
	}

	access, err := utils.GenerateToken(claims.UserID, utils.AccessToken, h.Cfg.JWTSecret,
		time.Duration(h.Cfg.JWTAccessTTLMin)*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token üretilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": access})
}

// ---------- 5) POST /api/auth/forgot-password ----------

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Kullanıcı bulunamasa bile aynı mesajı dönüyoruz — email enumeration
	// saldırılarına karşı standart bir güvenlik pratiği.
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err == nil {
		resetToken := uuid.NewString()
		exp := time.Now().Add(1 * time.Hour)
		h.DB.Model(&user).Updates(models.User{ResetToken: resetToken, ResetTokenExp: &exp})
		// NOT: Burada email gönderim servisi tetiklenir.
	}

	c.JSON(http.StatusOK, gin.H{"message": "eğer bu email kayıtlıysa sıfırlama linki gönderildi"})
}

// ---------- 6) POST /api/auth/reset-password ----------

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("reset_token = ?", req.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz token"})
		return
	}
	if user.ResetTokenExp == nil || time.Now().After(*user.ResetTokenExp) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token süresi dolmuş"})
		return
	}

	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "şifre işlenemedi"})
		return
	}

	h.DB.Model(&user).Updates(map[string]interface{}{
		"password_hash":   hash,
		"reset_token":     "",
		"reset_token_exp": nil,
	})

	c.JSON(http.StatusOK, gin.H{"message": "şifre başarıyla güncellendi"})
}

// ---------- 7) POST /api/auth/verify-email ----------

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("verify_token = ?", req.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz doğrulama token'ı"})
		return
	}

	h.DB.Model(&user).Updates(map[string]interface{}{
		"email_verified": true,
		"verify_token":   "",
	})

	c.JSON(http.StatusOK, gin.H{"message": "email doğrulandı"})
}

// ---------- 8) GET /api/auth/me ----------

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var user models.User
	if err := h.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "kullanıcı bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ---------- yardımcı ----------

func (h *AuthHandler) issueTokenPair(userID uuid.UUID) (string, string, error) {
	access, err := utils.GenerateToken(userID, utils.AccessToken, h.Cfg.JWTSecret,
		time.Duration(h.Cfg.JWTAccessTTLMin)*time.Minute)
	if err != nil {
		return "", "", err
	}
	refresh, err := utils.GenerateToken(userID, utils.RefreshToken, h.Cfg.JWTSecret,
		time.Duration(h.Cfg.JWTRefreshTTLHr)*time.Hour)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}
