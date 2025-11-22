package models

import (
	"time"

	"github.com/google/uuid"
)

type Reward struct {
	ID           uuid.UUID   `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ReceiverID   *uuid.UUID  `gorm:"type:uuid"`
	Receiver     *User       `gorm:"foreignKey:ReceiverID"`
	CoffeeShopID *uuid.UUID  `gorm:"type:uuid"`
	CoffeeShop   *CoffeeShop `gorm:"foreignKey:CoffeeShopID"`
	IdeaID       *uuid.UUID  `gorm:"type:uuid"`
	Idea         *Idea       `gorm:"foreignKey:IdeaID"`
	RewardTypeID *uuid.UUID  `gorm:"type:uuid"`
	RewardType   *RewardType `gorm:"foreignKey:RewardTypeID"`
	IsActivated  bool        `gorm:"default:false"`
	GivenAt      *time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (Reward) TableName() string {
	return "reward"
}

type RewardType struct {
	ID           uuid.UUID   `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CoffeeShopID *uuid.UUID  `gorm:"type:uuid"`
	CoffeeShop   *CoffeeShop `gorm:"foreignKey:CoffeeShopID"`
	Description  string      `gorm:"not null"`
	CreatedAt    time.Time   `gorm:"autoCreateTime"`
}

func (RewardType) TableName() string {
	return "reward_type"
}
