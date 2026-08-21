// internal/metrics/metrics.go — tüm Prometheus metrik tanımları
// Bu dosyadaki her metrik adı "mlcmon_" önekiyle başlar.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ============================================================
// A. HTTP katmanı — middleware/metrics.go tarafından kullanılır
// ============================================================

// HTTPRequestsTotal — her HTTP isteğinde artar; method, path, status etiketleriyle.
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_http_requests_total",
		Help: "Toplam HTTP istek sayısı (method, path, status)",
	},
	[]string{"method", "path", "status"},
)

// HTTPRequestDurationSeconds — istek süresi histogramı; method, path etiketleriyle.
var HTTPRequestDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mlcmon_http_request_duration_seconds",
		Help:    "HTTP istek yanıt süresi (saniye)",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	},
	[]string{"method", "path"},
)

// HTTPRequestsInFlight — anlık işlenmekte olan istek sayısı (gauge).
var HTTPRequestsInFlight = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "mlcmon_http_requests_in_flight",
		Help: "Anlık işlenmekte olan HTTP istek sayısı",
	},
)

// HTTPResponseSizeBytes — yanıt boyutu histogramı; method, path etiketleriyle.
var HTTPResponseSizeBytes = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mlcmon_http_response_size_bytes",
		Help:    "HTTP yanıt boyutu (bayt)",
		Buckets: prometheus.ExponentialBuckets(64, 2, 12), // 64B – 256KB
	},
	[]string{"method", "path"},
)

// ============================================================
// B. Auth domaini — handlers/auth.go ve middleware/auth.go
// ============================================================

// AuthLoginAttemptsTotal — giriş denemeleri; result: success|invalid_credentials|user_not_found
var AuthLoginAttemptsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_auth_login_attempts_total",
		Help: "Giriş denemesi sayısı (result)",
	},
	[]string{"result"},
)

// AuthTokenRefreshTotal — token yenileme; result: success|invalid_token
var AuthTokenRefreshTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_auth_token_refresh_total",
		Help: "Token yenileme sayısı (result)",
	},
	[]string{"result"},
)

// AuthRegistrationsTotal — kayıt denemeleri; result: success|email_exists
var AuthRegistrationsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_auth_registrations_total",
		Help: "Kayıt denemesi sayısı (result)",
	},
	[]string{"result"},
)

// AuthPasswordResetRequestedTotal — şifre sıfırlama talep sayısı.
var AuthPasswordResetRequestedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "mlcmon_auth_password_reset_requested_total",
		Help: "Şifre sıfırlama talep sayısı",
	},
)

// AuthPasswordResetCompletedTotal — şifre sıfırlama tamamlama; result: success|invalid_or_expired_token
var AuthPasswordResetCompletedTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_auth_password_reset_completed_total",
		Help: "Şifre sıfırlama tamamlama sayısı (result)",
	},
	[]string{"result"},
)

// AuthEmailVerificationTotal — email doğrulama; result: success|invalid_token
var AuthEmailVerificationTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_auth_email_verification_total",
		Help: "Email doğrulama sayısı (result)",
	},
	[]string{"result"},
)

// AuthUnauthorizedTotal — yetkisiz erişim; reason: missing_header|invalid_token|wrong_token_type
var AuthUnauthorizedTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_auth_unauthorized_total",
		Help: "Yetkisiz erişim sayısı (reason)",
	},
	[]string{"reason"},
)

// ============================================================
// C. Veritabanı katmanı — database.go ve handler'lardaki GORM çağrıları
// ============================================================

// DBQueryDurationSeconds — sorgu süresi histogramı; handler etiketiyle.
var DBQueryDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mlcmon_db_query_duration_seconds",
		Help:    "Veritabanı sorgu süresi (saniye, handler bazında)",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
	},
	[]string{"handler"},
)

// DBPoolOpenConnections — açık bağlantı sayısı (gauge).
var DBPoolOpenConnections = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "mlcmon_db_pool_open_connections",
		Help: "Veritabanı havuzundaki açık bağlantı sayısı",
	},
)

// DBPoolInUse — kullanımdaki bağlantı sayısı (gauge).
var DBPoolInUse = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "mlcmon_db_pool_in_use",
		Help: "Veritabanı havuzundaki kullanımdaki bağlantı sayısı",
	},
)

// DBPoolIdle — boşta bekleyen bağlantı sayısı (gauge).
var DBPoolIdle = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "mlcmon_db_pool_idle",
		Help: "Veritabanı havuzundaki boş bağlantı sayısı",
	},
)

// DBUp — veritabanı erişilebilir mi (1/0) (gauge).
var DBUp = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "mlcmon_db_up",
		Help: "Veritabanı erişilebilirlik durumu (1=erişilebilir, 0=değil)",
	},
)

// ============================================================
// D. Core product (Daily English Writing Coach) — handlers/writing.go
// ============================================================

// WritingEssaysSubmittedTotal — gönderilen essay sayısı.
var WritingEssaysSubmittedTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "mlcmon_writing_essays_submitted_total",
		Help: "Gönderilen essay sayısı",
	},
)

// WritingScoreValue — essay skor değeri histogramı; criterion:
// overall|task_achievement|coherence_cohesion|grammar_accuracy|vocabulary_range|spelling_mechanics|sentence_structure
var WritingScoreValue = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "mlcmon_writing_score_value",
		Help:    "Essay skor değeri dağılımı (criterion bazında)",
		Buckets: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	},
	[]string{"criterion"},
)

// WritingScoreComputationDurationSeconds — essay skor hesaplama süresi histogramı.
var WritingScoreComputationDurationSeconds = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "mlcmon_writing_score_computation_duration_seconds",
		Help:    "Essay skor hesaplama süresi (saniye)",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5},
	},
)

// WritingEssayWordCount — essay kelime sayısı dağılımı (histogram).
var WritingEssayWordCount = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "mlcmon_writing_essay_word_count",
		Help:    "Essay kelime sayısı dağılımı",
		Buckets: []float64{50, 100, 150, 200, 300, 400, 600, 1000},
	},
)

// ============================================================
// E. Güvenlik sinyalleri — middleware/cors.go
// ============================================================

// CORSRejectedTotal — reddedilen CORS istekleri; origin: mismatched|matched
var CORSRejectedTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mlcmon_cors_rejected_total",
		Help: "CORS tarafından reddedilen istek sayısı (origin durumu)",
	},
	[]string{"origin"},
)
