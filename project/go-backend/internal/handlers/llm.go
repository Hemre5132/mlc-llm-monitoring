package handlers

import (
	"encoding/json"
	"net/http"

	"masterfabric-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LLMHandler struct {
	DB *gorm.DB
}

func NewLLMHandler(db *gorm.DB) *LLMHandler {
	return &LLMHandler{DB: db}
}

// ---------- 11) POST /api/llm/sessions ----------
// Tarayıcıda WebLLM ile yeni bir oturum başlatıldığında frontend bunu çağırır.

type createSessionRequest struct {
	ModelName string `json:"model_name" binding:"required"`
	Title     string `json:"title"`
}

func (h *LLMHandler) CreateSession(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session := models.LLMSession{
		UserID:    userID,
		ModelName: req.ModelName,
		Title:     req.Title,
	}
	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturum oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// ---------- 12) GET /api/llm/sessions ----------

func (h *LLMHandler) ListSessions(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var sessions []models.LLMSession
	if err := h.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturumlar getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// ---------- 13) GET /api/llm/sessions/:id ----------

func (h *LLMHandler) GetSession(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// ---------- 14) DELETE /api/llm/sessions/:id ----------

func (h *LLMHandler) DeleteSession(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	result := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&models.LLMSession{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturum silinemedi"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "oturum silindi"})
}

// ---------- 15) POST /api/llm/sessions/:id/messages ----------
// Tarayıcıda üretilen ham (raw) prompt/response burada loglanır — "Raw LLM
// Monitoring" gereksiniminin karşılandığı ana endpoint budur.

type createMessageRequest struct {
	Role       string `json:"role" binding:"required,oneof=user assistant"`
	Content    string `json:"content" binding:"required"`
	RawOutput  string `json:"raw_output"`
	LatencyMs  int    `json:"latency_ms"`
	TokenCount int    `json:"token_count"`
}

func (h *LLMHandler) CreateMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	var req createMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sid, _ := uuid.Parse(sessionID)
	message := models.LLMMessage{
		SessionID:  sid,
		Role:       req.Role,
		Content:    req.Content,
		RawOutput:  req.RawOutput,
		LatencyMs:  req.LatencyMs,
		TokenCount: req.TokenCount,
	}
	if err := h.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mesaj kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

// ---------- 16) GET /api/llm/sessions/:id/messages ----------

func (h *LLMHandler) ListMessages(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	var messages []models.LLMMessage
	if err := h.DB.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mesajlar getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// ---------- 17) POST /api/llm/sessions/:id/score ----------
// "Deci.Scoring": bir mesaj için karar puanı hesaplar ve kaydeder.
// NOT: Skorlama algoritmasının kendisi ayrı bir adımda detaylandırılacak;
// burada basit bir placeholder hesaplama var — gerçek kriterler
// (tutarlılık/güvenlik/doğruluk) bir sonraki aşamada bu fonksiyonun içine
// eklenecek.

type createScoreRequest struct {
	MessageID string `json:"message_id" binding:"required"`
}

func (h *LLMHandler) CreateScore(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	var req createScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var message models.LLMMessage
	if err := h.DB.Where("id = ? AND session_id = ?", req.MessageID, sessionID).First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mesaj bulunamadı"})
		return
	}

	// TODO: gerçek Deci.Scoring algoritması burada çalışacak.
	criteria := map[string]float64{"coherence": 80, "safety": 95, "accuracy": 75}
	overall := (criteria["coherence"] + criteria["safety"] + criteria["accuracy"]) / 3
	criteriaJSON, _ := json.Marshal(criteria)

	score := models.LLMScore{
		MessageID: message.ID,
		Score:     overall,
		Criteria:  string(criteriaJSON),
	}
	if err := h.DB.Save(&score).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skor kaydedilemedi"})
		return
	}

	c.JSON(http.StatusCreated, score)
}

// ---------- 18) GET /api/llm/sessions/:id/score ----------

func (h *LLMHandler) GetScores(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	var scores []models.LLMScore
	h.DB.Joins("JOIN llm_messages ON llm_messages.id = llm_scores.message_id").
		Where("llm_messages.session_id = ?", sessionID).
		Find(&scores)

	c.JSON(http.StatusOK, gin.H{"scores": scores})
}
