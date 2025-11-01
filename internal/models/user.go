package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Phone     string
	RoleID    *uuid.UUID
	IsDeleted bool
	UpdatedAt time.Time
	CreatedAt time.Time
}

type BannedUser struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	CoffeeShopID *uuid.UUID
	CreatedAt    time.Time
}
