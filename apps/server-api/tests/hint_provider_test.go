package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStudentHintUsesDeepSeekWhenConfigured(t *testing.T) {
	var capturedRequest map[string]any

	deepSeekStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&capturedRequest))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "先把 Cat 的事件积木和移动积木连成一条完整脚本。"
					}
				}
			]
		}`))
	}))
	defer deepSeekStub.Close()

	t.Setenv("DEEPSEEK_BASE_URL", deepSeekStub.URL)
	t.Setenv("DEEPSEEK_API_KEY", "test-api-key")
	t.Setenv("DEEPSEEK_MODEL", "deepseek-v4-flash")

	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-deepseek-success", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-deepseek-success", "小青", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "DeepSeek Maze")

	assignStudentAndPublish(t, handler, teacherToken, assignmentID, studentID)
	studentToken := loginStudent(t, handler, "student-deepseek-success", "abc12345")
	reportStudentProgress(t, handler, studentToken, assignmentID)

	hintRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintRes.Code)
	requireBodyField(t, hintRes.Body.String(), "providerName", "deepseek")
	require.Contains(t, requireStringField(t, hintRes.Body.String(), "hintText"), "Cat")

	require.Equal(t, "deepseek-v4-flash", capturedRequest["model"])
	messages, ok := capturedRequest["messages"].([]any)
	require.True(t, ok)
	require.Len(t, messages, 2)

	systemMessage := messages[0].(map[string]any)
	userMessage := messages[1].(map[string]any)
	require.Equal(t, "system", systemMessage["role"])
	require.Equal(t, "user", userMessage["role"])
	require.Contains(t, userMessage["content"], "DeepSeek Maze")
	require.Contains(t, userMessage["content"], "Cat")
	require.Contains(t, userMessage["content"], "移动 10 步")
}

func TestStudentHintFallsBackWhenDeepSeekFails(t *testing.T) {
	deepSeekStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream down", http.StatusBadGateway)
	}))
	defer deepSeekStub.Close()

	t.Setenv("DEEPSEEK_BASE_URL", deepSeekStub.URL)
	t.Setenv("DEEPSEEK_API_KEY", "test-api-key")
	t.Setenv("DEEPSEEK_MODEL", "deepseek-v4-flash")

	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-deepseek-fallback", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-deepseek-fallback", "小蓝", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "Fallback Maze")

	assignStudentAndPublish(t, handler, teacherToken, assignmentID, studentID)
	studentToken := loginStudent(t, handler, "student-deepseek-fallback", "abc12345")
	reportStudentProgress(t, handler, studentToken, assignmentID)

	hintRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintRes.Code)
	requireBodyField(t, hintRes.Body.String(), "providerName", "fallback")
	require.Contains(t, requireStringField(t, hintRes.Body.String(), "hintText"), "Cat")
}
