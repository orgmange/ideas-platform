package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type UserUsecase struct {
	rep repository.IUserRep
}

// DeleteUser implements IUserUsecase.
func (u *UserUsecase) DeleteUser(ID uuid.UUID) error {
	return u.rep.DeleteUser(ID)
}

// GetUser implements IUserUsecase.
func (u *UserUsecase) GetUser(ID uuid.UUID) (*dto.UserResponse, error) {
	user, err := u.rep.GetUser(ID)
	if err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

// UpdateUser implements IUserUsecase.
func (u *UserUsecase) UpdateUser(req *dto.UpdateUserRequest) error {
	user, err := u.rep.GetUser(req.ID)
	if err != nil {
		return err
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	return u.rep.UpdateUser(user)
}

func NewUserUsecase(rep repository.IUserRep) IUserUsecase {
	return &UserUsecase{rep: rep}
}

func (u *UserUsecase) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	user := toUserModel(req)
	createdUser, err := u.rep.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return toUserResponse(createdUser), nil
}

func toUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:   user.ID,
		Name: user.Name,
	}
}

func toUserModel(req *dto.CreateUserRequest) *models.User {
	return &models.User{
		Name:  req.Name,
		Phone: req.Phone,
	}
}
