package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHealthHandlerReportsServiceUnavailableWhenReadinessFails(t *testing.T) {
	engine := gin.New()
	engine.GET("/health", newHealthHandler(func(context.Context) error {
		return errors.New("database unavailable")
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.JSONEq(t, `{"status":"error","message":"database unavailable"}`, recorder.Body.String())
}
