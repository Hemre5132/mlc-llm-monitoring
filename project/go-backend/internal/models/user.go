package models

import (
	"time"

	"github.com/google/uuid"
)

// User, sisteme kayıtlı bir kullanıcıyı temsil eder.
// PasswordHash asla JSON çıktısına dahil edilmez (json:"-").
type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email         string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash  string     `gorm:"not null" json:"-"`
	Name          string     `json:"name"`
	EmailVerified bool       `gorm:"default:false" json:"email_verified"`
	VerifyToken   string     `gorm:"index" json:"-"`
	ResetToken    string     `gorm:"index" json:"-"`
	ResetTokenExp *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
