package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

func allowCORS() gin.HandlerFunc {
	allowMethods := strings.Join(allowedCORSMethods, ", ")
	allowHeaders := strings.Join(allowedCORSHeaders, ", ")

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", allowMethods)
		c.Header("Access-Control-Allow-Headers", allowHeaders)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
