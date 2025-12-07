package usecase

import (
	"context"

	"github.com/google/uuid"
)

type LikeUsecase interface {
	LikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error
	UnlikeIdea(ctx context.Context, userID, ideaID uuid.UUID) error
}
