package repository

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type UserRep interface {
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, ID uuid.UUID) error
	GetUser(ctx context.Context, ID uuid.UUID) (*models.User, error)
	GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, error)
	IsUserExist(ctx context.Context, ID uuid.UUID) (bool, error)
}
