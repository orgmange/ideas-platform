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

type IdeaUsecaseImpl struct {
	ideaRepo     repository.IdeaRepository
	workerCsRepo repository.WorkerCoffeeShopRepository
	likeRepo     repository.LikeRepository
	logger       *slog.Logger
}

func NewIdeaUsecase(ideaRepo repository.IdeaRepository, workerCsRepo repository.WorkerCoffeeShopRepository, likeRepo repository.LikeRepository, logger *slog.Logger) IdeaUsecase {
	return &IdeaUsecaseImpl{
		ideaRepo:     ideaRepo,
		workerCsRepo: workerCsRepo,
		likeRepo:     likeRepo,
		logger:       logger,
	}
}

func (u *IdeaUsecaseImpl) CreateIdea(ctx context.Context, userID uuid.UUID, req *dto.CreateIdeaRequest) (*dto.IdeaResponse, error) {
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

	createdIdea, err := u.ideaRepo.CreateIdea(ctx, idea)
	if err != nil {
		logger.Error("failed to create idea", "error", err.Error())
		return nil, err
	}

	logger.Info("idea created successfully", "ideaID", createdIdea.ID.String())
	return toIdeaResponse(createdIdea, 0), nil
}

func (u *IdeaUsecaseImpl) GetIdea(ctx context.Context, ideaID uuid.UUID) (*dto.IdeaResponse, error) {
	logger := u.logger.With("method", "GetIdea", "ideaID", ideaID.String())
	logger.Debug("starting get idea")

	idea, err := u.ideaRepo.GetIdea(ctx, ideaID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("idea not found")
			return nil, err
		}
		logger.Error("failed to get idea", "error", err.Error())
		return nil, err
	}

	likes, err := u.likeRepo.GetLikesCount(ctx, idea.ID)
	if err != nil {
		logger.Error("failed to get likes count", "error", err.Error())
		return nil, err
	}

	logger.Info("idea fetched successfully")
	return toIdeaResponse(idea, int(likes)), nil
}

func (u *IdeaUsecaseImpl) GetAllIdeasByShop(ctx context.Context, shopID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error) {
	logger := u.logger.With("method", "GetAllIdeasByShop", "shopID", shopID.String(), "page", params.Page, "limit", params.Limit, "sort", params.Sort)
	logger.Debug("starting get all ideas by shop")

	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 25
	}
	if params.Page < 0 {
		params.Page = 0
	}

	ideas, err := u.ideaRepo.GetAllIdeasByShop(ctx, shopID, params.Limit, params.Page*params.Limit, params.Sort)
	if err != nil {
		logger.Error("failed to get all ideas by shop", "error", err.Error())
		return nil, err
	}

	logger.Info("ideas by shop fetched successfully", "count", len(ideas))
	return toIdeaResponses(ctx, ideas, u.likeRepo), nil
}

func (u *IdeaUsecaseImpl) GetAllIdeasByUser(ctx context.Context, userID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error) {
	logger := u.logger.With("method", "GetAllIdeasByUser", "userID", userID.String(), "page", params.Page, "limit", params.Limit, "sort", params.Sort)
	logger.Debug("starting get all ideas by user")

	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 25
	}
	if params.Page < 0 {
		params.Page = 0
	}

	ideas, err := u.ideaRepo.GetAllIdeasByUser(ctx, userID, params.Limit, params.Page*params.Limit, params.Sort)
	if err != nil {
		logger.Error("failed to get all ideas by user", "error", err.Error())
		return nil, err
	}

	logger.Info("ideas by user fetched successfully", "count", len(ideas))
	return toIdeaResponses(ctx, ideas, u.likeRepo), nil
}

func (u *IdeaUsecaseImpl) UpdateIdea(ctx context.Context, userID, ideaID uuid.UUID, req *dto.UpdateIdeaRequest) error {
	logger := u.logger.With("method", "UpdateIdea", "userID", userID.String(), "ideaID", ideaID.String())
	logger.Debug("starting update idea")

	idea, err := u.getIfCreator(ctx, userID, ideaID)
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

	err = u.ideaRepo.UpdateIdea(ctx, idea)
	if err != nil {
		logger.Error("failed to update idea", "error", err.Error())
		return err
	}

	logger.Info("idea updated successfully")
	return nil
}

func (u *IdeaUsecaseImpl) DeleteIdea(ctx context.Context, userID, ideaID uuid.UUID) error {
	logger := u.logger.With("method", "DeleteIdea", "userID", userID.String(), "ideaID", ideaID.String())
	logger.Debug("starting delete idea")

	idea, err := u.ideaRepo.GetIdea(ctx, ideaID)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("idea to delete not found")
			return err
		}
		logger.Error("failed to get idea for deletion check", "error", err.Error())
		return err
	}

	// Check if user is the creator
	isCreator := idea.CreatorID != nil && userID == *idea.CreatorID

	// If not the creator, check if they are an admin of the coffee shop
	if !isCreator {
		if idea.CoffeeShopID == nil {
			logger.Info("access denied: idea has no coffee shop and user is not creator")
			return apperrors.NewErrAccessDenied("access denied")
		}
		err := CheckShopAdminAccess(ctx, u.logger, u.workerCsRepo, userID, *idea.CoffeeShopID)
		if err != nil {
			logger.Info("access denied: user is not creator or shop admin")
			return err
		}
	}

	err = u.ideaRepo.DeleteIdea(ctx, ideaID)
	if err != nil {
		logger.Error("failed to delete idea", "error", err.Error())
		return err
	}

	logger.Info("idea deleted successfully")
	return nil
}

func (u *IdeaUsecaseImpl) getIfCreator(ctx context.Context, userID, ideaID uuid.UUID) (*models.Idea, error) {
	logger := u.logger.With("method", "getIfCreator", "userID", userID.String(), "ideaID", ideaID.String())
	logger.Debug("checking if user is creator of idea")

	idea, err := u.ideaRepo.GetIdea(ctx, ideaID)
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

func toIdeaResponse(idea *models.Idea, likes int) *dto.IdeaResponse {
	return &dto.IdeaResponse{
		ID:           idea.ID,
		CreatorID:    idea.CreatorID,
		CoffeeShopID: idea.CoffeeShopID,
		CategoryID:   idea.CategoryID,
		StatusID:     idea.StatusID,
		Title:        idea.Title,
		Description:  idea.Description,
		ImageURL:     idea.ImageURL,
		Likes:        likes,
	}
}

func toIdeaResponses(ctx context.Context, ideas []models.Idea, likeRepo repository.LikeRepository) []dto.IdeaResponse {
	res := make([]dto.IdeaResponse, len(ideas))
	for i, idea := range ideas {
		likes, err := likeRepo.GetLikesCount(ctx, idea.ID)
		if err != nil {
			// In case of an error, we can log it and set likes to 0
			// Or we can return the error, but that would complicate the calling functions
			// For now, let's log and default to 0
			// logger.Error("failed to get likes count for idea", "ideaID", idea.ID, "error", err)
			likes = 0
		}
		res[i] = *toIdeaResponse(&idea, int(likes))
	}
	return res
}
