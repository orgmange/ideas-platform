package handlers

import (
	"net/http"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/gin-gonic/gin"
)

func HealthCheck(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "ideas-platform",
			"version": cfg.App.Version,
			"env":     cfg.App.Env,
		})
	}
}
