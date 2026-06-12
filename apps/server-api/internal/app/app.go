package app

import (
	"net/http"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
	httpapp "github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/http"
)

func New() http.Handler {
	handler, err := NewWithConfig(config.FromEnv())
	if err != nil {
		panic(err)
	}
	return handler
}

func NewWithConfig(cfg config.Config) (http.Handler, error) {
	if err := cfg.ValidateForRuntime(); err != nil {
		return nil, err
	}
	return httpapp.NewRouter(cfg)
}
