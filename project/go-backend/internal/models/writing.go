package models

import (
	"time"

	"github.com/google/uuid"
)

// Topic, kullanıcıya sunulan bir essay konusunu temsil eder.
// "daily" konular sistem tarafından üretilir (tarihe göre deterministik seçilir),
// "custom" konular kullanıcının kendi girdiği konulardır.
type Topic struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Text            string     `gorm:"type:text;not null" json:"text"`
	Category        string     `gorm:"index" json:"category"`                 // opinion | narrative | argumentative | descriptive | compare_contrast
	Difficulty      string     `json:"difficulty"`                            // beginner | intermediate | advanced
	Source          string     `gorm:"not null;default:system" json:"source"` // system | user
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index" json:"created_by_user_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Essay, kullanıcının bir konu hakkında yazdığı metni tutar.
type Essay struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	TopicID   uuid.UUID `gorm:"type:uuid;index;not null" json:"topic_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	WordCount int       `json:"word_count"`
	CreatedAt time.Time `gorm:"index" json:"created_at"` // streak hesaplaması için index'li

	Topic Topic       `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	Score *EssayScore `gorm:"foreignKey:EssayID" json:"score,omitempty"`
}

// EssayScore, bir essay için LLM'in ürettiği puan + hata analizini tutar.
type EssayScore struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EssayID           uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"essay_id"`
	OverallScore      float64   `gorm:"not null" json:"overall_score"` // 0-100
	CEFREstimate      string    `json:"cefr_estimate"`                 // A1|A2|B1|B2|C1|C2 (tahmini)
	TaskAchievement   float64   `json:"task_achievement"`
	CoherenceCohesion float64   `json:"coherence_cohesion"`
	GrammarAccuracy   float64   `json:"grammar_accuracy"`
	VocabularyRange   float64   `json:"vocabulary_range"`
	SpellingMechanics float64   `json:"spelling_mechanics"`
	SentenceStructure float64   `json:"sentence_structure"`
	ErrorList         string    `gorm:"type:jsonb" json:"error_list"` // []EssayError JSON
	Strengths         string    `gorm:"type:jsonb" json:"strengths"`  // []string JSON
	Reasoning         string    `gorm:"type:text" json:"reasoning"`
	CreatedAt         time.Time `json:"created_at"`
}

// EssayError, tek bir hatayı temsil eder (JSON olarak EssayScore.ErrorList içinde saklanır).
type EssayError struct {
	Category    string `json:"category"`    // grammar | vocabulary | spelling | punctuation | sentence_structure
	Original    string `json:"original"`    // hatalı orijinal ifade
	Correction  string `json:"correction"`  // önerilen düzeltme
	Explanation string `json:"explanation"` // neden hatalı, kısa açıklama
}
