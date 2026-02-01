package main

import (
	"account-manager/config"
	"account-manager/internal/app"
	"log"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("Application failed: %s", err)
	}

}
