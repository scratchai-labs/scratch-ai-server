package http

import (
	"context"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// handleHealth godoc
//
//	@Summary		Health check
//	@Description	Read a basic process health status for uptime checks.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	healthResponse
//	@Router			/health [get]
func handleHealth(c *gin.Context) {
	newHealthHandler(nil)(c)
}

func newHealthHandler(readinessCheck func(context.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if readinessCheck != nil {
			if err := readinessCheck(c.Request.Context()); err != nil {
				writeJSON(c, 503, healthResponse{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}

		writeJSON(c, 200, healthResponse{Status: "ok"})
	}
}
