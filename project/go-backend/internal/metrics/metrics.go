package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP İsteği Sayaçları
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Toplam HTTP istek sayısı",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP İstek Süresi (Histogram)
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP istek yanıt süresi",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// LLM Model Latency (Ollama yanıt süresi)
	LLMLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_generate_duration_seconds",
			Help:    "Ollama model yanıt süresi (saniye)",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 20, 30},
		},
		[]string{"model"},
	)

	// LLM Token Count
	LLMTokenCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_tokens_generated",
			Help:    "Model tarafından üretilen token sayısı",
			Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000},
		},
		[]string{"model"},
	)

	// Hata Sayaçları
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Toplam hata sayısı",
		},
		[]string{"type", "endpoint"},
	)

	// Veritabanı Işlemleri
	DBOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Veritabanı işlem süresi",
			Buckets: []float64{.001, .01, .05, .1, .5, 1},
		},
		[]string{"operation"},
	)

	// Aktif Oturumlar
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llm_active_sessions",
			Help: "Aktif LLM oturum sayısı",
		},
	)

	// Online Kullanıcı Sayısı
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "llm_active_users",
			Help: "Anlık online kullanıcı sayısı",
		},
	)

	// Login Sayacı
	LoginsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_logins_total",
			Help: "Toplam başarılı giriş sayısı",
		},
	)

	// Register Sayacı
	RegistersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_registers_total",
			Help: "Toplam kayıt sayısı",
		},
	)

	// AI Generate Sayacı
	AICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_calls_total",
			Help: "Toplam AI model çağrı sayısı",
		},
		[]string{"model"},
	)
)
