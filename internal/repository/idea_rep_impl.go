package repository

import (
	"strings"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type ideaRepository struct {
	db *gorm.DB
}

func NewIdeaRepository(db *gorm.DB) IdeaRepository {
	return &ideaRepository{db}
}

func (r *ideaRepository) CreateIdea(idea *models.Idea) (*models.Idea, error) {
	if err := r.db.Create(idea).Error; err != nil {
		return nil, err
	}
	return idea, nil
}

func (r *ideaRepository) GetIdea(ideaID uuid.UUID) (*models.Idea, error) {
	var idea models.Idea
	if err := r.db.Preload("IdeaStatus").First(&idea, "id = ?", ideaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewErrNotFound("idea", ideaID.String())
		}
		return nil, err
	}
	return &idea, nil
}

func (r *ideaRepository) GetAllIdeasByShop(shopID uuid.UUID, limit, offset int, sort string) ([]models.Idea, error) {
	var ideas []models.Idea
	query := r.db.Model(&models.Idea{}).Where("coffee_shop_id = ?", shopID)

	query = applyIdeaSorting(query, sort)

	err := query.Limit(limit).Offset(offset).Find(&ideas).Error
	return ideas, err
}

func (r *ideaRepository) GetAllIdeasByUser(userID uuid.UUID, limit, offset int, sort string) ([]models.Idea, error) {
	var ideas []models.Idea
	query := r.db.Model(&models.Idea{}).Where("creator_id = ?", userID)

	query = applyIdeaSorting(query, sort)

	err := query.Limit(limit).Offset(offset).Find(&ideas).Error
	return ideas, err
}

func (r *ideaRepository) UpdateIdea(idea *models.Idea) error {
	return r.db.Save(idea).Error
}

func (r *ideaRepository) DeleteIdea(ideaID uuid.UUID) error {
	result := r.db.Delete(&models.Idea{}, "id = ?", ideaID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.NewErrNotFound("idea", ideaID.String())
	}
	return nil
}

func applyIdeaSorting(query *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		// Default sort order
		return query.Order("created_at DESC")
	}

	sorts := strings.SplitSeq(sort, ",")
	for s := range sorts {
		direction := "ASC"
		if strings.HasPrefix(s, "-") {
			direction = "DESC"
			s = s[1:]
		}

		switch s {
		case "status":
			query = query.Order("status_id " + direction)
		case "created_at":
			query = query.Order("created_at " + direction)
		case "likes":
			query = query.Joins("LEFT JOIN idea_likes ON idea_likes.idea_id = ideas.id").
				Group("ideas.id").
				Order("COUNT(idea_likes.id) " + direction)
		}
	}
	return query
}

