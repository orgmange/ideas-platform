package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthMiddleware(uc usecase.AuthUsecase, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Info("no authorization token", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := uc.ValidateJWTToken(c.Request.Context(), tokenString)
		if err != nil {
			handlers.HandleAppErrors(err, logger, c)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func AdminFilter(workerShopRepo repository.WorkerCoffeeShopRepository, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exist := c.Get("user_id")
		if !exist {
			logger.Info("user_id not found in context for AdminFilter", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "user_id not found in context"})
			c.Abort()
			return
		}

		userID, ok := userIDAny.(uuid.UUID)
		if !ok {
			logger.Error("failed to parse user_id from context in AdminFilter", slog.Any("user_id_type", userIDAny), "path", c.Request.URL.Path)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
			c.Abort()
			return
		}

		err := usecase.CheckAnyShopAdminAccess(c.Request.Context(), logger, workerShopRepo, userID)
		if err != nil {
			handlers.HandleAppErrors(err, logger, c)
			c.Abort()
			return
		}

		c.Next()
	}
}
