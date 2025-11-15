package handlers

import (
	"net/http"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/gin-gonic/gin"
)

// @Summary Health Check
// @Description Check the health of the service
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
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
