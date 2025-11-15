package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type UserUsecaseImpl struct {
	rep repository.UserRep
}

func NewUserUsecase(rep repository.UserRep) UserUsecase {
	return &UserUsecaseImpl{rep: rep}
}

// DeleteUser implements IUserUsecase.
func (u *UserUsecaseImpl) DeleteUser(ID uuid.UUID) error {
	return u.rep.DeleteUser(ID)
}

// GetAllUsers implements IUserUsecase.
func (u *UserUsecaseImpl) GetAllUsers(page int, limit int) ([]dto.UserResponse, error) {
	if limit <= 0 || limit > 25 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}
	users, err := u.rep.GetAllUsers(limit, limit*page)
	if err != nil {
		return nil, err
	}
	return toResponses(users), nil
}

// GetUser implements IUserUsecase.
func (u *UserUsecaseImpl) GetUser(ID uuid.UUID) (*dto.UserResponse, error) {
	user, err := u.rep.GetUser(ID)
	if err != nil {
		return nil, err
	}

	return toResponse(user), nil
}

// UpdateUser implements IUserUsecase.
func (u *UserUsecaseImpl) UpdateUser(ID uuid.UUID, req *dto.UpdateUserRequest) error {
	user, err := u.rep.GetUser(ID)
	if err != nil {
		return err
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	return u.rep.UpdateUser(user)
}

func toUser(req *dto.CreateUserRequest) *models.User {
	return &models.User{
		Name:  req.Name,
		Phone: req.Phone,
	}
}

func toResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Phone: user.Phone,
	}
}

func toResponses(users []models.User) []dto.UserResponse {
	res := make([]dto.UserResponse, len(users))
	for i := range users {
		res[i] = *toResponse(&users[i])
	}

	return res
}
