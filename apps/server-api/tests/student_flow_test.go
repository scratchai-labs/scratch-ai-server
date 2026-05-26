package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStudentCanLoginAndReadAssignedAssignments(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-student-login", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-login", "小明", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "Flight Game")

	assignRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/assign-students", assignmentID), map[string]any{
		"studentIds": []int64{studentID},
	})
	require.Equal(t, http.StatusOK, assignRes.Code)

	publishRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/publish", assignmentID), nil)
	require.Equal(t, http.StatusOK, publishRes.Code)

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-login",
		"password":   "abc12345",
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusOK, loginRes.Code)
	requireBodyField(t, loginRes.Body.String(), "studentName", "小明")
	studentToken := requireStringField(t, loginRes.Body.String(), "token")

	assignmentsRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodGet, "/api/student/assignments", nil)
	require.Equal(t, http.StatusOK, assignmentsRes.Code)
	requireJSONArrayLen(t, assignmentsRes.Body.String(), "items", 1)
}

func TestStudentLoginRejectsNonDesktopClient(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-client-type", "secret123")
	createStudent(t, handler, teacherToken, "student-client-type", "小端", "abc12345")

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-client-type",
		"password":   "abc12345",
		"clientType": "web",
	})

	require.Equal(t, http.StatusBadRequest, loginRes.Code)
}

func TestStudentCannotReadUnassignedAssignment(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-unassigned", "secret123")
	createStudent(t, handler, teacherToken, "student-unassigned", "小未", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "Unassigned Maze")

	publishRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/publish", assignmentID), nil)
	require.Equal(t, http.StatusOK, publishRes.Code)

	studentToken := loginStudent(t, handler, "student-unassigned", "abc12345")

	detailRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodGet, fmt.Sprintf("/api/student/assignments/%d", assignmentID), nil)
	require.Equal(t, http.StatusNotFound, detailRes.Code)
}

func TestStudentCanReportProgressRequestHintAndTeacherSeeLiveDashboard(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-progress", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-progress", "小红", "xyz98765")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "Maze Progress Game")

	assignRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/assign-students", assignmentID), map[string]any{
		"studentIds": []int64{studentID},
	})
	require.Equal(t, http.StatusOK, assignRes.Code)

	publishRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, fmt.Sprintf("/api/teacher/assignments/%d/publish", assignmentID), nil)
	require.Equal(t, http.StatusOK, publishRes.Code)

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-progress",
		"password":   "xyz98765",
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusOK, loginRes.Code)
	studentToken := requireStringField(t, loginRes.Body.String(), "token")

	progressRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/progress", assignmentID), map[string]any{
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

	hintRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusCreated, hintRes.Code)
	requireBodyField(t, hintRes.Body.String(), "providerName", "fallback")
	require.Contains(t, requireStringField(t, hintRes.Body.String(), "hintText"), "Cat")

	dashboardRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, fmt.Sprintf("/api/teacher/dashboard/assignments/%d/live", assignmentID), nil)
	require.Equal(t, http.StatusOK, dashboardRes.Code)
	requireJSONArrayLen(t, dashboardRes.Body.String(), "students", 1)
}

func TestStudentCannotRequestHintWithoutProgress(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-hint-progress", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-hint-progress", "小等", "abc12345")
	assignmentID := uploadAssignmentAndWaitReady(t, handler, teacherToken, "Hint Progress Maze")

	assignStudentAndPublish(t, handler, teacherToken, assignmentID, studentID)
	studentToken := loginStudent(t, handler, "student-hint-progress", "abc12345")

	hintRes := performAuthedJSONRequest(t, handler, studentToken, http.MethodPost, fmt.Sprintf("/api/student/assignments/%d/hints", assignmentID), nil)
	require.Equal(t, http.StatusConflict, hintRes.Code)
}
