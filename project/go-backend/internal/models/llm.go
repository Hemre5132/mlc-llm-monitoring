package models

import (
	"time"

	"github.com/google/uuid"
)

// LLMSession, tarayıcıda WebLLM (Gemma) ile başlatılan bir konuşma oturumunu
// temsil eder. Her session bir kullanıcıya ve bir model adına bağlıdır.
type LLMSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	ModelName string    `gorm:"not null" json:"model_name"` // örn: "gemma-2b-it-q4f16_1-MLC"
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Messages []LLMMessage `gorm:"foreignKey:SessionID" json:"messages,omitempty"`
}

// LLMMessage, bir oturum içindeki tek bir prompt/response çiftini (ham çıktı
// dahil) tutar. RawOutput, WebLLM'den gelen işlenmemiş ham metni saklar.
type LLMMessage struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SessionID  uuid.UUID `gorm:"type:uuid;index;not null" json:"session_id"`
	Role       string    `gorm:"not null" json:"role"` // "user" | "assistant"
	Content    string    `gorm:"type:text;not null" json:"content"`
	RawOutput  string    `gorm:"type:text" json:"raw_output,omitempty"`
	LatencyMs  int       `json:"latency_ms,omitempty"`
	TokenCount int       `json:"token_count,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// LLMScore, bir mesaj için hesaplanan "Deci.Scoring" (karar puanlama)
// sonucunu tutar. Criteria alanı, alt kriter kırılımını JSON olarak saklar
// (örn: {"coherence": 82, "safety": 95, "accuracy": 70}).
type LLMScore struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID        uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"message_id"`
	Score            float64   `gorm:"not null" json:"score"` // 0-100 arası genel skor
	Criteria         string    `gorm:"type:jsonb" json:"criteria"`
	Category         string    `gorm:"type:text" json:"category"`
	CategoryScores   string    `gorm:"type:jsonb" json:"category_scores"`
	RequiresRevision bool      `gorm:"not null;default:false" json:"requires_revision"`
	CreatedAt        time.Time `json:"created_at"`
}
