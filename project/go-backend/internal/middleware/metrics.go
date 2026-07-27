// internal/middleware/metrics.go — HTTP metrik middleware'i
// Her istek için mlcmon_* metriklerini kaydeder.
package middleware

import (
	"strconv"
	"time"

	"masterfabric-backend/internal/metrics"

	"github.com/gin-gonic/gin"
)

// PrometheusMetrics, her HTTP isteğini izleyen bir Gin middleware'idir.
// Global olarak (route gruplarından önce) kaydedilmelidir.
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// İşlenmekte olan istek sayacını artır
		metrics.HTTPRequestsInFlight.Inc()
		start := time.Now()

		// İsteğin tamamlanmasını bekle (yanıt yazıldıktan sonra)
		c.Next()

		// İstek tamamlandı, metrikleri güncelle
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Path: Gin'in FullPath()'ini kullan — böylece dinamik segmentler
		// (örn. :id) etiket patlamasına yol açmaz.
		path := c.FullPath()
		if path == "" {
			// Eşleşmeyen rotalar (404) için "unmatched" kullan
			path = "unmatched"
		}

		method := c.Request.Method

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDurationSeconds.WithLabelValues(method, path).Observe(duration)
		metrics.HTTPRequestsInFlight.Dec()
		metrics.HTTPResponseSizeBytes.WithLabelValues(method, path).Observe(float64(c.Writer.Size()))
	}
}
