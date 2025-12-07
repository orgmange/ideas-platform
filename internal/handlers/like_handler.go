package handlers

import (
	"log/slog"
	"net/http"

	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LikeHandler struct {
	usecase usecase.LikeUsecase
	logger  *slog.Logger
}

func NewLikeHandler(u usecase.LikeUsecase, l *slog.Logger) *LikeHandler {
	return &LikeHandler{
		usecase: u,
		logger:  l,
	}
}

func (h *LikeHandler) LikeIdea(c *gin.Context) {
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	ideaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid idea id"})
		return
	}

	err = h.usecase.LikeIdea(c.Request.Context(), userID, ideaID)
	if err != nil {
		h.logger.Error("failed to like idea", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like idea"})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *LikeHandler) UnlikeIdea(c *gin.Context) {
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	ideaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid idea id"})
		return
	}

	err = h.usecase.UnlikeIdea(c.Request.Context(), userID, ideaID)
	if err != nil {
		h.logger.Error("failed to unlike idea", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike idea"})
		return
	}

	c.Status(http.StatusNoContent)
}
