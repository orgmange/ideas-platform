package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func HandleAppErrors(err error, logger *slog.Logger, c *gin.Context) {
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

func parseUUID(logger *slog.Logger, c *gin.Context) (uuid.UUID, bool) {
	uuidRaw := c.Param("id")
	id, err := uuid.Parse(uuidRaw)
	if err != nil {
		logger.Error("invalid uuid: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return uuid.Nil, false
	}

	return id, true
}

func parseUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDAny, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "user not authorized"})
		return uuid.Nil, false
	}

	userID, ok := userIDAny.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return uuid.Nil, false
	}

	return userID, true
}
