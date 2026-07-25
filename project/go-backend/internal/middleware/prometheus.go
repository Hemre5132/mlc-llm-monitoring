package middleware

import (
	"strconv"
	"time"

	"masterfabric-backend/internal/metrics"

	"github.com/gin-gonic/gin"
)

// PrometheusMetrics, her HTTP isteğini sayar ve süresini ölçer.
// Gin router'ına eklenen middleware, istek bittiğinde metrikleri günceller.
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // örn: /api/auth/login, /api/llm/sessions/:id/generate

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		metrics.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)

		if c.Writer.Status() >= 400 {
			metrics.ErrorsTotal.WithLabelValues("http", path).Inc()
		}
	}
}
