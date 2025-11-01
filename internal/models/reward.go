package models

import (
	"time"

	"github.com/google/uuid"
)

type Reward struct {
	ID           uuid.UUID
	ReceiverID   *uuid.UUID
	CoffeeShopID *uuid.UUID
	IdeaID       *uuid.UUID
	RewardTypeID *uuid.UUID
	IsActivated  bool
	GivenAt      *time.Time
	CreatedAt    time.Time
}

type RewardType struct {
	ID           uuid.UUID
	CoffeeShopID *uuid.UUID
	Description  string
	CreatedAt    time.Time
}
