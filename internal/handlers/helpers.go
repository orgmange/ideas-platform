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
	logger.Info("get error from app", "error", err.Error())
	var errNotFound *apperrors.ErrNotFound
	var errNotValid *apperrors.ErrNotValid
	var authErr *apperrors.ErrUnauthorized
	var errRateLimit *apperrors.ErrRateLimit
	var errAccessDenied *apperrors.ErrAccessDenied
	if errors.As(err, &errNotFound) {
		c.JSON(http.StatusNotFound, &dto.ErrorResponse{Message: err.Error()})
		return
	}
	if errors.As(err, &errNotValid) {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{Message: err.Error()})
		return
	}
	if errors.As(err, &authErr) {
		c.JSON(http.StatusUnauthorized, &dto.ErrorResponse{Message: err.Error()})
		return
	}
	if errors.As(err, &errRateLimit) {
		c.JSON(http.StatusTooManyRequests, &dto.ErrorResponse{Message: err.Error()})
		return
	}
	if errors.As(err, &errAccessDenied) {
		c.JSON(http.StatusForbidden, &dto.ErrorResponse{Message: err.Error()})
		return
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

func parseUserIDFromContext(logger *slog.Logger, c *gin.Context) (uuid.UUID, bool) {
	userIDAny, exist := c.Get("user_id")
	if !exist {
		logger.Info("unathorized user", "path", c.Request.URL.Path)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "user not authorized"})
		return uuid.Nil, false
	}

	userID, ok := userIDAny.(uuid.UUID)
	if !ok {
		logger.Error("unexpected err", "path", c.Request.URL.Path, "user id", userID.String())
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return uuid.Nil, false
	}

	return userID, true
}

func parseRoleFromContext(logger *slog.Logger, c *gin.Context) (string, bool) {
	roleAny, exist := c.Get("role")
	if !exist {
		logger.Info("unathorized user", "path", c.Request.URL.Path)
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "user not authorized"})
		return "", false
	}

	role, ok := roleAny.(string)
	if !ok {
		logger.Error("unexpected err", "path", c.Request.URL.Path, "role", role)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return "", false
	}

	return role, true
}
