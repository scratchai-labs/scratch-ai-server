package http

import "github.com/gin-gonic/gin"

type healthResponse struct {
	Status string `json:"status"`
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
	writeJSON(c, 200, healthResponse{Status: "ok"})
}
