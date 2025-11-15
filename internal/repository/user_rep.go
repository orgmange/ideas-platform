package repository

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type UserRep interface {
	UpdateUser(*models.User) error
	DeleteUser(ID uuid.UUID) error
	GetUser(ID uuid.UUID) (*models.User, error)
	GetAllUsers(limit, offset int) ([]models.User, error)
	IsUserExist(ID uuid.UUID) (bool, error)
}
