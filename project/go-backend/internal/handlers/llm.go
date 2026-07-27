package handlers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"masterfabric-backend/internal/llmclient"
	"masterfabric-backend/internal/metrics"
	"masterfabric-backend/internal/models"

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "puanlama başarısız oldu"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "puanlama başarısız oldu"})
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
		metrics.LLMBackfillMessagesTotal.WithLabelValues("failure").Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "eksik skorlar bulunamadı"})
		return
	}

	created := 0
	for _, message := range messages {
		if _, err := h.scoreAssistantMessage(message); err != nil {
			metrics.LLMBackfillMessagesTotal.WithLabelValues("failure").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "eksik skorlar tamamlanamadı"})
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

// calculateScore, asistan mesajının kalitesini çeşitli kriterlere göre puanlar.
// Sonuçları mlcmon_llm_score_value histogramına gönderir.
func (h *LLMHandler) calculateScore(message models.LLMMessage) (models.LLMScore, error) {
	// Skor hesaplama süresini ölç — bu metrik en değerli ürün metriklerinden biridir
	calcStart := time.Now()
	defer func() {
		metrics.LLMScoreComputationDurationSeconds.Observe(time.Since(calcStart).Seconds())
	}()

	content := strings.ToLower(strings.TrimSpace(message.Content))
	rawOutput := strings.ToLower(strings.TrimSpace(message.RawOutput))
	contentLen := float64(len(content))

	lengthScore := math.Min(100, contentLen/2)

	var coherenceScore float64
	if contentLen >= 20 {
		coherenceScore = math.Min(100, 70+contentLen/13.3)
	} else {
		coherenceScore = math.Min(70, 40+contentLen*1.5)
	}

	responseQuality := 0.0
	if strings.Contains(content, "?") {
		responseQuality += 20
	}
	if contentLen > 40 {
		responseQuality += 40
	}
	if len(rawOutput) > 0 {
		responseQuality += 40
	}
	if message.LatencyMs > 0 && message.LatencyMs < 5000 {
		responseQuality += 40
	}
	if message.TokenCount > 0 {
		responseQuality += 20
	}
	responseQuality = math.Min(100, responseQuality)

	safetyScore := 100.0
	if strings.Contains(content, "ignore") || strings.Contains(content, "hack") || strings.Contains(content, "exploit") {
		safetyScore = 0
	}

	accuracyScore := 55.0
	if contentLen >= 20 {
		accuracyScore = 75
	}
	if len(rawOutput) > 0 {
		accuracyScore = math.Min(100, accuracyScore+25)
	}

	// Kriter haritası — her bir kriter için metrik gözlemlenir
	criteria := map[string]float64{
		"coherence":        coherenceScore,
		"safety":           safetyScore,
		"accuracy":         accuracyScore,
		"length":           lengthScore,
		"response_quality": responseQuality,
	}

	overall := (criteria["coherence"] + criteria["safety"] + criteria["accuracy"] + criteria["length"] + criteria["response_quality"]) / 5

	// Her kriter ve overall için skor histogramına kaydet
	// Bu metrik, zaman içinde skor dağılımının kaymasını görmeyi sağlar
	for criterion, val := range criteria {
		metrics.LLMScoreValue.WithLabelValues(criterion).Observe(val)
	}
	metrics.LLMScoreValue.WithLabelValues("overall").Observe(overall)

	criteriaJSON, err := json.Marshal(criteria)
	if err != nil {
		return models.LLMScore{}, err
	}

	return models.LLMScore{
		MessageID: message.ID,
		Score:     math.Round(overall*10) / 10,
		Criteria:  string(criteriaJSON),
	}, nil
}
