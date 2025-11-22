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

type IdeaUsecaseImpl struct {
	ideaRepo repository.IdeaRepository
	logger   *slog.Logger
}

func NewIdeaUsecase(ideaRepo repository.IdeaRepository, logger *slog.Logger) IdeaUsecase {
	return &IdeaUsecaseImpl{ideaRepo: ideaRepo, logger: logger}
}

func (u *IdeaUsecaseImpl) CreateIdea(userID uuid.UUID, req *dto.CreateIdeaRequest) (*dto.IdeaResponse, error) {
	logger := u.logger.With("method", "CreateIdea", "userID", userID.String())
	logger.Debug("starting create idea")

	idea := &models.Idea{
		CreatorID:    &userID,
		CoffeeShopID: &req.CoffeeShopID,
		CategoryID:   &req.CategoryID,
		Title:        req.Title,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
	}

	createdIdea, err := u.ideaRepo.CreateIdea(idea)
	if err != nil {
		logger.Error("failed to create idea", "error", err.Error())
		return nil, err
	}

	logger.Info("idea created successfully", "ideaID", createdIdea.ID.String())
	return toIdeaResponse(createdIdea), nil
}

func (u *IdeaUsecaseImpl) GetIdea(ideaID uuid.UUID) (*dto.IdeaResponse, error) {
	logger := u.logger.With("method", "GetIdea", "ideaID", ideaID.String())
	logger.Debug("starting get idea")

	idea, err := u.ideaRepo.GetIdea(ideaID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("idea not found")
			return nil, err
		}
		logger.Error("failed to get idea", "error", err.Error())
		return nil, err
	}

	logger.Info("idea fetched successfully")
	return toIdeaResponse(idea), nil
}

func (u *IdeaUsecaseImpl) GetAllIdeasByShop(shopID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error) {
	logger := u.logger.With("method", "GetAllIdeasByShop", "shopID", shopID.String(), "page", params.Page, "limit", params.Limit, "sort", params.Sort)
	logger.Debug("starting get all ideas by shop")

	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 25
	}
	if params.Page < 0 {
		params.Page = 0
	}

	ideas, err := u.ideaRepo.GetAllIdeasByShop(shopID, params.Limit, params.Page*params.Limit, params.Sort)
	if err != nil {
		logger.Error("failed to get all ideas by shop", "error", err.Error())
		return nil, err
	}

	logger.Info("ideas by shop fetched successfully", "count", len(ideas))
	return toIdeaResponses(ideas), nil
}

func (u *IdeaUsecaseImpl) GetAllIdeasByUser(userID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error) {
	logger := u.logger.With("method", "GetAllIdeasByUser", "userID", userID.String(), "page", params.Page, "limit", params.Limit, "sort", params.Sort)
	logger.Debug("starting get all ideas by user")

	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 25
	}
	if params.Page < 0 {
		params.Page = 0
	}

	ideas, err := u.ideaRepo.GetAllIdeasByUser(userID, params.Limit, params.Page*params.Limit, params.Sort)
	if err != nil {
		logger.Error("failed to get all ideas by user", "error", err.Error())
		return nil, err
	}

	logger.Info("ideas by user fetched successfully", "count", len(ideas))
	return toIdeaResponses(ideas), nil
}

func (u *IdeaUsecaseImpl) UpdateIdea(userID, ideaID uuid.UUID, req *dto.UpdateIdeaRequest) error {
	logger := u.logger.With("method", "UpdateIdea", "userID", userID.String(), "ideaID", ideaID.String())
	logger.Debug("starting update idea")

	idea, err := u.getIfCreator(userID, ideaID)
	if err != nil {
		return err
	}

	if req.CategoryID != nil {
		idea.CategoryID = req.CategoryID
	}
	if req.StatusID != nil {
		idea.StatusID = req.StatusID
	}
	if req.Title != nil {
		idea.Title = *req.Title
	}
	if req.Description != nil {
		idea.Description = *req.Description
	}
	if req.ImageURL != nil {
		idea.ImageURL = req.ImageURL
	}

	err = u.ideaRepo.UpdateIdea(idea)
	if err != nil {
		logger.Error("failed to update idea", "error", err.Error())
		return err
	}

	logger.Info("idea updated successfully")
	return nil
}

func (u *IdeaUsecaseImpl) DeleteIdea(userID, ideaID uuid.UUID) error {
	logger := u.logger.With("method", "DeleteIdea", "userID", userID.String(), "ideaID", ideaID.String())
	logger.Debug("starting delete idea")

	_, err := u.getIfCreator(userID, ideaID)
	if err != nil {
		return err
	}

	err = u.ideaRepo.DeleteIdea(ideaID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("idea to delete not found")
			return err
		}
		logger.Error("failed to delete idea", "error", err.Error())
		return err
	}

	logger.Info("idea deleted successfully")
	return nil
}

func (u *IdeaUsecaseImpl) getIfCreator(userID, ideaID uuid.UUID) (*models.Idea, error) {
	logger := u.logger.With("method", "getIfCreator", "userID", userID.String(), "ideaID", ideaID.String())
	logger.Debug("checking if user is creator of idea")

	idea, err := u.ideaRepo.GetIdea(ideaID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("idea not found for creator check")
			return nil, err
		}
		logger.Error("failed to get idea for creator check", "error", err.Error())
		return nil, err
	}

	if idea.CreatorID == nil || userID != *idea.CreatorID {
		logger.Info("user is not creator of the idea", "creatorID", idea.CreatorID)
		return nil, apperrors.NewErrUnauthorized("access denied")
	}

	logger.Debug("user is confirmed as creator")
	return idea, nil
}

func toIdeaResponse(idea *models.Idea) *dto.IdeaResponse {
	return &dto.IdeaResponse{
		ID:           idea.ID,
		CreatorID:    idea.CreatorID,
		CoffeeShopID: idea.CoffeeShopID,
		CategoryID:   idea.CategoryID,
		StatusID:     idea.StatusID,
		Title:        idea.Title,
		Description:  idea.Description,
		ImageURL:     idea.ImageURL,
	}
}

func toIdeaResponses(ideas []models.Idea) []dto.IdeaResponse {
	res := make([]dto.IdeaResponse, len(ideas))
	for i, idea := range ideas {
		res[i] = *toIdeaResponse(&idea)
	}
	return res
}
