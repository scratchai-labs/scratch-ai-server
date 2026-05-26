package main

import (
	"log"
	"net/http"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/app"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

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
