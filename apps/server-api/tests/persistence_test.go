package tests

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/sb3"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

func TestServerDataPersistsAcrossAppRestartsWithFileDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SERVER_API_DB_PATH", filepath.Join(t.TempDir(), "server-api.sqlite3"))

	firstHandler := newTestHandler()
	teacherToken := registerTeacher(t, firstHandler, "teacher-persist", "secret123")
	studentID := createStudent(t, firstHandler, teacherToken, "student-persist", "小青", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, firstHandler, teacherToken, "Persistent Maze")

	assignStudentAndPublish(t, firstHandler, teacherToken, assignmentID, studentID)
	studentToken := loginStudent(t, firstHandler, "student-persist", "abc12345")
	reportStudentProgress(t, firstHandler, studentToken, assignmentID)

	hintRes := performAuthedJSONRequest(t, firstHandler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintRes.Code)

	secondHandler := newTestHandler()
	teacherLoginRes := performJSONRequest(t, secondHandler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-persist",
		"password": "secret123",
	})
	require.Equal(t, http.StatusOK, teacherLoginRes.Code)
	teacherToken2 := requireStringField(t, teacherLoginRes.Body.String(), "token")

	studentToken2 := loginStudent(t, secondHandler, "student-persist", "abc12345")

	assignmentsRes := performAuthedJSONRequest(t, secondHandler, studentToken2, http.MethodGet, "/api/student/assignments", nil)
	require.Equal(t, http.StatusOK, assignmentsRes.Code)
	requireJSONArrayLen(t, assignmentsRes.Body.String(), "items", 1)

	historyRes := performAuthedJSONRequest(t, secondHandler, teacherToken2, http.MethodGet, fmt.Sprintf("/api/teacher/dashboard/students/%d/history", studentID), nil)
	require.Equal(t, http.StatusOK, historyRes.Code)
	requireJSONArrayLen(t, historyRes.Body.String(), "items", 1)
}

func TestPendingAssignmentAnalysisResumesOnAppStartup(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SERVER_API_DB_PATH", filepath.Join(t.TempDir(), "server-api.sqlite3"))
	t.Setenv("SB3_STORAGE_DIR", filepath.Join(t.TempDir(), "sb3"))

	cfg := config.FromEnv()
	store, err := memory.NewStore(cfg)
	require.NoError(t, err)

	storage := sb3.NewLocalStorage(cfg.SB3StorageDir)
	sb3Path, err := storage.Save(t.Context(), "resume.sb3", createAdvancedSB3(t))
	require.NoError(t, err)

	teacher, err := store.CreateTeacher("teacher-resume-analysis", "hash")
	require.NoError(t, err)

	const teacherToken = "resume-analysis-token"
	store.SaveTeacherToken(teacherToken, teacher.ID)

	assignment := store.CreateAssignment(teacher.ID, memory.CreateAssignmentInput{
		Title:       "Resume Analysis",
		Goal:        "重启后恢复分析",
		Description: "待分析任务",
		FileName:    "resume.sb3",
		SB3FilePath: sb3Path,
		SB3Data:     createAdvancedSB3(t),
	})

	handler := newTestHandler()

	require.Eventually(t, func() bool {
		analysisRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/assignments/%d/analysis", assignment.ID), nil)
		if analysisRes.Code != http.StatusOK {
			return false
		}

		record := parseObject(t, analysisRes.Body.String())
		status, ok := record["analysisStatus"].(string)
		if !ok || status != "ready" {
			return false
		}

		requireJSONArrayLen(t, analysisRes.Body.String(), "broadcastMessages", 1)
		return true
	}, 2*time.Second, 20*time.Millisecond)
}
