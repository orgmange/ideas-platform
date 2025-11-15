package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func handleAppErrors(err error, logger *slog.Logger, c *gin.Context) {
	var errNotFound *apperrors.ErrNotFound
	var errNotValid *apperrors.ErrNotValid
	var authErr *apperrors.AuthErr
	if errors.As(err, &errNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if errors.As(err, &errNotValid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.As(err, &authErr) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}

	logger.Error("internal server error: ", slog.String("error", err.Error()))
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func parseUUID(logger *slog.Logger, c *gin.Context) (*uuid.UUID, bool) {
	uuidRaw := c.Param("id")
	uuid, err := uuid.Parse(uuidRaw)
	if err != nil {
		logger.Error("invalid uuid: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return nil, false
	}

	return &uuid, true
}
