package http

import "github.com/gin-gonic/gin"

type healthResponse struct {
	Status string `json:"status"`
}

func handleHealth(c *gin.Context) {
	writeJSON(c, 200, healthResponse{Status: "ok"})
}
