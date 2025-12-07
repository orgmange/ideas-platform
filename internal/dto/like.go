package dto

import "github.com/google/uuid"

type LikeRequest struct {
	IdeaID uuid.UUID `json:"idea_id"`
}
