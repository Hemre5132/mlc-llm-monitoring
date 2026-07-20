package database

import (
	"log"

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
