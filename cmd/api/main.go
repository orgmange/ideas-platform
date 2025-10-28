package main

import (
	"fmt"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/GeorgiiMalishev/ideas-platform/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}
	r := router.SetupRouter(cfg)
	err = r.Run(":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
