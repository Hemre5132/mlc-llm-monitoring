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
		"env":                  h.Cfg.Env,
		"scoring_enabled":      true,
		"max_session_messages": 200,
	})
}

// ---------- 10) GET /api/config/models ----------
// WebLLM'in tarayıcıda hangi Gemma varyantlarını yükleyebileceğini backend
// belirler — böylece model listesi frontend'e hardcode edilmez, tek yerden
// yönetilir.

type modelInfo struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	SizeMB      int    `json:"size_mb"`
	Recommended bool   `json:"recommended"`
}

func (h *ConfigHandler) GetModels(c *gin.Context) {
	models := []modelInfo{
		{ID: "gemma-2b-it-q4f16_1-MLC", Label: "Gemma 2B Instruct (q4f16)", SizeMB: 1500, Recommended: true},
		{ID: "gemma-2-2b-it-q4f16_1-MLC", Label: "Gemma 2 2B Instruct (q4f16)", SizeMB: 1600, Recommended: false},
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}
