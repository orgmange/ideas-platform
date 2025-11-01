package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID           uuid.UUID
	CoffeeShopID *uuid.UUID
	Title        string
	Description  *string
	IsDeleted    bool
	UpdatedAt    time.Time
	CreatedAt    time.Time
}
