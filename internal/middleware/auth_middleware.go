package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
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
		claims, err := uc.ValidateJWTToken(tokenString)
		if err != nil {
			handlers.HandleAppErrors(err, logger, c)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminFilter(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleAny, exist := c.Get("role")
		if !exist {
			logger.Info("user not authorized", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Message: "user not authorized"})
			c.Abort()
			return
		}
		role, ok := roleAny.(string)
		if !ok {
			logger.Info("unexpected happens when parsing user role", "path", c.Request.URL.Path)
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
			c.Abort()
			return
		}

		if role != "admin" {
			logger.Info("user not admin", "path", c.Request.URL.Path)
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Message: "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}
