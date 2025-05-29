package main

import (
	"log"

	"trade/internal/infrastructure/config"
	"trade/internal/infrastructure/di"

	_ "trade/docs"

	"github.com/gofiber/swagger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	app, err := di.BuildApp(cfg)
	if err != nil {
		log.Fatalf("failed to build app: %v", err)
	}
	app.Get("/docs/*", swagger.HandlerDefault)

	addr := ":" + cfg.HTTPPort
	log.Printf("starting server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
