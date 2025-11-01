package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type IUserUsecase interface {
	CreateUser(*dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(*dto.UpdateUserRequest) error
	GetUser(ID uuid.UUID) (*dto.UserResponse, error)
	DeleteUser(ID uuid.UUID) error
}
