package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestTeacherLiveDashboardShowsDifferentHintsForDifferentStudents(t *testing.T) {
	deepSeekStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

		var requestPayload map[string]any
		defer r.Body.Close()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&requestPayload))

		messages, ok := requestPayload["messages"].([]any)
		require.True(t, ok)
		require.Len(t, messages, 2)

		userMessage, ok := messages[1].(map[string]any)
		require.True(t, ok)
		userContent, ok := userMessage["content"].(string)
		require.True(t, ok)

		hintText := "先把当前角色的事件和动作连成一条最小脚本。"
		switch {
		case strings.Contains(userContent, "让 Cat 角色移动起来"):
			hintText = "先把 Cat 的移动脚本补成完整一条，再测试碰到边缘后的表现。"
		case strings.Contains(userContent, "先让 Sprite1 说一句话"):
			hintText = "先给 Sprite1 接上“说 2 秒”，再按绿旗确认流程跑通。"
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{
			"choices": [
				{
					"message": {
						"content": %q
					}
				}
			]
		}`, hintText)))
	}))
	defer deepSeekStub.Close()

	t.Setenv("DEEPSEEK_BASE_URL", deepSeekStub.URL)
	t.Setenv("DEEPSEEK_API_KEY", "test-api-key")
	t.Setenv("DEEPSEEK_MODEL", "deepseek-v4-flash")

	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-deepseek-multi-student", "secret123")
	studentAID := createStudent(t, handler, teacherToken, "student-deepseek-a", "小青", "abc12345")
	studentBID := createStudent(t, handler, teacherToken, "student-deepseek-b", "小蓝", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "DeepSeek Multi Student Maze")

	assignRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/assign-students", assignmentID), map[string]any{
		"studentIds": []int64{studentAID, studentBID},
	})
	require.Equal(t, http.StatusOK, assignRes.Code)

	publishRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/publish", assignmentID), nil)
	require.Equal(t, http.StatusOK, publishRes.Code)

	studentAToken := loginStudent(t, handler, "student-deepseek-a", "abc12345")
	studentBToken := loginStudent(t, handler, "student-deepseek-b", "abc12345")

	reportStudentProgress(t, handler, studentAToken, assignmentID)
	progressResB := performAuthedJSONRequest(t, handler, studentBToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/progress", assignmentID), map[string]any{
		"currentTarget":    "先让 Sprite1 说一句话",
		"stepSummary":      "已经放上开始事件，但还没接外观积木",
		"localProjectHash": "hash-student-b",
		"reportedAt":       "2026-05-25T10:05:00Z",
		"snapshot": map[string]any{
			"currentRoleName": "Sprite1",
			"roles": []map[string]any{
				{
					"roleName": "Stage",
					"roleType": "stage",
					"blocks":   []string{"当绿旗被点击"},
				},
				{
					"roleName": "Sprite1",
					"roleType": "sprite",
					"blocks":   []string{"当绿旗被点击"},
				},
			},
		},
	})
	require.Equal(t, http.StatusCreated, progressResB.Code)

	hintResA := performAuthedJSONRequest(t, handler, studentAToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintResA.Code)
	requireBodyField(t, hintResA.Body.String(), "providerName", "deepseek")
	requireBodyField(t, hintResA.Body.String(), "hintText", "先把 Cat 的移动脚本补成完整一条，再测试碰到边缘后的表现。")

	hintResB := performAuthedJSONRequest(t, handler, studentBToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintResB.Code)
	requireBodyField(t, hintResB.Body.String(), "providerName", "deepseek")
	requireBodyField(t, hintResB.Body.String(), "hintText", "先给 Sprite1 接上“说 2 秒”，再按绿旗确认流程跑通。")

	dashboardRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/dashboard/assignments/%d/live", assignmentID), nil)
	require.Equal(t, http.StatusOK, dashboardRes.Code)
	requireJSONArrayLen(t, dashboardRes.Body.String(), "students", 2)

	record := parseObject(t, dashboardRes.Body.String())
	students, ok := record["students"].([]any)
	require.True(t, ok)

	studentByName := map[string]map[string]any{}
	for _, rawStudent := range students {
		studentRecord, ok := rawStudent.(map[string]any)
		require.True(t, ok)
		studentName, ok := studentRecord["studentName"].(string)
		require.True(t, ok)
		studentByName[studentName] = studentRecord
	}

	require.Equal(t, "让 Cat 角色移动起来", studentByName["小青"]["currentTarget"])
	require.Equal(t, "先把 Cat 的移动脚本补成完整一条，再测试碰到边缘后的表现。", studentByName["小青"]["lastHintText"])
	require.Equal(t, "先让 Sprite1 说一句话", studentByName["小蓝"]["currentTarget"])
	require.Equal(t, "先给 Sprite1 接上“说 2 秒”，再按绿旗确认流程跑通。", studentByName["小蓝"]["lastHintText"])
}
