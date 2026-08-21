package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
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

type EssayHandler struct {
	DB     *gorm.DB
	Ollama *llmclient.OllamaClient
}

func NewEssayHandler(db *gorm.DB, ollama *llmclient.OllamaClient) *EssayHandler {
	return &EssayHandler{DB: db, Ollama: ollama}
}

// ---------- POST /api/essays ----------
// Kullanıcının essay'ini kaydeder, Ollama ile puanlar ve skoru döner.
type createEssayRequest struct {
	TopicID string `json:"topic_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (h *EssayHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createEssayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Topic'i doğrula
	topicID, err := uuid.Parse(req.TopicID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "geçersiz topic_id"})
		return
	}

	var topic models.Topic
	if err := h.DB.Where("id = ?", topicID).First(&topic).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "konu bulunamadı"})
		return
	}

	// Essay'i kaydet
	essay := models.Essay{
		UserID:    userID,
		TopicID:   topicID,
		Content:   req.Content,
		WordCount: countWords(req.Content),
	}
	if err := h.DB.Create(&essay).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "essay kaydedilemedi"})
		return
	}

	// Ollama ile puanla
	calcStart := time.Now()
	result, err := services.ScoreEssay(context.Background(), h.Ollama, req.Content, topic.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skorlama şu an yapılamadı, tekrar deneyin"})
		return
	}
	metrics.WritingScoreComputationDurationSeconds.Observe(time.Since(calcStart).Seconds())

	// Essay skorunu kaydet
	errorListJSON, _ := json.Marshal(result.ErrorList)
	strengthsJSON, _ := json.Marshal(result.Strengths)

	score := models.EssayScore{
		EssayID:           essay.ID,
		OverallScore:      result.OverallScore,
		CEFREstimate:      result.CEFREstimate,
		TaskAchievement:   result.TaskAchievement,
		CoherenceCohesion: result.CoherenceCohesion,
		GrammarAccuracy:   result.GrammarAccuracy,
		VocabularyRange:   result.VocabularyRange,
		SpellingMechanics: result.SpellingMechanics,
		SentenceStructure: result.SentenceStructure,
		ErrorList:         string(errorListJSON),
		Strengths:         string(strengthsJSON),
		Reasoning:         result.Reasoning,
	}
	if err := h.DB.Create(&score).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "skor kaydedilemedi"})
		return
	}

	// Metrikleri güncelle
	metrics.WritingEssaysSubmittedTotal.Inc()
	metrics.WritingScoreValue.WithLabelValues("overall").Observe(result.OverallScore)
	metrics.WritingScoreValue.WithLabelValues("task_achievement").Observe(result.TaskAchievement)
	metrics.WritingScoreValue.WithLabelValues("coherence_cohesion").Observe(result.CoherenceCohesion)
	metrics.WritingScoreValue.WithLabelValues("grammar_accuracy").Observe(result.GrammarAccuracy)
	metrics.WritingScoreValue.WithLabelValues("vocabulary_range").Observe(result.VocabularyRange)
	metrics.WritingScoreValue.WithLabelValues("spelling_mechanics").Observe(result.SpellingMechanics)
	metrics.WritingScoreValue.WithLabelValues("sentence_structure").Observe(result.SentenceStructure)
	metrics.WritingEssayWordCount.Observe(float64(essay.WordCount))

	// Essay'i topic + score ile birlikte döndür
	h.DB.Preload("Topic").Preload("Score").First(&essay, essay.ID)

	c.JSON(http.StatusCreated, gin.H{"essay": essay, "score": score})
}

// ---------- GET /api/essays ----------
// Kullanıcının essay'lerini created_at desc sırayla, sayfalama ile döner.
func (h *EssayHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	var essays []models.Essay
	if err := h.DB.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Preload("Topic").
		Preload("Score").
		Find(&essays).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "essay'ler getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"essays": essays})
}

// ---------- GET /api/essays/:id ----------
// Tek essay detayı (topic + score + error_list dahil), sahiplik kontrolü ile.
func (h *EssayHandler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	essayID := c.Param("id")

	var essay models.Essay
	if err := h.DB.Where("id = ? AND user_id = ?", essayID, userID).
		Preload("Topic").
		Preload("Score").
		First(&essay).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "essay bulunamadı veya yetkiniz yok"})
		return
	}

	c.JSON(http.StatusOK, essay)
}

// countWords, metindeki kelime sayısını hesaplar.
func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}
