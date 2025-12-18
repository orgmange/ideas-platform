package usecase

import (
	"context"
	"errors"
	"log/slog"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type IdeaStatusUsecaseImpl struct {
	statusRepo repository.IdeaStatusRepository
	logger     *slog.Logger
}

func NewIdeaStatusUsecase(statusRepo repository.IdeaStatusRepository, logger *slog.Logger) IdeaStatusUsecase {
	return &IdeaStatusUsecaseImpl{
		statusRepo: statusRepo,
		logger:     logger,
	}
}

func (u *IdeaStatusUsecaseImpl) Create(ctx context.Context, req dto.CreateIdeaStatusRequest) (uuid.UUID, error) {
	logger := u.logger.With("method", "CreateStatus", "title", req.Title)
	
	// Check if title already exists
	existing, err := u.statusRepo.GetByTitle(ctx, req.Title)
	if err == nil {
		// If found (no error), then it's a conflict
		logger.Warn("idea status already exists", "id", existing.ID)
		return uuid.Nil, apperrors.NewErrConflict("idea status with this title already exists")
	}

	// Expecting NotFound error here
	var errNotFound *apperrors.ErrNotFound
	if !errors.As(err, &errNotFound) {
		logger.Error("failed to check existing status", "error", err.Error())
		return uuid.Nil, err
	}

	newStatus := &models.IdeaStatus{
		Title: req.Title,
	}

	id, err := u.statusRepo.Create(ctx, newStatus)
	if err != nil {
		logger.Error("failed to create idea status", "error", err.Error())
		return uuid.Nil, err
	}

	logger.Info("idea status created successfully", "id", id)
	return id, nil
}

func (u *IdeaStatusUsecaseImpl) Update(ctx context.Context, id uuid.UUID, req dto.UpdateIdeaStatusRequest) error {
	logger := u.logger.With("method", "UpdateStatus", "id", id)

	existing, err := u.statusRepo.GetByID(ctx, id)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Warn("idea status not found for update")
			return err
		}
		logger.Error("failed to get idea status", "error", err.Error())
		return err
	}

	// Check title uniqueness if changed
	if existing.Title != req.Title {
		_, err := u.statusRepo.GetByTitle(ctx, req.Title)
		if err == nil {
			return apperrors.NewErrConflict("idea status with this title already exists")
		}
		var errNotFound *apperrors.ErrNotFound
		if !errors.As(err, &errNotFound) {
			return err
		}
	}

	existing.Title = req.Title
	if err := u.statusRepo.Update(ctx, &existing); err != nil {
		logger.Error("failed to update idea status", "error", err.Error())
		return err
	}

	logger.Info("idea status updated successfully")
	return nil
}

func (u *IdeaStatusUsecaseImpl) Delete(ctx context.Context, id uuid.UUID) error {
	logger := u.logger.With("method", "DeleteStatus", "id", id)

	if err := u.statusRepo.Delete(ctx, id); err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Warn("idea status not found for deletion")
			return err
		}
		logger.Error("failed to delete idea status", "error", err.Error())
		return err
	}

	logger.Info("idea status deleted successfully")
	return nil
}

func (u *IdeaStatusUsecaseImpl) GetByID(ctx context.Context, id uuid.UUID) (dto.IdeaStatusResponse, error) {
	status, err := u.statusRepo.GetByID(ctx, id)
	if err != nil {
		// Repo returns apperrors.ErrNotFound, so we just pass it
		return dto.IdeaStatusResponse{}, err
	}

	return dto.IdeaStatusResponse{
		ID:    status.ID,
		Title: status.Title,
	}, nil
}

func (u *IdeaStatusUsecaseImpl) GetAll(ctx context.Context) ([]dto.IdeaStatusResponse, error) {
	statuses, err := u.statusRepo.GetAll(ctx)
	if err != nil {
		u.logger.Error("failed to get all statuses", "error", err.Error())
		return nil, err
	}

	var responses []dto.IdeaStatusResponse
	for _, s := range statuses {
		responses = append(responses, dto.IdeaStatusResponse{
			ID:    s.ID,
			Title: s.Title,
		})
	}

	return responses, nil
}
