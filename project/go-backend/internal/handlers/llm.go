package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"

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

	// 1. Önce oturuma ait yetkinin doğrulanması için oturumu bul
	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı veya yetkiniz yok"})
		return
	}

	// 2. Bu oturumdaki mesajlara ait olan skorları sil (Raw SQL ile alt sorgu)
	h.DB.Exec("DELETE FROM llm_scores WHERE message_id IN (SELECT id FROM llm_messages WHERE session_id = ?)", sessionID)

	// 3. Bu oturuma ait tüm mesajları sil
	h.DB.Where("session_id = ?", sessionID).Delete(&models.LLMMessage{})

	// 4. Artık bağlı hiçbir alt kayıt kalmadığı için oturumu güvenle silebiliriz
	result := h.DB.Where("id = ?", sessionID).Delete(&models.LLMSession{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturum silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "oturum ve bağlı tüm veriler başarıyla silindi"})
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

	var savedScore *models.LLMScore
	if req.Role == "assistant" {
		var savedMessage models.LLMMessage
		if err := h.DB.Where("id = ?", message.ID).First(&savedMessage).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "kaydedilen mesaj bulunamadı"})
			return
		}

		var scoreErr error
		savedScore, scoreErr = h.scoreAssistantMessage(savedMessage)
		if scoreErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "puanlama başarısız oldu"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": message, "score": savedScore})
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

	score, err := h.calculateScore(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var existing models.LLMScore
	err = h.DB.Where("message_id = ?", message.ID).First(&existing).Error
	if err == nil {
		existing.Score = score.Score
		existing.Criteria = score.Criteria
		if saveErr := h.DB.Save(&existing).Error; saveErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "skor güncellenemedi"})
			return
		}
		c.JSON(http.StatusOK, existing)
		return
	}
	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skor sorgulanamadı"})
		return
	}

	if saveErr := h.DB.Create(&score).Error; saveErr != nil {
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

// BackfillScores creates scores only for existing assistant messages that do
// not have a score yet. Existing scores are left unchanged.
func (h *LLMHandler) BackfillScores(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var messages []models.LLMMessage
	if err := h.DB.Where("role = ? AND session_id IN (?)", "assistant",
		h.DB.Table("llm_sessions").Select("id").Where("user_id = ?", userID),
	).Where("id NOT IN (?)", h.DB.Table("llm_scores").Select("message_id")).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "eksik skorlar bulunamadı"})
		return
	}

	created := 0
	for _, message := range messages {
		if _, err := h.scoreAssistantMessage(message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "eksik skorlar tamamlanamadı"})
			return
		}
		created++
	}

	c.JSON(http.StatusOK, gin.H{"created": created})
}

func (h *LLMHandler) scoreAssistantMessage(message models.LLMMessage) (*models.LLMScore, error) {
	score, err := h.calculateScore(message)
	if err != nil {
		return nil, err
	}

	var existing models.LLMScore
	err = h.DB.Where("message_id = ?", message.ID).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err := h.DB.Create(&score).Error; err != nil {
		return nil, err
	}

	return &score, nil
}

func (h *LLMHandler) calculateScore(message models.LLMMessage) (models.LLMScore, error) {
	content := strings.ToLower(strings.TrimSpace(message.Content))
	rawOutput := strings.ToLower(strings.TrimSpace(message.RawOutput))

	lengthScore := 0.0
	if len(content) > 0 {
		lengthScore = math.Min(25, float64(len(content))/8)
	}

	coherenceScore := 0.0
	if len(content) >= 20 {
		coherenceScore = 70 + math.Min(20, float64(len(content))/20)
	} else {
		coherenceScore = 40 + math.Min(20, float64(len(content))/10)
	}

	responseQuality := 0.0
	if strings.Contains(content, "?") {
		responseQuality += 5
	}
	if len(content) > 40 {
		responseQuality += 10
	}
	if len(rawOutput) > 0 {
		responseQuality += 10
	}
	if message.LatencyMs > 0 && message.LatencyMs < 5000 {
		responseQuality += 10
	}
	if message.TokenCount > 0 {
		responseQuality += 5
	}
	responseQuality = math.Min(25, responseQuality)

	safetyScore := 0.0
	if !strings.Contains(content, "ignore") && !strings.Contains(content, "hack") && !strings.Contains(content, "exploit") {
		safetyScore = 95
	}

	accuracyScore := 0.0
	if len(content) >= 20 {
		accuracyScore = 75
	} else {
		accuracyScore = 55
	}
	if len(rawOutput) > 0 {
		accuracyScore = math.Min(100, accuracyScore+5)
	}

	criteria := map[string]float64{
		"coherence":        coherenceScore,
		"safety":           safetyScore,
		"accuracy":         accuracyScore,
		"length":           lengthScore,
		"response_quality": responseQuality,
	}

	overall := (criteria["coherence"] + criteria["safety"] + criteria["accuracy"] + criteria["length"] + criteria["response_quality"]) / 5

	criteriaJSON, _ := json.Marshal(criteria)
	return models.LLMScore{
		MessageID: message.ID,
		Score:     math.Round(overall*10) / 10,
		Criteria:  string(criteriaJSON),
	}, nil
}
