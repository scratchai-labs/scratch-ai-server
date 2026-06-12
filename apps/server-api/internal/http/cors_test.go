package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

func TestAllowCORSReflectsConfiguredOrigin(t *testing.T) {
	engine := gin.New()
	engine.Use(allowCORS(config.Config{
		CORSAllowedOrigins: []string{"https://teacher.example"},
	}))
	engine.POST("/api/teacher/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/teacher/login", nil)
	req.Header.Set("Origin", "https://teacher.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "authorization, content-type")

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Equal(t, "https://teacher.example", recorder.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "Origin", recorder.Header().Get("Vary"))
}

func TestAllowCORSDropsDisallowedOrigin(t *testing.T) {
	engine := gin.New()
	engine.Use(allowCORS(config.Config{
		CORSAllowedOrigins: []string{"https://teacher.example"},
	}))
	engine.POST("/api/teacher/login", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/teacher/login", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	require.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
}
