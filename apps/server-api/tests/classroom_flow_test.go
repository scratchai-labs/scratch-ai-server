package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeacherCanCreateListAndDeleteClassroom(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-classroom", "secret123")

	createRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/classes", map[string]any{
		"name": "四年级一班",
	})
	require.Equal(t, http.StatusCreated, createRes.Code)
	classroomID := requireInt64Field(t, createRes.Body.String(), "id")

	listRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/classes", nil)
	require.Equal(t, http.StatusOK, listRes.Code)
	requireJSONArrayLen(t, listRes.Body.String(), "items", 1)
	requireBodyField(t, listRes.Body.String(), "items.0.name", "四年级一班")

	deleteRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodDelete, "/api/teacher/classes/"+itoa(classroomID), nil)
	require.Equal(t, http.StatusOK, deleteRes.Code)
}

func TestTeacherCannotDeleteNonEmptyClassroom(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-classroom-non-empty", "secret123")
	classroomID := createClassroom(t, handler, teacherToken, "四年级二班")
	createStudentInClassroom(t, handler, teacherToken, classroomID, "student-classroom", "小班", "abc12345")

	deleteRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodDelete, "/api/teacher/classes/"+itoa(classroomID), nil)
	require.Equal(t, http.StatusConflict, deleteRes.Code)
}

func TestTeacherCanCreateStudentsAndProjectsInsideClassroom(t *testing.T) {
	handler := newTestHandler()
	teacherToken := registerTeacher(t, handler, "teacher-classroom-scope", "secret123")
	classroomID := createClassroom(t, handler, teacherToken, "五年级一班")

	studentRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/classes/"+itoa(classroomID)+"/students/batch", map[string]any{
		"students": []map[string]any{
			{
				"username":        "student-in-class",
				"displayName":     "小五",
				"initialPassword": "abc12345",
			},
		},
	})
	require.Equal(t, http.StatusCreated, studentRes.Code)
	studentID := requireInt64Field(t, studentRes.Body.String(), "created.0.id")

	projectID := uploadAssignmentToClassroomAndWaitReady(t, handler, teacherToken, classroomID, "班级项目")
	assignStudentAndPublish(t, handler, teacherToken, projectID, studentID)

	projectListRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/classes/"+itoa(classroomID)+"/projects", nil)
	require.Equal(t, http.StatusOK, projectListRes.Code)
	requireJSONArrayLen(t, projectListRes.Body.String(), "items", 1)

	detailRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/projects/"+itoa(projectID), nil)
	require.Equal(t, http.StatusOK, detailRes.Code)

	analysisRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/projects/"+itoa(projectID)+"/analysis", nil)
	require.Equal(t, http.StatusOK, analysisRes.Code)

	liveRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/teacher/projects/"+itoa(projectID)+"/live", nil)
	require.Equal(t, http.StatusOK, liveRes.Code)

	assignAliasRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/projects/"+itoa(projectID)+"/assign-students", map[string]any{
		"studentIds": []int64{studentID},
	})
	require.Equal(t, http.StatusOK, assignAliasRes.Code)

	publishAliasRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/projects/"+itoa(projectID)+"/publish", nil)
	require.Equal(t, http.StatusOK, publishAliasRes.Code)

	archiveAliasRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/teacher/projects/"+itoa(projectID)+"/archive", nil)
	require.Equal(t, http.StatusOK, archiveAliasRes.Code)
}
