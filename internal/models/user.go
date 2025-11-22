package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string    `gorm:"not null;size:100"`
	Phone     string    `gorm:"not null;unique;size:15"`
	RoleID    uuid.UUID `gorm:"type:uuid"`
	Role      *Role     `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:SET NULL"`
	IsDeleted bool      `gorm:"default:false"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (User) TableName() string {
	return "users"
}

type BannedUser struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       *uuid.UUID `gorm:"type:uuid"`
	User         User       `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	CoffeeShopID *uuid.UUID `gorm:"type:uuid"`
	CoffeeShop   CoffeeShop `gorm:"foreignKey:CoffeeShopID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}

func (BannedUser) TableName() string {
	return "banned_user"
}
