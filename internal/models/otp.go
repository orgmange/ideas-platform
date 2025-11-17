package models

import "time"

type OTP struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Phone         string    `gorm:"column:phone;not null;size:20" json:"phone"`
	CodeHash      string    `gorm:"column:code_hash;not null;size:255" json:"-"`
	ExpiresAt     time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	Verified      bool      `gorm:"column:verified;default:false" json:"verified"`
	AttemptsLeft  int       `gorm:"column:attempts_left;default:3" json:"attempts_left"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ResendCount   int       `gorm:"column:resend_count;default:0" json:"resend_count"`
	NextAllowedAt time.Time `gorm:"column:next_allowed_at" json:"next_allowed_at"`
}

func (OTP) TableName() string {
	return "otps"
}
