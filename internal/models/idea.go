package models

import (
	"time"

	"github.com/google/uuid"
)

type Idea struct {
	ID           uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatorID    *uuid.UUID `gorm:"type:uuid"`
	Creator      User       `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:SET NULL"`
	CoffeeShopID *uuid.UUID `gorm:"type:uuid"`
	CoffeeShop   CoffeeShop `gorm:"foreignKey:CoffeeShopID;references:ID;constraint:OnDelete:SET NULL"`
	CategoryID   *uuid.UUID `gorm:"type:uuid"`
	Category     Category   `gorm:"foreignKey:CategoryID;references:ID;constraint:OnDelete:SET NULL"`
	StatusID     *uuid.UUID `gorm:"type:uuid"`
	Status       IdeaStatus `gorm:"foreignKey:StatusID;references:ID;constraint:OnDelete:SET NULL"`
	Title        string     `gorm:"not null;size:150"`
	Description  string     `gorm:"not null"`
	ImageURL     *string    `gorm:"size:255"`
	IsDeleted    bool       `gorm:"default:false"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
}

func (Idea) TableName() string {
	return "idea"
}

type IdeaLike struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    *uuid.UUID `gorm:"type:uuid"`
	User      User       `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	IdeaID    *uuid.UUID `gorm:"type:uuid"`
	Idea      Idea       `gorm:"foreignKey:IdeaID;references:ID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (IdeaLike) TableName() string {
	return "idea_like"
}

type IdeaComment struct {
	ID        uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CreatorID *uuid.UUID `gorm:"type:uuid"`
	Creator   User       `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:CASCADE"`
	IdeaID    *uuid.UUID `gorm:"type:uuid"`
	Idea      Idea       `gorm:"foreignKey:IdeaID;references:ID;constraint:OnDelete:CASCADE"`
	Text      string     `gorm:"not null"`
	IsDeleted bool       `gorm:"default:false"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (IdeaComment) TableName() string {
	return "idea_comment"
}

type IdeaStatus struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title     string    `gorm:"not null;unique;size:50"`
	IsDeleted bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (IdeaStatus) TableName() string {
	return "status"
}
