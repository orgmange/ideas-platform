package usecase

import (
	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type IdeaUsecaseImpl struct {
	ideaRepo repository.IdeaRepository
}

func NewIdeaUsecase(ideaRepo repository.IdeaRepository) IdeaUsecase {
	return &IdeaUsecaseImpl{ideaRepo: ideaRepo}
}

func (u *IdeaUsecaseImpl) CreateIdea(userID uuid.UUID, req *dto.CreateIdeaRequest) (*dto.IdeaResponse, error) {
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
		return nil, err
	}

	return toIdeaResponse(createdIdea), nil
}

func (u *IdeaUsecaseImpl) GetIdea(ideaID uuid.UUID) (*dto.IdeaResponse, error) {
	idea, err := u.ideaRepo.GetIdea(ideaID)
	if err != nil {
		return nil, err
	}
	return toIdeaResponse(idea), nil
}

func (u *IdeaUsecaseImpl) GetAllIdeasByShop(shopID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error) {
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 25
	}
	if params.Page < 0 {
		params.Page = 0
	}

	ideas, err := u.ideaRepo.GetAllIdeasByShop(shopID, params.Limit, params.Page*params.Limit, params.Sort)
	if err != nil {
		return nil, err
	}
	return toIdeaResponses(ideas), nil
}

func (u *IdeaUsecaseImpl) GetAllIdeasByUser(userID uuid.UUID, params dto.GetIdeasRequest) ([]dto.IdeaResponse, error) {
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 25
	}
	if params.Page < 0 {
		params.Page = 0
	}

	ideas, err := u.ideaRepo.GetAllIdeasByUser(userID, params.Limit, params.Page*params.Limit, params.Sort)
	if err != nil {
		return nil, err
	}
	return toIdeaResponses(ideas), nil
}

func (u *IdeaUsecaseImpl) UpdateIdea(userID, ideaID uuid.UUID, req *dto.UpdateIdeaRequest) error {
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

	return u.ideaRepo.UpdateIdea(idea)
}

func (u *IdeaUsecaseImpl) DeleteIdea(userID, ideaID uuid.UUID) error {
	_, err := u.getIfCreator(userID, ideaID)
	if err != nil {
		return err
	}
	return u.ideaRepo.DeleteIdea(ideaID)
}

func (u *IdeaUsecaseImpl) getIfCreator(userID, ideaID uuid.UUID) (*models.Idea, error) {
	idea, err := u.ideaRepo.GetIdea(ideaID)
	if err != nil {
		return nil, err
	}
	if idea.CreatorID == nil || userID != *idea.CreatorID {
		return nil, apperrors.NewAuthErr("access denied")
	}
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
