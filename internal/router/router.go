package router

import (
	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/gin-gonic/gin"
)

type AppRouter struct {
	cfg         *config.Config
	userHandler *handlers.UserHandler
}

func NewRouter(cfg *config.Config, userHandler *handlers.UserHandler) *AppRouter {
	return &AppRouter{cfg: cfg, userHandler: userHandler}
}

func (ar AppRouter) SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", handlers.HealthCheck(ar.cfg))
	r.POST("/users", ar.userHandler.CreateUser)
	r.GET("/users", ar.userHandler.GetAllUsers)
	r.DELETE("/users/:id", ar.userHandler.DeleteUser)
	r.PUT("/users/:id", ar.userHandler.UpdateUser)
	r.GET("/users/:id", ar.userHandler.GetUser)
	return r
}
