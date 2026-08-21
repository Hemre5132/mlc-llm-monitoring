package handlers

import (
	"net/http"

	"masterfabric-backend/internal/database"
	"masterfabric-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TopicHandler struct {
	DB *gorm.DB
}

func NewTopicHandler(db *gorm.DB) *TopicHandler {
	return &TopicHandler{DB: db}
}

// ---------- GET /api/topics/daily ----------
// Günün konusunu deterministik olarak döner (tüm kullanıcılar aynı konuyu görür).
func (h *TopicHandler) GetDaily(c *gin.Context) {
	topic, err := database.GetDailyTopic(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "günün konusu alınamadı"})
		return
	}

	c.JSON(http.StatusOK, topic)
}

// ---------- GET /api/topics/random ----------
// Random bir sistem konusu döner; opsiyonel "exclude" query param'ı ile
// son gösterilen konu hariç tutulabilir.
func (h *TopicHandler) GetRandom(c *gin.Context) {
	excludeParam := c.Query("exclude")
	var excludeID *uuid.UUID
	if excludeParam != "" {
		parsed, err := uuid.Parse(excludeParam)
		if err == nil {
			excludeID = &parsed
		}
	}

	topic, err := database.GetRandomTopic(h.DB, excludeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rastgele konu alınamadı"})
		return
	}

	c.JSON(http.StatusOK, topic)
}

// ---------- POST /api/topics/custom ----------
// Kullanıcının kendi yazdığı konuyu kaydeder (source="user").
type createCustomTopicRequest struct {
	Text string `json:"text" binding:"required"`
}

func (h *TopicHandler) CreateCustom(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createCustomTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	topic := models.Topic{
		Text:            req.Text,
		Category:        "custom",
		Difficulty:      "intermediate",
		Source:          "user",
		CreatedByUserID: &userID,
	}
	if err := h.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "konu oluşturulamadı"})
		return
	}

	c.JSON(http.StatusCreated, topic)
}
