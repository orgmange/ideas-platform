package models

import (
	"time"

	"github.com/google/uuid"
)

type CoffeeShop struct {
	ID             uuid.UUID
	Name           string
	Address        string
	Contacts       *string
	WelcomeMessage *string
	Rules          *string
	UpdatedAt      time.Time
	CreatedAt      time.Time
}

type WorkerCoffeeShop struct {
	ID           uuid.UUID
	WorkerID     *uuid.UUID
	CoffeeShopID *uuid.UUID
	IsDeleted    bool
	CreatedAt    time.Time
}
