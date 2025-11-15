package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type UserUsecase interface {
	UpdateUser(ID uuid.UUID, req *dto.UpdateUserRequest) error
	GetAllUsers(page, limit int) ([]dto.UserResponse, error)
	GetUser(ID uuid.UUID) (*dto.UserResponse, error)
	DeleteUser(ID uuid.UUID) error
}
