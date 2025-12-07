package usecase

import (
	"context"
	"log/slog"

	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/google/uuid"
)

type likeUsecase struct {
	likeRepo repository.LikeRepository
	logger   *slog.Logger
}

func NewLikeUsecase(likeRepo repository.LikeRepository, logger *slog.Logger) LikeUsecase {
	return &likeUsecase{
		likeRepo: likeRepo,
		logger:   logger,
	}
}

func (u *likeUsecase) LikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error {
	return u.likeRepo.LikeIdea(ctx, userID, ideaID)
}

func (u *likeUsecase) UnlikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error {
	return u.likeRepo.UnlikeIdea(ctx, userID, ideaID)
}
