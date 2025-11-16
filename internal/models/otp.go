package models

import "time"

type OTP struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Phone         string    `gorm:"column:phone" json:"phone"`
	CodeHash      string    `gorm:"column:code_hash" json:"-"`
	ExpiresAt     time.Time `gorm:"column:expires_at" json:"expires_at"`
	Verified      bool      `gorm:"column:verified" json:"verified"`
	AttemptsLeft  int       `gorm:"column:attempts_left" json:"attempts_left"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	ResendCount   int       `gorm:"column:resend_count;default:0" json:"resend_count"`
	NextAllowedAt time.Time `gorm:"column:next_allowed_at" json:"next_allowed_at"`
}

func (OTP) TableName() string {
	return "otps"
}
