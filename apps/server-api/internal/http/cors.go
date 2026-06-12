package http

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

var allowedCORSMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodOptions,
}

var allowedCORSHeaders = []string{
	"Authorization",
	"Content-Type",
}

func allowCORS(cfg config.Config) gin.HandlerFunc {
	allowMethods := strings.Join(allowedCORSMethods, ", ")
	allowHeaders := strings.Join(allowedCORSHeaders, ", ")
	allowAllOrigins := len(cfg.CORSAllowedOrigins) == 0

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		switch {
		case allowAllOrigins:
			c.Header("Access-Control-Allow-Origin", "*")
		case origin != "" && slices.Contains(cfg.CORSAllowedOrigins, origin):
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", allowMethods)
		c.Header("Access-Control-Allow-Headers", allowHeaders)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
