package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRefreshToken struct {
	UserID       uuid.UUID `gorm:"not null;foreignKey:UserID;constraint:OnDelete:CASCADE"`
	RefreshToken string    `gorm:"primaryKey;type:text"`
	ExpiresAt    time.Time `gorm:"not null"`
}

func (UserRefreshToken) TableName() string {
	return "user_refresh_tokens"
}
