package middleware

import (
	"net/http"
	"strings"

	"masterfabric-backend/internal/metrics"
	"masterfabric-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// RequireAuth, "Authorization: Bearer <token>" header'ını doğrulayan bir
// middleware'dir. Geçerliyse user_id'yi context'e ekler; değilse 401 döner.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			metrics.AuthUnauthorizedTotal.WithLabelValues("missing_header").Inc()
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "yetkilendirme başlığı eksik"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(tokenStr, jwtSecret)
		if err != nil {
			metrics.AuthUnauthorizedTotal.WithLabelValues("invalid_token").Inc()
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "geçersiz token"})
			return
		}
		if claims.Type != utils.AccessToken {
			metrics.AuthUnauthorizedTotal.WithLabelValues("wrong_token_type").Inc()
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "access token bekleniyor"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
