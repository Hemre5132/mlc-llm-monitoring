package handlers

import (
	"net/http"

	"masterfabric-backend/internal/config"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	Cfg *config.Config
}

func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{Cfg: cfg}
}

// ---------- 9) GET /api/config ----------
// Frontend'in açılışta okuyacağı genel ayarlar (feature flag, ortam bilgisi vb.)

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"env":             h.Cfg.Env,
		"scoring_enabled": true,
		"min_essay_words": 100,
	})
}
