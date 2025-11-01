package models

import (
	"time"

	"github.com/google/uuid"
)

type Idea struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	CoffeeShopID *uuid.UUID
	CategoryID   *uuid.UUID
	StatusID     *uuid.UUID
	Title        string
	Description  string
	ImageURL     *string
	IsDeleted    bool
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

type IdeaLike struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	IdeaID    *uuid.UUID
	CreatedAt time.Time
}

type IdeaComment struct {
	ID        uuid.UUID
	CreatorID *uuid.UUID
	IdeaID    *uuid.UUID
	Text      string
	IsDeleted bool
	UpdatedAt time.Time
	CreatedAt time.Time
}

type IdeaStatus struct {
	ID        uuid.UUID
	Title     string
	IsDeleted bool
	CreatedAt time.Time
}
