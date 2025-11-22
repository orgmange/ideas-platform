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
	rewardHandler     *handlers.RewardHandler
	rewardTypeHandler *handlers.RewardTypeHandler

	authUsecase usecase.AuthUsecase
	logger      *slog.Logger
}

func NewRouter(cfg *config.Config,
	userHandler *handlers.UserHandler,
	coffeeShopHandler *handlers.CoffeeShopHandler,
	authHandler *handlers.AuthHandler,
	ideaHandler *handlers.IdeaHandler,
	rewardHandler *handlers.RewardHandler,
	rewardTypeHandler *handlers.RewardTypeHandler,

	authUsecase usecase.AuthUsecase,
	logger *slog.Logger,
) *AppRouter {
	return &AppRouter{
		cfg:               cfg,
		userHandler:       userHandler,
		coffeeShopHandler: coffeeShopHandler,
		authHandler:       authHandler,
		ideaHandler:       ideaHandler,
		rewardHandler:     rewardHandler,
		rewardTypeHandler: rewardTypeHandler,

		authUsecase: authUsecase,
		logger:      logger,
	}
}

func (ar AppRouter) SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		// coffee_shop
		v1.GET("/coffee-shops", ar.coffeeShopHandler.GetAllCoffeeShops)
		v1.GET("/coffee-shops/:id", ar.coffeeShopHandler.GetCoffeeShop)

		// auth
		v1.GET("/auth/:phone", ar.authHandler.GetOTP)
		v1.POST("/auth", ar.authHandler.VerifyOTP)
		v1.POST("/auth/refresh", ar.authHandler.Refresh)

		// ideas
		v1.GET("/ideas/:id", ar.ideaHandler.GetIdea)
		v1.GET("/coffee-shops/:id/ideas", ar.ideaHandler.GetIdeasFromShop)

		// rewards
		v1.GET("/rewards/:id", ar.rewardHandler.GetReward)
	}

	authRequired := v1.Group("")
	authRequired.Use(middleware.AuthMiddleware(ar.authUsecase, ar.logger))
	{
		// users
		authRequired.GET("/users", ar.userHandler.GetAllUsers)
		authRequired.GET("/users/:id", ar.userHandler.GetUser)
		authRequired.GET("/users/me", ar.userHandler.GetCurrentAuthentificatedUser)
		authRequired.PUT("/users/:id", ar.userHandler.UpdateUser)
		authRequired.DELETE("/users/:id", ar.userHandler.DeleteUser)
		authRequired.GET("/users/me/rewards", ar.rewardHandler.GetMyRewards)
		authRequired.GET("/users/me/ideas", ar.ideaHandler.GetIdeasFromUser)

		// auth
		authRequired.POST("/logout", ar.authHandler.Logout)
		authRequired.POST("/logout-everywhere", ar.authHandler.LogoutEverywhere)

		// coffee-shops
		authRequired.POST("/coffee-shops", ar.coffeeShopHandler.CreateCoffeeShop)
		authRequired.DELETE("/coffee-shops/:id", ar.coffeeShopHandler.DeleteCoffeeShop)
		authRequired.PUT("/coffee-shops/:id", ar.coffeeShopHandler.UpdateCoffeeShop)
		authRequired.GET("/coffee-shops/:id/rewards", ar.rewardHandler.GetRewardsForCoffeeShop)
		authRequired.GET("/coffee-shops/:id/rewards/type", ar.rewardTypeHandler.GetRewardTypesByCoffeeShop)

		// ideas
		authRequired.POST("/ideas", ar.ideaHandler.CreateIdea)
		authRequired.PUT("/ideas/:id", ar.ideaHandler.UpdateIdea)
		authRequired.DELETE("/ideas/:id", ar.ideaHandler.DeleteIdea)

		authRequired.GET("/rewards/type/:id", ar.rewardTypeHandler.GetRewardType)
	}

	adminRequired := authRequired.Group("/admin")
	adminRequired.Use(middleware.AdminFilter(ar.logger))
	{
		adminRequired.GET("/health", handlers.HealthCheck(ar.cfg))
		adminRequired.POST("/rewards", ar.rewardHandler.GiveReward)
		adminRequired.DELETE("/rewards/:id", ar.rewardHandler.RevokeReward)

		adminRequired.POST("/rewards/type", ar.rewardTypeHandler.CreateRewardType)
		adminRequired.PUT("/rewards/type/:id", ar.rewardTypeHandler.UpdateRewardType)
		adminRequired.DELETE("/rewards/type/:id", ar.rewardTypeHandler.DeleteRewardType)
	}
	return r
}
