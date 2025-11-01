package main

import (
	"fmt"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/GeorgiiMalishev/ideas-platform/internal/db"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/GeorgiiMalishev/ideas-platform/internal/router"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}
	err = db.RunMigrations(cfg)
	if err != nil {
		fmt.Println("Failed to run migrations:", err)
	}
	db, err := db.InitDB(cfg)
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("Failed to get sql.DB from gorm.DB:", err)
	}

	defer sqlDB.Close()
	if err = sqlDB.Ping(); err != nil {
		fmt.Println("Failed to ping database:", err)
	}
	userRepo := repository.NewUserRepository(sqlDB)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handlers.NewUserHandler(userUsecase)
	ar := router.NewRouter(cfg, userHandler)
	r := ar.SetupRouter()
	err = r.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
