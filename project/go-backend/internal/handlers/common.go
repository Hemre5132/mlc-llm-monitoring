package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const AppVersion = "0.1.0"

type CommonHandler struct {
	DB *gorm.DB
}

func NewCommonHandler(db *gorm.DB) *CommonHandler {
	return &CommonHandler{DB: db}
}

// ---------- 19) GET /api/health ----------
// Render'ın health check probe'u bu endpoint'i düzenli olarak yoklar.

func (h *CommonHandler) Health(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	dbOK := err == nil && sqlDB.Ping() == nil

	status := http.StatusOK
	if !dbOK {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": map[bool]string{true: "ok", false: "degraded"}[dbOK],
		"db":     dbOK,
	})
}

// ---------- 20) GET /api/version ----------

func (h *CommonHandler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": AppVersion})
}

// ---------- 21) GET /api/stats/dashboard ----------
// Dashboard view'ının besleneceği genel (tüm sistem) istatistikler.

func (h *CommonHandler) DashboardStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "yetkisiz erişim"})
		return
	}

	uid := userID.(uuid.UUID)
	var sessionCount, messageCount, scoredMessages int64
	var avgScore float64

	if err := h.DB.Table("llm_sessions").Where("user_id = ?", uid).Count(&sessionCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturum istatistikleri alınamadı"})
		return
	}
	if err := h.DB.Table("llm_messages").Joins("JOIN llm_sessions ON llm_sessions.id = llm_messages.session_id").Where("llm_sessions.user_id = ?", uid).Count(&messageCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mesaj istatistikleri alınamadı"})
		return
	}
	if err := h.DB.Table("llm_scores").Joins("JOIN llm_messages ON llm_messages.id = llm_scores.message_id").Joins("JOIN llm_sessions ON llm_sessions.id = llm_messages.session_id").Where("llm_sessions.user_id = ?", uid).Select("COALESCE(AVG(llm_scores.score), 0)").Scan(&avgScore).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skor istatistikleri alınamadı"})
		return
	}
	if err := h.DB.Table("llm_scores").Joins("JOIN llm_messages ON llm_messages.id = llm_scores.message_id").Joins("JOIN llm_sessions ON llm_sessions.id = llm_messages.session_id").Where("llm_sessions.user_id = ?", uid).Count(&scoredMessages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skor sayısı alınamadı"})
		return
	}

	var sessions []struct {
		ID             string  `json:"id"`
		Title          string  `json:"title"`
		ModelName      string  `json:"model_name"`
		AverageScore   float64 `json:"average_score"`
		ScoredMessages int64   `json:"scored_messages"`
	}

	err := h.DB.Table("llm_sessions as s").
		Select("s.id, s.title, s.model_name, COALESCE(AVG(ls.score), 0) as average_score, COUNT(ls.id) as scored_messages").
		Where("s.user_id = ?", uid).
		Joins("LEFT JOIN llm_messages as m ON m.session_id = s.id").
		Joins("LEFT JOIN llm_scores as ls ON ls.message_id = m.id").
		Group("s.id, s.title, s.model_name, s.created_at").
		Order("s.created_at desc").
		Scan(&sessions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturum skorları alınamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_sessions":  sessionCount,
		"total_messages":  messageCount,
		"average_score":   avgScore,
		"scored_messages": scoredMessages,
		"sessions":        sessions,
	})
}

// ---------- 22) GET /api/stats/me ----------
// Giriş yapmış kullanıcıya özel istatistikler.

func (h *CommonHandler) MyStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var sessionCount, scoredMessages int64
	var avgScore float64

	if err := h.DB.Table("llm_sessions").Where("user_id = ?", userID).Count(&sessionCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kişisel oturum istatistikleri alınamadı"})
		return
	}
	if err := h.DB.Table("llm_scores").
		Joins("JOIN llm_messages ON llm_messages.id = llm_scores.message_id").
		Joins("JOIN llm_sessions ON llm_sessions.id = llm_messages.session_id").
		Where("llm_sessions.user_id = ?", userID).
		Select("COALESCE(AVG(llm_scores.score), 0)").
		Scan(&avgScore).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kişisel skor istatistikleri alınamadı"})
		return
	}
	if err := h.DB.Table("llm_scores").
		Joins("JOIN llm_messages ON llm_messages.id = llm_scores.message_id").
		Joins("JOIN llm_sessions ON llm_sessions.id = llm_messages.session_id").
		Where("llm_sessions.user_id = ?", userID).
		Count(&scoredMessages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kişisel skor sayısı alınamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_count":   sessionCount,
		"average_score":   avgScore,
		"scored_messages": scoredMessages,
	})
}
