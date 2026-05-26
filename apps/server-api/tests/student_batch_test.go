package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeacherCanBatchCreateStudents(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-batch", "secret123")

	firstBatch := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/students/batch", map[string]any{
		"students": []map[string]any{
			{
				"username":        "student-1",
				"displayName":     "小明",
				"initialPassword": "abc12345",
			},
		},
	})

	require.Equal(t, http.StatusCreated, firstBatch.Code)
	requireJSONArrayLen(t, firstBatch.Body.String(), "created", 1)
	requireJSONArrayLen(t, firstBatch.Body.String(), "conflicts", 0)

	secondBatch := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/students/batch", map[string]any{
		"students": []map[string]any{
			{
				"username":        "student-1",
				"displayName":     "小明",
				"initialPassword": "abc12345",
			},
			{
				"username":        "student-2",
				"displayName":     "小红",
				"initialPassword": "xyz98765",
			},
		},
	})

	require.Equal(t, http.StatusCreated, secondBatch.Code)
	requireJSONArrayLen(t, secondBatch.Body.String(), "created", 1)
	requireJSONArrayLen(t, secondBatch.Body.String(), "conflicts", 1)

	listRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/students", nil)
	require.Equal(t, http.StatusOK, listRes.Code)
	requireJSONArrayLen(t, listRes.Body.String(), "items", 2)
}
