package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"masterfabric-backend/internal/llmclient"
	"masterfabric-backend/internal/metrics"
	"masterfabric-backend/internal/models"
	"masterfabric-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LLMHandler struct {
	DB     *gorm.DB
	Ollama *llmclient.OllamaClient
}

func NewLLMHandler(db *gorm.DB, ollama *llmclient.OllamaClient) *LLMHandler {
	return &LLMHandler{DB: db, Ollama: ollama}
}

// ---------- 11) POST /api/llm/sessions ----------
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

	metrics.LLMSessionsCreatedTotal.Inc()

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

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı veya yetkiniz yok"})
		return
	}

	h.DB.Exec("DELETE FROM llm_scores WHERE message_id IN (SELECT id FROM llm_messages WHERE session_id = ?)", sessionID)
	h.DB.Where("session_id = ?", sessionID).Delete(&models.LLMMessage{})
	result := h.DB.Where("id = ?", sessionID).Delete(&models.LLMSession{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oturum silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "oturum ve bağlı tüm veriler başarıyla silindi"})
}

// ---------- 15) POST /api/llm/sessions/:id/messages ----------
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

	// mlcmon_llm_messages_logged_total — her kaydedilen mesaj için
	metrics.LLMMessagesLoggedTotal.WithLabelValues(req.Role).Inc()

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "skorlama şu an yapılamadı, tekrar deneyin"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": message, "score": savedScore})
}

// ---------- POST /api/llm/sessions/:id/generate ----------
type generateChatRequest struct {
	Messages []llmclient.ChatMessage `json:"messages" binding:"required"`
}

func (h *LLMHandler) GenerateChat(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	var req generateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startedAt := time.Now()
	log.Printf("GenerateChat[%s]: %d messages, sending to Ollama", sessionID, len(req.Messages))
	reply, err := h.Ollama.Chat(req.Messages)
	if err != nil {
		log.Printf("GenerateChat[%s]: Ollama error: %v", sessionID, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "model yanıt üretemedi: " + err.Error()})
		return
	}
	log.Printf("GenerateChat[%s]: Ollama replied in %v", sessionID, time.Since(startedAt))
	latencyMs := int(time.Since(startedAt).Milliseconds())

	sid, _ := uuid.Parse(sessionID)
	message := models.LLMMessage{
		SessionID:  sid,
		Role:       "assistant",
		Content:    reply,
		RawOutput:  reply,
		LatencyMs:  latencyMs,
		TokenCount: estimateWordCount(reply),
	}
	if err := h.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mesaj kaydedilemedi"})
		return
	}

	// mlcmon_llm_messages_logged_total — generate sonucu kaydedilen assistant mesajı
	metrics.LLMMessagesLoggedTotal.WithLabelValues("assistant").Inc()

	score, err := h.scoreAssistantMessage(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skorlama şu an yapılamadı, tekrar deneyin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message, "score": score})
}

func estimateWordCount(text string) int {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	return len(words)
}

// ---------- POST /api/llm/sessions/:id/analyze ----------
type analyzeTextRequest struct {
	Text string `json:"text" binding:"required"`
}

func (h *LLMHandler) AnalyzeText(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	sessionID := c.Param("id")

	var session models.LLMSession
	if err := h.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "oturum bulunamadı"})
		return
	}

	var req analyzeTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := services.ScoreSalesScript(context.Background(), h.Ollama, req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "analiz şu an yapılamadı, tekrar deneyin"})
		return
	}

	// Kullanıcının metnini kaydet
	sid, _ := uuid.Parse(sessionID)
	message := models.LLMMessage{
		SessionID:  sid,
		Role:       "user",
		Content:    req.Text,
		TokenCount: estimateWordCount(req.Text),
	}
	if err := h.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mesaj kaydedilemedi"})
		return
	}
	metrics.LLMMessagesLoggedTotal.WithLabelValues("user").Inc()

	// Skoru kaydet
	criteria := map[string]float64{
		"opening_hook":        result.OpeningHook,
		"discovery":           result.Discovery,
		"value_proposition":   result.ValueProposition,
		"objection_handling":  result.ObjectionHandling,
		"closing_power":       result.ClosingPower,
		"persuasiveness_tone": result.PersuasivenessTone,
		"compliance_safety":   result.ComplianceSafety,
		"personalization":     result.Personalization,
		"structure_flow":      result.StructureFlow,
	}
	for criterion, val := range criteria {
		metrics.LLMScoreValue.WithLabelValues(criterion).Observe(val)
	}
	metrics.LLMScoreValue.WithLabelValues("overall").Observe(result.Overall)

	criteriaJSON, _ := json.Marshal(criteria)
	categoryScoresJSON, _ := json.Marshal(map[string]float64{
		"effectiveness": result.Effectiveness,
		"structure":     result.Structure,
		"safety":        result.ComplianceSafety,
	})

	score := models.LLMScore{
		MessageID:        message.ID,
		Score:            result.Overall,
		Criteria:         string(criteriaJSON),
		Category:         "sales_script",
		CategoryScores:   string(categoryScoresJSON),
		RequiresRevision: result.RequiresRevision,
	}
	if err := h.DB.Create(&score).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skor kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": result, "message_id": message.ID})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skorlama şu an yapılamadı, tekrar deneyin"})
		return
	}

	var existing models.LLMScore
	err = h.DB.Where("message_id = ?", message.ID).First(&existing).Error
	if err == nil {
		existing.Score = score.Score
		existing.Criteria = score.Criteria
		existing.Category = score.Category
		existing.CategoryScores = score.CategoryScores
		existing.RequiresRevision = score.RequiresRevision
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
		metrics.LLMBackfillMessagesTotal.WithLabelValues("failure").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "eksik skorlar bulunamadı"})
		return
	}

	created := 0
	for _, message := range messages {
		if _, err := h.scoreAssistantMessage(message); err != nil {
			metrics.LLMBackfillMessagesTotal.WithLabelValues("failure").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "skorlama şu an yapılamadı, tekrar deneyin"})
			return
		}
		created++
	}

	metrics.LLMBackfillMessagesTotal.WithLabelValues("success").Inc()

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

// calculateScore, asistan mesajının kalitesini satış scripti rubriğine göre puanlar.
// Sonuçları mlcmon_llm_score_value histogramına gönderir.
func (h *LLMHandler) calculateScore(message models.LLMMessage) (models.LLMScore, error) {
	calcStart := time.Now()
	defer func() {
		metrics.LLMScoreComputationDurationSeconds.Observe(time.Since(calcStart).Seconds())
	}()

	if h.Ollama == nil {
		return models.LLMScore{}, context.DeadlineExceeded
	}

	result, err := services.ScoreSalesScript(context.Background(), h.Ollama, message.Content)
	if err != nil {
		return models.LLMScore{}, err
	}

	criteria := map[string]float64{
		"opening_hook":        result.OpeningHook,
		"discovery":           result.Discovery,
		"value_proposition":   result.ValueProposition,
		"objection_handling":  result.ObjectionHandling,
		"closing_power":       result.ClosingPower,
		"persuasiveness_tone": result.PersuasivenessTone,
		"compliance_safety":   result.ComplianceSafety,
		"personalization":     result.Personalization,
		"structure_flow":      result.StructureFlow,
	}
	for criterion, val := range criteria {
		metrics.LLMScoreValue.WithLabelValues(criterion).Observe(val)
	}
	metrics.LLMScoreValue.WithLabelValues("overall").Observe(result.Overall)

	criteriaJSON, err := json.Marshal(criteria)
	if err != nil {
		return models.LLMScore{}, err
	}

	categoryScoresJSON, err := json.Marshal(map[string]float64{
		"effectiveness": result.Effectiveness,
		"structure":     result.Structure,
		"safety":        result.ComplianceSafety,
	})
	if err != nil {
		return models.LLMScore{}, err
	}

	return models.LLMScore{
		MessageID:        message.ID,
		Score:            result.Overall,
		Criteria:         string(criteriaJSON),
		Category:         "sales_script",
		CategoryScores:   string(categoryScoresJSON),
		RequiresRevision: result.RequiresRevision,
	}, nil
}
