package router

import (
	"log/slog"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/GeorgiiMalishev/ideas-platform/internal/middleware"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type AppRouter struct {
	cfg               *config.Config
	userHandler       *handlers.UserHandler
	coffeeShopHandler *handlers.CoffeeShopHandler
	authHandler       *handlers.AuthHandler
	ideaHandler       *handlers.IdeaHandler

	authUsecase usecase.AuthUsecase
	logger      *slog.Logger
}

func NewRouter(cfg *config.Config,
	userHandler *handlers.UserHandler,
	coffeeShopHandler *handlers.CoffeeShopHandler,
	authHandler *handlers.AuthHandler,
	ideaHandler *handlers.IdeaHandler,

	authUsecase usecase.AuthUsecase,
	logger *slog.Logger,
) *AppRouter {
	return &AppRouter{
		cfg:               cfg,
		userHandler:       userHandler,
		coffeeShopHandler: coffeeShopHandler,
		authHandler:       authHandler,
		ideaHandler:       ideaHandler,

		authUsecase: authUsecase,
		logger:      logger,
	}
}

func (ar AppRouter) SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/health", handlers.HealthCheck(ar.cfg))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		// users
		v1.GET("/users", ar.userHandler.GetAllUsers)
		v1.GET("/users/:id", ar.userHandler.GetUser)

		// coffee_shop
		v1.GET("/coffee-shops", ar.coffeeShopHandler.GetAllCoffeeShops)
		v1.GET("/coffee-shops/:id", ar.coffeeShopHandler.GetCoffeeShop)

		// auth
		v1.GET("/auth/:phone", ar.authHandler.GetOTP)
		v1.POST("/auth", ar.authHandler.VerifyOTP)
		v1.POST("/auth/refresh", ar.authHandler.Refresh)

	}

	authRequired := v1.Group("")
	authRequired.Use(middleware.AuthMiddleware(ar.authUsecase, ar.logger))
	{
		authRequired.GET("/users/me", ar.userHandler.GetCurrentAuthentificatedUser)
		authRequired.PUT("/users/:id", ar.userHandler.UpdateUser)
		authRequired.DELETE("/users/:id", ar.userHandler.DeleteUser)

		authRequired.POST("/logout", ar.authHandler.Logout)
		authRequired.POST("/logout-everywhere", ar.authHandler.LogoutEverywhere)

		authRequired.POST("/coffee-shops", ar.coffeeShopHandler.CreateCoffeeShop)
		authRequired.DELETE("/coffee-shops/:id", ar.coffeeShopHandler.DeleteCoffeeShop)
		authRequired.PUT("/coffee-shops/:id", ar.coffeeShopHandler.UpdateCoffeeShop)
	}

	return r
}
