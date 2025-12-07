package repository

import (
	"context"

	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) LikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error {
	like := models.IdeaLike{
		UserID: &userID,
		IdeaID: &ideaID,
	}
	return r.db.WithContext(ctx).Create(&like).Error
}

func (r *likeRepository) UnlikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND idea_id = ?", userID, ideaID).Delete(&models.IdeaLike{}).Error
}

func (r *likeRepository) GetLikesCount(ctx context.Context, ideaID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.IdeaLike{}).Where("idea_id = ?", ideaID).Count(&count).Error
	return count, err
}
