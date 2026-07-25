package main

import (
	"log"

	"masterfabric-backend/internal/config"
	"masterfabric-backend/internal/database"
	"masterfabric-backend/internal/handlers"
	"masterfabric-backend/internal/llmclient"
	"masterfabric-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg.DatabaseURL)

	ollamaClient := llmclient.NewOllamaClient(cfg.OllamaURL, cfg.OllamaModel)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(middleware.CORS(cfg.AllowedOrigin))
	router.Use(middleware.PrometheusMetrics())

	// Prometheus metrik endpoint'i — Gin dışında, doğrudan router'a bağlı
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	authHandler := handlers.NewAuthHandler(db, cfg)
	configHandler := handlers.NewConfigHandler(cfg)
	llmHandler := handlers.NewLLMHandler(db, ollamaClient)
	commonHandler := handlers.NewCommonHandler(db)

	auth := middleware.RequireAuth(cfg.JWTSecret)

	api := router.Group("/api")
	{
		// ---- Auth [8] ----
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/forgot-password", authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authHandler.ResetPassword)
			authGroup.POST("/verify-email", authHandler.VerifyEmail)
			authGroup.GET("/me", auth, authHandler.Me)
		}

		// ---- Config [2] ----
		configGroup := api.Group("/config")
		{
			configGroup.GET("", configHandler.GetConfig)
			configGroup.GET("/models", configHandler.GetModels)
		}

		// ---- Web MLC-LLM [8] ----
		llmGroup := api.Group("/llm", auth)
		{
			llmGroup.POST("/sessions", llmHandler.CreateSession)
			llmGroup.GET("/sessions", llmHandler.ListSessions)
			llmGroup.GET("/sessions/:id", llmHandler.GetSession)
			llmGroup.DELETE("/sessions/:id", llmHandler.DeleteSession)
			llmGroup.POST("/sessions/:id/messages", llmHandler.CreateMessage)
			llmGroup.GET("/sessions/:id/messages", llmHandler.ListMessages)
			llmGroup.POST("/sessions/:id/score", llmHandler.CreateScore)
			llmGroup.GET("/sessions/:id/score", llmHandler.GetScores)
			llmGroup.POST("/scores/backfill", llmHandler.BackfillScores)
			llmGroup.POST("/sessions/:id/generate", llmHandler.GenerateChat)
		}

		// ---- Common Services [4] ----
		api.GET("/health", commonHandler.Health)
		api.GET("/version", commonHandler.Version)
		api.GET("/stats/dashboard", auth, commonHandler.DashboardStats)
		api.GET("/stats/me", auth, commonHandler.MyStats)
	}

	log.Printf("sunucu :%s portunda başlatılıyor (env=%s)", cfg.Port, cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("sunucu başlatılamadı: %v", err)
	}
}
