package models

import (
	"time"

	"github.com/google/uuid"
)

type CoffeeShop struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatorID      uuid.UUID `gorm:"type:uuid;not null"`
	Creator        User      `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:CASCADE"`
	Name           string    `gorm:"not null;size:100"`
	Address        string    `gorm:"not null;size:255"`
	Contacts       *string   `gorm:"size:100"`
	WelcomeMessage *string
	Rules          *string
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

type WorkerCoffeeShop struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WorkerID     *uuid.UUID `gorm:"type:uuid"`
	Worker       User       `gorm:"foreignKey:WorkerID;references:ID;constraint:OnDelete:CASCADE"`
	CoffeeShopID *uuid.UUID `gorm:"type:uuid"`
	CoffeeShop   CoffeeShop `gorm:"foreignKey:CoffeeShopID;references:ID;constraint:OnDelete:CASCADE"`
	IsDeleted    bool       `gorm:"default:false"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}

func (CoffeeShop) TableName() string {
	return "coffee_shop"
}

func (WorkerCoffeeShop) TableName() string {
	return "worker_coffee_shop"
}
