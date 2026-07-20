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
	var sessionCount, messageCount int64
	var avgScore float64

	h.DB.Table("llm_sessions").Count(&sessionCount)
	h.DB.Table("llm_messages").Count(&messageCount)
	h.DB.Table("llm_scores").Select("COALESCE(AVG(score), 0)").Scan(&avgScore)

	c.JSON(http.StatusOK, gin.H{
		"total_sessions": sessionCount,
		"total_messages": messageCount,
		"average_score":  avgScore,
	})
}

// ---------- 22) GET /api/stats/me ----------
// Giriş yapmış kullanıcıya özel istatistikler.

func (h *CommonHandler) MyStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var sessionCount int64
	var avgScore float64

	h.DB.Table("llm_sessions").Where("user_id = ?", userID).Count(&sessionCount)
	h.DB.Table("llm_scores").
		Joins("JOIN llm_messages ON llm_messages.id = llm_scores.message_id").
		Joins("JOIN llm_sessions ON llm_sessions.id = llm_messages.session_id").
		Where("llm_sessions.user_id = ?", userID).
		Select("COALESCE(AVG(llm_scores.score), 0)").
		Scan(&avgScore)

	c.JSON(http.StatusOK, gin.H{
		"session_count": sessionCount,
		"average_score": avgScore,
	})
}
