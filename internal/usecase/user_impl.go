package usecase

import (
	"errors"
	"log/slog"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type UserUsecaseImpl struct {
	rep    repository.UserRep
	logger *slog.Logger
}

func NewUserUsecase(rep repository.UserRep, logger *slog.Logger) UserUsecase {
	return &UserUsecaseImpl{rep: rep, logger: logger}
}

// DeleteUser implements IUserUsecase.
func (u *UserUsecaseImpl) DeleteUser(requesterID, ID uuid.UUID) error {
	logger := u.logger.With("method", "DeleteUser", "requesterID", requesterID.String(), "userID", ID.String())
	logger.Debug("starting delete user")

	if !isOwner(requesterID, ID) {
		logger.Info("access denied")
		return apperrors.NewErrAccessDenied("forbidden")
	}
	err := u.rep.DeleteUser(ID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("user to delete not found")
			return err
		}
		logger.Error("failed to delete user", "error", err.Error())
		return err
	}

	logger.Info("user deleted successfully")
	return nil
}

// GetAllUsers implements IUserUsecase.
func (u *UserUsecaseImpl) GetAllUsers(role string, page int, limit int) ([]dto.UserResponse, error) {
	logger := u.logger.With("method", "GetAllUsers", "page", page, "limit", limit)
	logger.Debug("starting get all users")

	if role != "admin" {
		logger.Info("access denied")
		return nil, apperrors.NewErrAccessDenied("forbidden")
	}

	if limit <= 0 || limit > 25 {
		limit = 25
	}
	if page < 0 {
		page = 0
	}
	users, err := u.rep.GetAllUsers(limit, limit*page)
	if err != nil {
		logger.Error("failed to get all users", "error", err.Error())
		return nil, err
	}

	logger.Info("users fetched successfully", "count", len(users))
	return toResponses(users), nil
}

// GetUser implements IUserUsecase.
func (u *UserUsecaseImpl) GetUser(role string, requesterID, ID uuid.UUID) (*dto.UserResponse, error) {
	logger := u.logger.With("method", "GetUser", "userID", ID.String())
	logger.Debug("starting get user")

	if !isOwner(requesterID, ID) && role != "admin" {
		logger.Info("access denied")
		return nil, apperrors.NewErrAccessDenied("forbidden")
	}

	user, err := u.rep.GetUser(ID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("user not found")
			return nil, err
		}
		logger.Error("failed to get user", "error", err.Error())
		return nil, err
	}

	logger.Info("user fetched successfully")
	return toResponse(user), nil
}

// UpdateUser implements IUserUsecase.
func (u *UserUsecaseImpl) UpdateUser(requesterID, ID uuid.UUID, req *dto.UpdateUserRequest) error {
	logger := u.logger.With("method", "UpdateUser", "userID", ID.String())
	logger.Debug("starting update user")

	if !isOwner(requesterID, ID) {
		logger.Info("access denied")
		return apperrors.NewErrAccessDenied("forbidden")
	}
	user, err := u.rep.GetUser(ID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("user to update not found")
			return err
		}
		logger.Error("failed to get user for update", "error", err.Error())
		return err
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	err = u.rep.UpdateUser(user)
	if err != nil {
		logger.Error("failed to update user", "error", err.Error())
		return err
	}

	logger.Info("user updated successfully")
	return nil
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

func isOwner(requesterID, userID uuid.UUID) bool {
	return requesterID == userID
}
