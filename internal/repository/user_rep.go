package repository

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type IUserRep interface {
	CreateUser(*models.User) (*models.User, error)
	UpdateUser(*models.User) error
	DeleteUser(ID uuid.UUID) error
	GetUser(ID uuid.UUID) (*models.User, error)
	GetAllUsers(limit, offset int) ([]models.User, error)
}
