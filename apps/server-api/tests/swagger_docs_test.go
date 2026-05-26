package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/app"
)

func TestSwaggerDocJSONExposesCoreAPIContract(t *testing.T) {
	handler := app.New()

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)

	var spec map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &spec))

	require.Equal(t, "2.0", spec["swagger"])

	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok, "swagger spec should expose paths")
	require.Equal(t, []string{
		"/api/student/assignments",
		"/api/student/assignments/{id}",
		"/api/student/assignments/{id}/hints",
		"/api/student/assignments/{id}/progress",
		"/api/student/login",
		"/api/student/logout",
		"/api/student/me",
		"/api/teacher/assignments",
		"/api/teacher/assignments/{id}",
		"/api/teacher/assignments/{id}/analysis",
		"/api/teacher/assignments/{id}/archive",
		"/api/teacher/assignments/{id}/assign-students",
		"/api/teacher/assignments/{id}/publish",
		"/api/teacher/dashboard/assignments/{id}/live",
		"/api/teacher/dashboard/students/{id}/history",
		"/api/teacher/login",
		"/api/teacher/logout",
		"/api/teacher/me",
		"/api/teacher/register",
		"/api/teacher/students",
		"/api/teacher/students/batch",
		"/api/teacher/students/{id}/reset-password",
		"/health",
	}, sortedKeys(paths))

	studentsPath, ok := paths["/api/teacher/students"].(map[string]any)
	require.True(t, ok, "students path should exist")

	getOperation, ok := studentsPath["get"].(map[string]any)
	require.True(t, ok, "students GET should exist")
	require.Nil(t, getOperation["parameters"], "students GET should not require a request body")
	getResponses, ok := getOperation["responses"].(map[string]any)
	require.True(t, ok, "students GET should expose responses")
	require.NotContains(t, getResponses, "201", "students GET should not document create semantics")

	postOperation, ok := studentsPath["post"].(map[string]any)
	require.True(t, ok, "students POST should exist")
	postParameters, ok := postOperation["parameters"].([]any)
	require.True(t, ok, "students POST should document a request body")
	require.Len(t, postParameters, 1)
	postResponses, ok := postOperation["responses"].(map[string]any)
	require.True(t, ok, "students POST should expose responses")
	require.Contains(t, postResponses, "201", "students POST should document create semantics")

	securityDefinitions, ok := spec["securityDefinitions"].(map[string]any)
	require.True(t, ok, "swagger spec should expose securityDefinitions")
	require.Contains(t, securityDefinitions, "BearerAuth")
}

func TestSwaggerUIIsServed(t *testing.T) {
	handler := app.New()

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Swagger UI")
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
