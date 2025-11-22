package dto

import (
	"time"

	"github.com/google/uuid"
)

type GiveRewardRequest struct {
	IdeaID       uuid.UUID `json:"idea_id" binding:"required"`
	RewardTypeID uuid.UUID `json:"reward_type_id" binding:"required"`
}

type RewardResponse struct {
	ID           uuid.UUID  `json:"id"`
	ReceiverID   *uuid.UUID `json:"receiver_id,omitempty"`
	CoffeeShopID *uuid.UUID `json:"coffee_shop_id,omitempty"`
	IdeaID       *uuid.UUID `json:"idea_id,omitempty"`
	RewardTypeID *uuid.UUID `json:"reward_type_id,omitempty"`
	IsActivated  bool       `json:"is_activated"`
	GivenAt      *time.Time `json:"given_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
