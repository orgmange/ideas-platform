package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CoffeeShopID *uuid.UUID `gorm:"type:uuid"`
	CoffeeShop   CoffeeShop `gorm:"foreignKey:CoffeeShopID;references:ID;constraint:OnDelete:CASCADE"`
	Title        string     `gorm:"not null;size:50"`
	Description  *string
	IsDeleted    bool      `gorm:"default:false"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (Category) TableName() string {
	return "category"
}
