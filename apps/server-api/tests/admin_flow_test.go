package tests

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBootstrapAdminCanLoginAndManageTeachers(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "admin",
		"password": "admin12345",
	})
	require.Equal(t, http.StatusOK, loginRes.Code)
	requireBodyField(t, loginRes.Body.String(), "teacherName", "admin")
	requireBodyField(t, loginRes.Body.String(), "role", "admin")
	adminToken := requireStringField(t, loginRes.Body.String(), "token")

	listRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodGet, "/api/admin/teachers", nil)
	require.Equal(t, http.StatusOK, listRes.Code)
	requireJSONArrayLen(t, listRes.Body.String(), "items", 1)
	requireBodyField(t, listRes.Body.String(), "items.0.username", "admin")
	requireBodyField(t, listRes.Body.String(), "items.0.role", "admin")
	requireBodyField(t, listRes.Body.String(), "items.0.status", "active")

	createRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers", map[string]any{
		"username":        "teacher-managed",
		"initialPassword": "secret123",
	})
	require.Equal(t, http.StatusCreated, createRes.Code)
	requireBodyField(t, createRes.Body.String(), "username", "teacher-managed")
	requireBodyField(t, createRes.Body.String(), "role", "teacher")
	requireBodyField(t, createRes.Body.String(), "status", "active")
}

func TestAdminCanResetDisableAndEnableTeacher(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()
	adminToken := loginTeacherToken(t, handler, "admin", "admin12345")

	createRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers", map[string]any{
		"username":        "teacher-reset",
		"initialPassword": "secret123",
	})
	require.Equal(t, http.StatusCreated, createRes.Code)
	teacherID := requireInt64Field(t, createRes.Body.String(), "id")

	oldLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-reset",
		"password": "secret123",
	})
	require.Equal(t, http.StatusOK, oldLoginRes.Code)
	requireBodyField(t, oldLoginRes.Body.String(), "role", "teacher")

	resetRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/reset-password", map[string]any{
		"newPassword": "updated123",
	})
	require.Equal(t, http.StatusOK, resetRes.Code)
	requireBodyField(t, resetRes.Body.String(), "status", "active")

	disabledRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/disable", nil)
	require.Equal(t, http.StatusOK, disabledRes.Code)
	requireBodyField(t, disabledRes.Body.String(), "status", "disabled")

	disabledLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-reset",
		"password": "updated123",
	})
	require.Equal(t, http.StatusForbidden, disabledLoginRes.Code)

	enabledRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/enable", nil)
	require.Equal(t, http.StatusOK, enabledRes.Code)
	requireBodyField(t, enabledRes.Body.String(), "status", "active")

	newLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-reset",
		"password": "updated123",
	})
	require.Equal(t, http.StatusOK, newLoginRes.Code)
}

func TestTeacherCannotAccessAdminRoutes(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()

	teacherToken := registerTeacher(t, handler, "teacher-no-admin", "secret123")

	listRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/admin/teachers", nil)
	require.Equal(t, http.StatusForbidden, listRes.Code)

	disableSelfRes := performAuthedJSONRequest(t, handler, loginTeacherToken(t, handler, "admin", "admin12345"), http.MethodPost, "/api/admin/teachers/1/disable", nil)
	require.Equal(t, http.StatusConflict, disableSelfRes.Code)
}

func loginTeacherToken(t *testing.T, handler http.Handler, username string, password string) string {
	t.Helper()

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": username,
		"password": password,
	})
	require.Equal(t, http.StatusOK, loginRes.Code)
	return requireStringField(t, loginRes.Body.String(), "token")
}
