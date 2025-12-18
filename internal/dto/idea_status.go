package dto

import "github.com/google/uuid"

type CreateIdeaStatusRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateIdeaStatusRequest struct {
	Title string `json:"title" binding:"required"`
}

type IdeaStatusResponse struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}
