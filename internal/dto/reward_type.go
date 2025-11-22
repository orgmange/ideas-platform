package dto

import "github.com/google/uuid"

type RewardTypeResponse struct {
	ID           uuid.UUID
	CoffeeShopID uuid.UUID
	Description  string
}

type CreateRewardTypeRequest struct {
	CoffeeShopID uuid.UUID
	Description  string
}

type UpdateRewardTypeRequest struct {
	Description *string
}
