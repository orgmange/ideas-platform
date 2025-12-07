package usecase

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type UserUsecase interface {
	UpdateUser(ctx context.Context, actorID, ID uuid.UUID, req *dto.UpdateUserRequest) error
	GetAllUsers(ctx context.Context, actorID uuid.UUID, page, limit int) ([]dto.UserResponse, error)
	GetUser(ctx context.Context, actorID, ID uuid.UUID) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, actorID, ID uuid.UUID) error
}
