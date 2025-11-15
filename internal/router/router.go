package router

import (
	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type AppRouter struct {
	cfg               *config.Config
	userHandler       *handlers.UserHandler
	coffeeShopHandler *handlers.CoffeeShopHandler
	authHandler       *handlers.AuthHandler
}

func NewRouter(cfg *config.Config,
	userHandler *handlers.UserHandler,
	coffeeShopHandler *handlers.CoffeeShopHandler,
	authHandler *handlers.AuthHandler,
) *AppRouter {
	return &AppRouter{
		cfg:               cfg,
		userHandler:       userHandler,
		coffeeShopHandler: coffeeShopHandler,
		authHandler:       authHandler,
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
		v1.PUT("/users/:id", ar.userHandler.UpdateUser)
		v1.DELETE("/users/:id", ar.userHandler.DeleteUser)

		// coffe_shop
		v1.POST("/coffee-shops", ar.coffeeShopHandler.CreateCoffeeShop)
		v1.GET("/coffee-shops", ar.coffeeShopHandler.GetAllCoffeeShops)
		v1.GET("/coffee-shops/:id", ar.coffeeShopHandler.GetCoffeeShop)
		v1.PUT("/coffee-shops/:id", ar.coffeeShopHandler.UpdateCoffeeShop)
		v1.DELETE("/coffee-shops/:id", ar.coffeeShopHandler.DeleteCoffeeShop)

		// auth
		v1.GET("/auth/:phone", ar.authHandler.GetOTP)
		v1.POST("/auth", ar.authHandler.VerifyOTP)
	}

	return r
}
