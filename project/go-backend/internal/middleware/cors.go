package middleware

import (
	"masterfabric-backend/internal/metrics"

	"github.com/gin-gonic/gin"
)

// CORS, sadece belirlenen origin'e (Vercel frontend URL'in) izin verir.
// Geliştirme sırasında ALLOWED_ORIGIN=http://localhost:3000 kullan,
// production'da Vercel domain'ini (örn. https://senin-projen.vercel.app) ayarla.
func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Eğer Origin header'ı varsa ve izin verilen origin ile eşleşmiyorsa
		// mlcmon_cors_rejected_total sayacını artır (güvenlik sinyali).
		// Origin değerini doğrudan label olarak kullanmıyoruz — kardinalite
		// ve enjeksiyon riskine karşı "mismatched" veya "matched" olarak
		// sınıflandırıyoruz.
		if origin != "" {
			if origin == allowedOrigin {
				metrics.CORSRejectedTotal.WithLabelValues("matched").Inc()
			} else {
				metrics.CORSRejectedTotal.WithLabelValues("mismatched").Inc()
			}
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
