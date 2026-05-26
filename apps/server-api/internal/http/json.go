package http

import "github.com/gin-gonic/gin"

func writeJSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

func writeJSONError(c *gin.Context, status int, message string) {
	writeJSON(c, status, map[string]string{"error": message})
}
