package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeacherCanRegisterAndLogin(t *testing.T) {
	handler := newTestHandler()

	registerRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/register", map[string]any{
		"username": "teacher-1",
		"password": "secret123",
	})

	require.Equal(t, http.StatusCreated, registerRes.Code)
	requireBodyField(t, registerRes.Body.String(), "teacherName", "teacher-1")
	registerToken := requireStringField(t, registerRes.Body.String(), "token")
	require.NotEmpty(t, registerToken)

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-1",
		"password": "secret123",
	})

	require.Equal(t, http.StatusOK, loginRes.Code)
	requireBodyField(t, loginRes.Body.String(), "teacherName", "teacher-1")
	loginToken := requireStringField(t, loginRes.Body.String(), "token")
	require.NotEmpty(t, loginToken)
}

func TestTeacherRegisterRejectsDuplicateUsername(t *testing.T) {
	handler := newTestHandler()

	firstRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/register", map[string]any{
		"username": "teacher-duplicate",
		"password": "secret123",
	})
	require.Equal(t, http.StatusCreated, firstRes.Code)

	secondRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/register", map[string]any{
		"username": "teacher-duplicate",
		"password": "secret123",
	})
	require.Equal(t, http.StatusConflict, secondRes.Code)
}

func TestTeacherLoginRejectsInvalidPassword(t *testing.T) {
	handler := newTestHandler()
	registerTeacher(t, handler, "teacher-invalid-login", "secret123")

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-invalid-login",
		"password": "wrong-pass",
	})
	require.Equal(t, http.StatusUnauthorized, loginRes.Code)
}

func TestTeacherAuthRejectsMissingRequiredFields(t *testing.T) {
	handler := newTestHandler()

	registerRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/register", map[string]any{
		"username": "teacher-missing-password",
	})
	require.Equal(t, http.StatusBadRequest, registerRes.Code)

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-missing-password",
	})
	require.Equal(t, http.StatusBadRequest, loginRes.Code)
}

func TestTeacherRouteRequiresBearerToken(t *testing.T) {
	handler := newTestHandler()

	meRes := performJSONRequest(t, handler, http.MethodGet, "/api/teacher/me", nil)
	require.Equal(t, http.StatusUnauthorized, meRes.Code)
}

func TestStudentTokenCannotCreateTeacherStudents(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-student-guard", "secret123")
	createStudent(t, handler, teacherToken, "student-guard", "小守", "abc12345")
	studentToken := loginStudent(t, handler, "student-guard", "abc12345")

	createRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, "/api/teacher/students/batch", map[string]any{
		"students": []map[string]any{
			{
				"username":        "student-guard-2",
				"displayName":     "小守二号",
				"initialPassword": "abc12345",
			},
		},
	})

	require.Equal(t, http.StatusUnauthorized, createRes.Code)
}

func TestTeacherCannotReadOtherTeachersAssignment(t *testing.T) {
	handler := newTestHandler()
	teacherTokenA := registerTeacher(t, handler, "teacher-assignment-owner", "secret123")
	teacherTokenB := registerTeacher(t, handler, "teacher-assignment-other", "secret123")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherTokenA, "Private Maze")

	detailRes := performAuthedJSONRequest(t, handler, teacherTokenB, http.MethodGet, "/api/teacher/assignments/"+fmt.Sprint(assignmentID), nil)
	require.Equal(t, http.StatusNotFound, detailRes.Code)
}
