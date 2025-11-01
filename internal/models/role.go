package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID        uuid.UUID
	Name      string
	IsDeleted bool
	UpdatedAt time.Time
	CreatedAt time.Time
}
