package repository

import (
	"context"

	"github.com/google/uuid"
)

type LikeRepository interface {
	LikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error
	UnlikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error
	GetLikesCount(ctx context.Context, ideaID uuid.UUID) (int64, error)
}
