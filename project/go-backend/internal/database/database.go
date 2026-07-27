package database

import (
	"database/sql"
	"log"
	"time"

	"masterfabric-backend/internal/metrics"
	"masterfabric-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect, Postgres'e bağlanır ve modelleri otomatik migrate eder.
// Render'da Postgres eklentisi oluşturduğunda sana bir "Internal Database URL"
// verir; onu DATABASE_URL olarak kullan.
func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("veritabanına bağlanılamadı: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.LLMSession{},
		&models.LLMMessage{},
		&models.LLMScore{},
	); err != nil {
		log.Fatalf("migrasyon başarısız: %v", err)
	}

	log.Println("veritabanı bağlantısı kuruldu ve migrasyon tamamlandı")
	return db
}

// StartPoolStatsCollector, her 15 saniyede bir veritabanı havuz istatistiklerini
// (açık bağlantı, kullanımdaki, boştaki) ve bağlantı durumunu (Ping) ölçüp
// ilgili Prometheus gauges'larına yazar. main.go'da Connect()'ten sonra
// goroutine olarak çalıştırılmalıdır.
func StartPoolStatsCollector(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("StartPoolStatsCollector: sql.DB alınamadı: %v", err)
		return
	}

	go func() {
		// İlk okumayı hemen yap
		collectPoolStats(sqlDB)

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			collectPoolStats(sqlDB)
		}
	}()

	log.Println("veritabanı havuz istatistik toplayıcı başlatıldı (15s aralık)")
}

func collectPoolStats(sqlDB *sql.DB) {
	stats := sqlDB.Stats()
	metrics.DBPoolOpenConnections.Set(float64(stats.OpenConnections))
	metrics.DBPoolInUse.Set(float64(stats.InUse))
	metrics.DBPoolIdle.Set(float64(stats.Idle))

	if err := sqlDB.Ping(); err != nil {
		metrics.DBUp.Set(0)
	} else {
		metrics.DBUp.Set(1)
	}
}
