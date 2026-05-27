package tests

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/app"
)

func newTestHandler() http.Handler {
	return app.New()
}

func performJSONRequest(t *testing.T, handler http.Handler, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func performAuthedJSONRequest(t *testing.T, handler http.Handler, token string, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func performMultipartAuthedRequest(t *testing.T, handler http.Handler, token string, method string, path string, fields map[string]string, fileField string, fileName string, fileBody []byte) *httptest.ResponseRecorder {
	t.Helper()

	return performMultipartAuthedRequestWithFileContentType(t, handler, token, method, path, fields, fileField, fileName, "application/octet-stream", fileBody)
}

func performMultipartAuthedRequestWithFileContentType(t *testing.T, handler http.Handler, token string, method string, path string, fields map[string]string, fileField string, fileName string, fileContentType string, fileBody []byte) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		require.NoError(t, writer.WriteField(key, value))
	}

	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     fileField,
		"filename": fileName,
	}))
	fileHeader.Set("Content-Type", fileContentType)

	fileWriter, err := writer.CreatePart(fileHeader)
	require.NoError(t, err)
	_, err = fileWriter.Write(fileBody)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(method, path, &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func registerTeacher(t *testing.T, handler http.Handler, username string, password string) string {
	t.Helper()

	res := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/register", map[string]any{
		"username": username,
		"password": password,
	})
	require.Equal(t, http.StatusCreated, res.Code)
	return requireStringField(t, res.Body.String(), "token")
}

func requireStringField(t *testing.T, raw string, field string) string {
	t.Helper()

	value, ok := lookupField(t, raw, field).(string)
	require.True(t, ok, "field %s should be a string", field)
	return value
}

func requireBodyField(t *testing.T, raw string, field string, expected string) {
	t.Helper()
	require.Equal(t, expected, requireStringField(t, raw, field))
}

func requireJSONArrayLen(t *testing.T, raw string, field string, expected int) {
	t.Helper()

	items, ok := lookupField(t, raw, field).([]any)
	require.True(t, ok, "field %s should be an array", field)
	require.Len(t, items, expected)
}

func requireInt64Field(t *testing.T, raw string, field string) int64 {
	t.Helper()

	value, ok := lookupField(t, raw, field).(float64)
	require.True(t, ok, "field %s should be a number", field)
	return int64(value)
}

func parseObject(t *testing.T, raw string) map[string]any {
	t.Helper()

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &decoded))
	return decoded
}

func lookupField(t *testing.T, raw string, field string) any {
	t.Helper()

	var current any = parseObject(t, raw)
	for _, segment := range strings.Split(field, ".") {
		if index, err := strconv.Atoi(segment); err == nil {
			items, ok := current.([]any)
			require.True(t, ok, "field %s should contain an array before index %s", field, segment)
			require.Greater(t, len(items), index, "field %s index %d out of range", field, index)
			current = items[index]
			continue
		}

		record, ok := current.(map[string]any)
		require.True(t, ok, "field %s should contain an object before key %s", field, segment)
		value, exists := record[segment]
		require.True(t, exists, "field %s should include key %s", field, segment)
		current = value
	}

	return current
}

func createSampleSB3(t *testing.T) []byte {
	t.Helper()

	var body bytes.Buffer
	writer := zip.NewWriter(&body)
	projectFile, err := writer.Create("project.json")
	require.NoError(t, err)

	projectJSON := `{
  "targets": [
    {
      "isStage": true,
      "name": "Stage",
      "blocks": {
        "stage-1": { "opcode": "event_whenflagclicked", "next": "stage-2", "topLevel": true },
        "stage-2": { "opcode": "event_broadcast", "next": null, "parent": "stage-1" }
      }
    },
    {
      "isStage": false,
      "name": "Cat",
      "blocks": {
        "cat-1": { "opcode": "event_whenbroadcastreceived", "next": "cat-2", "topLevel": true },
        "cat-2": { "opcode": "motion_movesteps", "next": "cat-3", "parent": "cat-1" },
        "cat-3": { "opcode": "motion_ifonedgebounce", "next": null, "parent": "cat-2" }
      }
    }
  ],
  "extensions": []
}`

	_, err = io.WriteString(projectFile, projectJSON)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return body.Bytes()
}

func createAdvancedSB3(t *testing.T) []byte {
	t.Helper()

	var body bytes.Buffer
	writer := zip.NewWriter(&body)
	projectFile, err := writer.Create("project.json")
	require.NoError(t, err)

	projectJSON := `{
  "targets": [
    {
      "isStage": true,
      "name": "Stage",
      "broadcasts": {
        "message1": "开始"
      },
      "variables": {
        "score": ["分数", 0]
      },
      "lists": {
        "todo": ["步骤列表", []]
      },
      "blocks": {
        "stage-1": { "opcode": "event_whenflagclicked", "next": "stage-2", "topLevel": true },
        "stage-2": { "opcode": "event_broadcast", "next": null, "parent": "stage-1" }
      }
    },
    {
      "isStage": false,
      "name": "Cat",
      "blocks": {
        "cat-1": { "opcode": "event_whenbroadcastreceived", "next": "cat-2", "topLevel": true },
        "cat-2": { "opcode": "pen_clear", "next": null, "parent": "cat-1" },
        "cat-3": { "opcode": "looks_say", "next": null, "topLevel": true }
      }
    }
  ],
  "extensions": ["pen"]
}`

	_, err = io.WriteString(projectFile, projectJSON)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return body.Bytes()
}

func createStudent(t *testing.T, handler http.Handler, teacherToken string, username string, displayName string, password string) int64 {
	t.Helper()

	res := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/students/batch", map[string]any{
		"students": []map[string]any{
			{
				"username":        username,
				"displayName":     displayName,
				"initialPassword": password,
			},
		},
	})
	require.Equal(t, http.StatusCreated, res.Code)

	record := parseObject(t, res.Body.String())
	created, ok := record["created"].([]any)
	require.True(t, ok)
	require.Len(t, created, 1)

	studentRecord, ok := created[0].(map[string]any)
	require.True(t, ok)
	return int64(studentRecord["id"].(float64))
}

func uploadAssignmentAndWaitReady(t *testing.T, handler http.Handler, teacherToken string, title string) int64 {
	t.Helper()

	res := performMultipartAuthedRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments", map[string]string{
		"title":       title,
		"goal":        "先完成作品主流程",
		"description": "测试任务",
	}, "sb3", "fixture.sb3", createSampleSB3(t))
	require.Equal(t, http.StatusCreated, res.Code)
	assignmentID := requireInt64Field(t, res.Body.String(), "id")

	require.Eventually(t, func() bool {
		analysisRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/assignments/"+strconv.FormatInt(assignmentID, 10)+"/analysis", nil)
		if analysisRes.Code != http.StatusOK {
			return false
		}

		record := parseObject(t, analysisRes.Body.String())
		status, ok := record["analysisStatus"].(string)
		return ok && status == "ready"
	}, 2*time.Second, 20*time.Millisecond)

	return assignmentID
}

func assignStudentAndPublish(t *testing.T, handler http.Handler, teacherToken string, assignmentID int64, studentID int64) {
	t.Helper()

	assignRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments/"+strconv.FormatInt(assignmentID, 10)+"/assign-students", map[string]any{
		"studentIds": []int64{studentID},
	})
	require.Equal(t, http.StatusOK, assignRes.Code)

	publishRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/assignments/"+strconv.FormatInt(assignmentID, 10)+"/publish", nil)
	require.Equal(t, http.StatusOK, publishRes.Code)
}

func loginStudent(t *testing.T, handler http.Handler, username string, password string) string {
	t.Helper()

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   username,
		"password":   password,
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusOK, loginRes.Code)
	return requireStringField(t, loginRes.Body.String(), "token")
}

func reportStudentProgress(t *testing.T, handler http.Handler, studentToken string, assignmentID int64) {
	t.Helper()

	progressRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, "/api/student/assignments/"+strconv.FormatInt(assignmentID, 10)+"/progress", map[string]any{
		"currentTarget":    "让 Cat 角色移动起来",
		"stepSummary":      "已经把事件积木接上了",
		"localProjectHash": "hash-1",
		"reportedAt":       "2026-05-25T10:00:00Z",
		"snapshot": map[string]any{
			"currentRoleName": "Cat",
			"roles": []map[string]any{
				{
					"roleName": "Stage",
					"roleType": "stage",
					"blocks":   []string{"当绿旗被点击"},
				},
				{
					"roleName": "Cat",
					"roleType": "sprite",
					"blocks":   []string{"当接收到 开始", "移动 10 步"},
				},
			},
		},
	})
	require.Equal(t, http.StatusCreated, progressRes.Code)
}
