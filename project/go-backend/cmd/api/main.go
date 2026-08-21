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

	// Sistem essay konularını (60 adet) ilk açılışta ekle
	database.SeedTopics(db)

	// Veritabanı havuz istatistik toplayıcısını başlat (15sn aralıklarla
	// mlcmon_db_pool_open_connections, mlcmon_db_pool_in_use,
	// mlcmon_db_pool_idle ve mlcmon_db_up gauge'larını günceller).
	database.StartPoolStatsCollector(db)

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
	topicHandler := handlers.NewTopicHandler(db)
	essayHandler := handlers.NewEssayHandler(db, ollamaClient)
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

		// ---- Config [1] ----
		configGroup := api.Group("/config")
		{
			configGroup.GET("", configHandler.GetConfig)
		}

		// ---- Topics [3] ----
		topicsGroup := api.Group("/topics")
		{
			topicsGroup.GET("/daily", topicHandler.GetDaily)
			topicsGroup.GET("/random", topicHandler.GetRandom)
			topicsGroup.POST("/custom", auth, topicHandler.CreateCustom)
		}

		// ---- Essays [3] ----
		essaysGroup := api.Group("/essays", auth)
		{
			essaysGroup.POST("", essayHandler.Create)
			essaysGroup.GET("", essayHandler.List)
			essaysGroup.GET("/:id", essayHandler.Get)
		}

		// ---- Common Services [5] ----
		api.GET("/health", commonHandler.Health)
		api.GET("/version", commonHandler.Version)
		api.GET("/stats/dashboard", auth, commonHandler.DashboardStats)
		api.GET("/stats/me", auth, commonHandler.MyStats)
		api.GET("/stats/streak", auth, commonHandler.Streak)
	}

	log.Printf("sunucu :%s portunda başlatılıyor (env=%s)", cfg.Port, cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("sunucu başlatılamadı: %v", err)
	}
}
