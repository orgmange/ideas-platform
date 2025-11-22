package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	_ "github.com/GeorgiiMalishev/ideas-platform/docs"
	"github.com/GeorgiiMalishev/ideas-platform/internal/db"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/GeorgiiMalishev/ideas-platform/internal/router"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server
// @termsOfService http://swagger.io/terms/

// @BasePath /api/v1
// @host localhost:8080
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	db, err := db.InitDB(cfg)
	if err != nil {
		logger.Error("Failed to connect to database:", slog.String("error", err.Error()))
		return
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.BannedUser{},
		&models.Role{},
		&models.CoffeeShop{},
		&models.WorkerCoffeeShop{},
		&models.Category{},
		&models.Idea{},
		&models.IdeaLike{},
		&models.IdeaComment{},
		&models.Reward{},
		&models.RewardType{},
		&models.OTP{},
		&models.UserRefreshToken{},
	)
	if err != nil {
		logger.Error("Failed to auto-migrate database:", slog.String("error", err.Error()))
		return
	}

	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, logger)
	userHandler := handlers.NewUserHandler(userUsecase, logger)

	coffeeShopRepo := repository.NewCoffeeShopRepository(db)
	csUscase := usecase.NewCoffeeShopUsecase(coffeeShopRepo, logger)
	csHandler := handlers.NewCoffeeShopHandler(csUscase, logger)

	authRepo := repository.NewAuthRepository(db)
	authUsecase := usecase.NewAuthUsecase(authRepo, "1234567890", &cfg.AuthConfig, logger)
	authHandler := handlers.NewAuthHandler(authUsecase, logger)

	ideaRepo := repository.NewIdeaRepository(db)
	ideaUsecase := usecase.NewIdeaUsecase(ideaRepo, logger)
	ideaHandler := handlers.NewIdeaHandler(ideaUsecase, logger)

	ar := router.NewRouter(cfg, userHandler, csHandler, authHandler, ideaHandler, nil, authUsecase, logger)
	r := ar.SetupRouter()
	err = r.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
