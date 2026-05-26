package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTeacherCanUploadAssignmentSB3(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-assignment", "secret123")

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Space Game",
		"goal":        "让角色移动起来",
		"description": "第一阶段任务",
	}, "sb3", "space-game.sb3", createSampleSB3(t))

	require.Equal(t, http.StatusCreated, res.Code)
	requireBodyField(t, res.Body.String(), "title", "Space Game")
	requireBodyField(t, res.Body.String(), "analysisStatus", "pending")
	require.NotZero(t, requireInt64Field(t, res.Body.String(), "id"))
}

func TestAssignmentAnalysisTransitionsToReady(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-analysis", "secret123")

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Maze Game",
		"goal":        "让角色按事件响应",
		"description": "第二阶段任务",
	}, "sb3", "maze-game.sb3", createSampleSB3(t))

	require.Equal(t, http.StatusCreated, res.Code)
	assignmentID := requireInt64Field(t, res.Body.String(), "id")

	require.Eventually(t, func() bool {
		analysisRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/assignments/%d/analysis", assignmentID), nil)
		if analysisRes.Code != http.StatusOK {
			return false
		}

		record := parseObject(t, analysisRes.Body.String())
		status, ok := record["analysisStatus"].(string)
		if !ok || status != "ready" {
			return false
		}

		roles, ok := record["roleNames"].([]any)
		return ok && len(roles) == 2
	}, 2*time.Second, 20*time.Millisecond)
}

func TestTeacherCanListAssignments(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-list-assignments", "secret123")

	uploadAssignmentAndWaitReady(t, handler, teacherToken, "Loop Practice")

	listRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/assignments", nil)
	require.Equal(t, http.StatusOK, listRes.Code)
	requireJSONArrayLen(t, listRes.Body.String(), "items", 1)
}

func TestTeacherUploadRejectsNonSB3Extension(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-invalid-extension", "secret123")

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Bad Upload",
		"goal":        "测试非法扩展名",
		"description": "测试任务",
	}, "sb3", "bad-upload.txt", createSampleSB3(t))

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestTeacherUploadRejectsUnsupportedSB3MIMEType(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-invalid-mime", "secret123")

	res := performMultipartAuthedRequestWithFileContentType(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Bad MIME",
		"goal":        "测试非法 MIME",
		"description": "测试任务",
	}, "sb3", "bad-mime.sb3", "image/png", createSampleSB3(t))

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestTeacherUploadRejectsOversizedSB3(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-oversized-sb3", "secret123")

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Too Large",
		"goal":        "测试超大文件",
		"description": "测试任务",
	}, "sb3", "too-large.sb3", bytes.Repeat([]byte("a"), 17<<20))

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestAssignmentInvalidZIPTransitionsToFailedAndSB3IsStored(t *testing.T) {
	storageDir := t.TempDir()
	t.Setenv("SB3_STORAGE_DIR", storageDir)

	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-invalid-zip", "secret123")

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Broken ZIP",
		"goal":        "测试错误路径",
		"description": "损坏文件",
	}, "sb3", "broken.sb3", []byte("not-a-zip"))

	require.Equal(t, http.StatusCreated, res.Code)
	assignmentID := requireInt64Field(t, res.Body.String(), "id")

	require.Eventually(t, func() bool {
		analysisRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/assignments/%d/analysis", assignmentID), nil)
		if analysisRes.Code != http.StatusOK {
			return false
		}

		record := parseObject(t, analysisRes.Body.String())
		status, ok := record["analysisStatus"].(string)
		return ok && status == "failed"
	}, 2*time.Second, 20*time.Millisecond)

	entries, err := os.ReadDir(storageDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestTeacherCannotPublishAssignmentWhenAnalysisIsNotReady(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-publish-not-ready", "secret123")

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Broken Publish",
		"goal":        "测试未就绪发布",
		"description": "损坏文件",
	}, "sb3", "broken-publish.sb3", []byte("not-a-zip"))

	require.Equal(t, http.StatusCreated, res.Code)
	assignmentID := requireInt64Field(t, res.Body.String(), "id")

	require.Eventually(t, func() bool {
		analysisRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/assignments/%d/analysis", assignmentID), nil)
		if analysisRes.Code != http.StatusOK {
			return false
		}

		record := parseObject(t, analysisRes.Body.String())
		status, ok := record["analysisStatus"].(string)
		return ok && status == "failed"
	}, 2*time.Second, 20*time.Millisecond)

	publishRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/publish", assignmentID), nil)
	require.Equal(t, http.StatusConflict, publishRes.Code)
}

func TestAssignmentAnalysisIncludesBroadcastsVariablesListsAndExtensions(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-rich-analysis", "secret123")

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       "Rich Analysis",
		"goal":        "测试丰富分析",
		"description": "测试任务",
	}, "sb3", "rich-analysis.sb3", createAdvancedSB3(t))

	require.Equal(t, http.StatusCreated, res.Code)
	assignmentID := requireInt64Field(t, res.Body.String(), "id")

	require.Eventually(t, func() bool {
		analysisRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/assignments/%d/analysis", assignmentID), nil)
		if analysisRes.Code != http.StatusOK {
			return false
		}

		record := parseObject(t, analysisRes.Body.String())
		status, ok := record["analysisStatus"].(string)
		if !ok || status != "ready" {
			return false
		}

		scriptCounts, ok := record["scriptCounts"].(map[string]any)
		require.True(t, ok, "field scriptCounts should be an object")
		require.Equal(t, float64(1), scriptCounts["Stage"])
		require.Equal(t, float64(2), scriptCounts["Cat"])
		requireJSONArrayLen(t, analysisRes.Body.String(), "broadcastMessages", 1)
		requireJSONArrayLen(t, analysisRes.Body.String(), "variableNames", 1)
		requireJSONArrayLen(t, analysisRes.Body.String(), "listNames", 1)
		requireJSONArrayLen(t, analysisRes.Body.String(), "extensions", 1)
		return true
	}, 2*time.Second, 20*time.Millisecond)
}
