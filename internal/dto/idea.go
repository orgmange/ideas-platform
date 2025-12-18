package dto

import "github.com/google/uuid"

type CreateIdeaRequest struct {
	CoffeeShopID uuid.UUID `form:"coffee_shop_id"`
	CategoryID   uuid.UUID `form:"category_id"`
	Title        string    `form:"title"`
	Description  string    `form:"description"`
}

type UpdateIdeaRequest struct {
	CategoryID  *uuid.UUID `json:"category_id"`
	StatusID    *uuid.UUID `json:"status_id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	ImageURL    *string    `json:"image_url"`
}

type IdeaResponse struct {
	ID           uuid.UUID  `json:"id"`
	CreatorID    *uuid.UUID `json:"creator_id"`
	CoffeeShopID *uuid.UUID `json:"coffee_shop_id"`
	CategoryID   *uuid.UUID `json:"category_id"`
	StatusID     *uuid.UUID `json:"status_id"`
	StatusName   string     `json:"status_name"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ImageURL     *string    `json:"image_url"`
	Likes        int        `json:"likes"`
}

type GetIdeasRequest struct {
	Page  int
	Limit int
	Sort  string
}
