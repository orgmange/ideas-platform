package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type UserUsecase interface {
	UpdateUser(requesterID, ID uuid.UUID, req *dto.UpdateUserRequest) error
	GetAllUsers(role string, page, limit int) ([]dto.UserResponse, error)
	GetUser(role string, requesterID, ID uuid.UUID) (*dto.UserResponse, error)
	DeleteUser(requesterID, ID uuid.UUID) error
}
