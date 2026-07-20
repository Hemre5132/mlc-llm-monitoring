package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config, uygulamanın çalışması için gereken tüm ortam değişkenlerini tutar.
type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	JWTAccessTTLMin int    // dakika cinsinden access token ömrü
	JWTRefreshTTLHr int    // saat cinsinden refresh token ömrü
	AllowedOrigin   string // CORS için Vercel frontend URL'i
	Env             string // "development" | "production"
}

// Load, .env dosyasını (varsa) okur ve environment değişkenlerinden Config üretir.
// Render/Vercel gibi platformlarda .env dosyası olmaz, doğrudan platform env
// değişkenleri kullanılır; bu yüzden .env bulunamazsa hata fırlatmıyoruz.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("uyarı: .env dosyası bulunamadı, ortam değişkenleri doğrudan sistemden okunacak")
	}

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		JWTAccessTTLMin: 15,
		JWTRefreshTTLHr: 24 * 7, // 7 gün
		AllowedOrigin:   getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		Env:             getEnv("ENV", "development"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL ortam değişkeni zorunludur")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET ortam değişkeni zorunludur")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
