package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	_ "github.com/GeorgiiMalishev/ideas-platform/docs"
	"github.com/GeorgiiMalishev/ideas-platform/internal/db"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
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
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	err = db.RunMigrations(cfg)
	if err != nil {
		logger.Error("Failed to run migrations:", slog.String("error", err.Error()))
	}

	db, err := db.InitDB(cfg)
	if err != nil {
		logger.Error("Failed to connect to database:", slog.String("error", err.Error()))
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get sql.DB from gorm.DB:", slog.String("error", err.Error()))
		return
	}

	defer sqlDB.Close()
	if err = sqlDB.Ping(); err != nil {
		logger.Error("Failed to ping database:", slog.String("error", err.Error()))
		return
	}

	userRepo := repository.NewUserRepository(sqlDB)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handlers.NewUserHandler(userUsecase, logger)

	coffeeShopRepo := repository.NewCoffeeShopRepository(sqlDB)
	csUscase := usecase.NewCoffeeShopUsecase(coffeeShopRepo)
	csHandler := handlers.NewCoffeeShopHandler(csUscase, logger)

	authRepo := repository.NewAuthRepository(db)
	authUsecase := usecase.NewAuthUsecase(authRepo, "1234567890")
	authHandler := handlers.NewAuthHandler(authUsecase, logger)

	ar := router.NewRouter(cfg, userHandler, csHandler, authHandler)
	r := ar.SetupRouter()
	err = r.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
