package tests

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/sb3"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
	_ "modernc.org/sqlite"
)

func TestTeacherAndStudentCanReadMeAndLogout(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-me", "secret123")

	teacherMeRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/me", nil)
	require.Equal(t, http.StatusOK, teacherMeRes.Code)
	requireBodyField(t, teacherMeRes.Body.String(), "teacherName", "teacher-me")

	teacherLogoutRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/logout", nil)
	require.Equal(t, http.StatusOK, teacherLogoutRes.Code)

	teacherMeAfterLogoutRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/me", nil)
	require.Equal(t, http.StatusUnauthorized, teacherMeAfterLogoutRes.Code)

	teacherToken = registerTeacher(t, handler, "teacher-student-me", "secret123")
	createStudent(t, handler, teacherToken, "student-me", "小灰", "abc12345")
	studentToken := loginStudent(t, handler, "student-me", "abc12345")

	studentMeRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodGet, "/api/student/me", nil)
	require.Equal(t, http.StatusOK, studentMeRes.Code)
	requireBodyField(t, studentMeRes.Body.String(), "studentName", "小灰")

	studentLogoutRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, "/api/student/logout", nil)
	require.Equal(t, http.StatusOK, studentLogoutRes.Code)

	studentMeAfterLogoutRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodGet, "/api/student/me", nil)
	require.Equal(t, http.StatusUnauthorized, studentMeAfterLogoutRes.Code)
}

func TestTeacherCanResetStudentPassword(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-reset-password", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-reset-password", "小紫", "old-pass-1")

	resetRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/students/%d/reset-password", studentID), map[string]any{
		"newPassword": "new-pass-2",
	})
	require.Equal(t, http.StatusOK, resetRes.Code)

	oldLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-reset-password",
		"password":   "old-pass-1",
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusUnauthorized, oldLoginRes.Code)

	newLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-reset-password",
		"password":   "new-pass-2",
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusOK, newLoginRes.Code)
}

func TestTeacherAndStudentCanReadAssignmentDetailsAndTeacherCanArchive(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-assignment-detail", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-assignment-detail", "小黄", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "Detail Maze")

	assignStudentAndPublish(t, handler, teacherToken, assignmentID, studentID)
	studentToken := loginStudent(t, handler, "student-assignment-detail", "abc12345")
	reportStudentProgress(t, handler, studentToken, assignmentID)

	hintRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintRes.Code)

	teacherDetailRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/assignments/%d", assignmentID), nil)
	require.Equal(t, http.StatusOK, teacherDetailRes.Code)
	requireBodyField(t, teacherDetailRes.Body.String(), "title", "Detail Maze")
	requireJSONArrayLen(t, teacherDetailRes.Body.String(), "assignedStudents", 1)

	studentDetailRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodGet, fmt.Sprintf("/api/student/assignments/%d", assignmentID), nil)
	require.Equal(t, http.StatusOK, studentDetailRes.Code)
	requireBodyField(t, studentDetailRes.Body.String(), "title", "Detail Maze")

	studentDetailRecord := parseObject(t, studentDetailRes.Body.String())
	latestProgress, ok := studentDetailRecord["latestProgress"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "让 Cat 角色移动起来", latestProgress["currentTarget"])

	latestHint, ok := studentDetailRecord["latestHint"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, latestHint["hintText"])

	archiveRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/archive", assignmentID), nil)
	require.Equal(t, http.StatusOK, archiveRes.Code)
	requireBodyField(t, archiveRes.Body.String(), "status", "archived")

	assignmentsRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodGet, "/api/student/assignments", nil)
	require.Equal(t, http.StatusOK, assignmentsRes.Code)
	requireJSONArrayLen(t, assignmentsRes.Body.String(), "items", 0)
}

func TestTeacherCanReadStudentHistory(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-history", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-history", "小橙", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "History Maze")

	assignStudentAndPublish(t, handler, teacherToken, assignmentID, studentID)
	studentToken := loginStudent(t, handler, "student-history", "abc12345")
	reportStudentProgress(t, handler, studentToken, assignmentID)

	hintRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintRes.Code)

	historyRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/dashboard/students/%d/history", studentID), nil)
	require.Equal(t, http.StatusOK, historyRes.Code)
	requireBodyField(t, historyRes.Body.String(), "studentName", "小橙")
	requireJSONArrayLen(t, historyRes.Body.String(), "items", 1)

	historyRecord := parseObject(t, historyRes.Body.String())
	items := historyRecord["items"].([]any)
	firstItem := items[0].(map[string]any)
	require.Equal(t, "History Maze", firstItem["assignmentTitle"])
	require.NotEmpty(t, firstItem["hintText"])
}

func TestStudentCanReadPendingAssignmentDetailButCannotRequestHint(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SERVER_API_DB_PATH", filepath.Join(t.TempDir(), "server-api.sqlite3"))
	t.Setenv("SB3_STORAGE_DIR", filepath.Join(t.TempDir(), "sb3"))

	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-pending-detail", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-pending-detail", "小候", "abc12345")
	studentToken := loginStudent(t, handler, "student-pending-detail", "abc12345")

	cfg := config.FromEnv()
	store, err := memory.NewStore(cfg)
	require.NoError(t, err)

	teacher, ok := store.FindTeacherByUsername("teacher-pending-detail")
	require.True(t, ok)

	storage := sb3.NewLocalStorage(cfg.SB3StorageDir)
	sb3Path, err := storage.Save(t.Context(), "pending-detail.sb3", createSampleSB3(t))
	require.NoError(t, err)

	assignment := store.CreateAssignment(teacher.ID, memory.CreateAssignmentInput{
		Title:       "Pending Detail Maze",
		Goal:        "查看未完成分析状态",
		Description: "测试任务",
		FileName:    "pending-detail.sb3",
		SB3FilePath: sb3Path,
		SB3Data:     createSampleSB3(t),
	})

	require.NoError(t, store.AssignStudents(teacher.ID, assignment.ID, []int64{studentID}))

	db, err := sql.Open("sqlite", cfg.DatabasePath)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("UPDATE assignments SET status = ? WHERE id = ?", "published", assignment.ID)
	require.NoError(t, err)

	detailRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodGet, fmt.Sprintf("/api/student/assignments/%d", assignment.ID), nil)
	require.Equal(t, http.StatusOK, detailRes.Code)
	requireBodyField(t, detailRes.Body.String(), "analysisStatus", "pending")

	hintRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignment.ID), nil)
	require.Equal(t, http.StatusConflict, hintRes.Code)
}

func TestTeacherCannotResetOtherTeachersStudentPassword(t *testing.T) {
	handler := newTestHandler()
	teacherTokenA := registerTeacher(t, handler, "teacher-password-owner", "secret123")
	teacherTokenB := registerTeacher(t, handler, "teacher-password-other", "secret123")
	studentID := createStudent(t, handler, teacherTokenA, "student-other-teacher", "小隔离", "abc12345")

	resetRes := performAuthedJSONRequest(t, handler, teacherTokenB, http.MethodPost, fmt.Sprintf("/api/teacher/students/%d/reset-password", studentID), map[string]any{
		"newPassword": "new-pass-2",
	})
	require.Equal(t, http.StatusNotFound, resetRes.Code)
}
