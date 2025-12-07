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

	adminRole := models.Role{
		Name: "admin",
	}
	db.FirstOrCreate(&adminRole, "name = ?", "admin")
	adminRoleID := adminRole.ID

	workerCsRepo := repository.NewWorkerCoffeeShopRepository(db)

	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, workerCsRepo, logger)
	userHandler := handlers.NewUserHandler(userUsecase, logger)

	coffeeShopRepo := repository.NewCoffeeShopRepository(db)
	csUscase := usecase.NewCoffeeShopUsecase(coffeeShopRepo, workerCsRepo, adminRoleID, logger)
	csHandler := handlers.NewCoffeeShopHandler(csUscase, logger)

	authRepo := repository.NewAuthRepository(db)
	authUsecase := usecase.NewAuthUsecase(authRepo, "1234567890", &cfg.AuthConfig, logger)
	authHandler := handlers.NewAuthHandler(authUsecase, logger)

	ideaRepo := repository.NewIdeaRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	ideaUsecase := usecase.NewIdeaUsecase(ideaRepo, workerCsRepo, likeRepo, logger)
	ideaHandler := handlers.NewIdeaHandler(ideaUsecase, logger)

	likeUsecase := usecase.NewLikeUsecase(likeRepo, logger)
	likeHandler := handlers.NewLikeHandler(likeUsecase, logger)

	rewardRepo := repository.NewRewardRepository(db)
	rewardUsecase := usecase.NewRewardUsecase(rewardRepo, ideaRepo, logger)
	rewardHandler := handlers.NewRewardHandler(rewardUsecase, logger)

	rewardTypeRepo := repository.NewRewardTypeRepository(db)
	rewardTypeUsecase := usecase.NewRewardTypeUsecase(rewardTypeRepo, coffeeShopRepo, workerCsRepo, logger)
	rewardTypeHandler := handlers.NewRewardTypeHandler(rewardTypeUsecase, logger)

	workerCoffeeShopUsecase := usecase.NewWorkerCoffeeShopUsecase(workerCsRepo, coffeeShopRepo, userRepo, logger)
	workerCoffeeShopHandler := handlers.NewWorkerCoffeeShopHandler(workerCoffeeShopUsecase, logger)

	ar := router.NewRouter(cfg, userHandler, csHandler, authHandler, ideaHandler, rewardHandler, rewardTypeHandler, workerCoffeeShopHandler, likeHandler, workerCsRepo, authUsecase, logger)
	r := ar.SetupRouter()
	err = r.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
