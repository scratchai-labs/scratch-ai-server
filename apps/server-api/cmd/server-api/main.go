package main

import (
	"log"
	"net/http"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/app"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

// @title			Scratch AI Server API
// @version		1.0
// @description		OpenAPI contract for the Scratch AI teaching server.
// @BasePath		/
// @schemes		http https
// @securityDefinitions.apikey	BearerAuth
// @in				header
// @name			Authorization
func main() {
	cfg := config.FromEnv()
	handler, err := app.NewWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	log.Printf("server-api listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
