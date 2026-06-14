package tests

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBootstrapAdminCanLoginAndManageTeachers(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "admin",
		"password": "admin12345",
	})
	require.Equal(t, http.StatusOK, loginRes.Code)
	requireBodyField(t, loginRes.Body.String(), "teacherName", "admin")
	requireBodyField(t, loginRes.Body.String(), "role", "admin")
	adminToken := requireStringField(t, loginRes.Body.String(), "token")

	listRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodGet, "/api/admin/teachers", nil)
	require.Equal(t, http.StatusOK, listRes.Code)
	requireJSONArrayLen(t, listRes.Body.String(), "items", 1)
	requireBodyField(t, listRes.Body.String(), "items.0.username", "admin")
	requireBodyField(t, listRes.Body.String(), "items.0.role", "admin")
	requireBodyField(t, listRes.Body.String(), "items.0.status", "active")

	createRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers", map[string]any{
		"username":        "teacher-managed",
		"initialPassword": "secret123",
	})
	require.Equal(t, http.StatusCreated, createRes.Code)
	requireBodyField(t, createRes.Body.String(), "username", "teacher-managed")
	requireBodyField(t, createRes.Body.String(), "role", "teacher")
	requireBodyField(t, createRes.Body.String(), "status", "active")
}

func TestAdminCanResetDisableAndEnableTeacher(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()
	adminToken := loginTeacherToken(t, handler, "admin", "admin12345")

	createRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers", map[string]any{
		"username":        "teacher-reset",
		"initialPassword": "secret123",
	})
	require.Equal(t, http.StatusCreated, createRes.Code)
	teacherID := requireInt64Field(t, createRes.Body.String(), "id")

	oldLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-reset",
		"password": "secret123",
	})
	require.Equal(t, http.StatusOK, oldLoginRes.Code)
	requireBodyField(t, oldLoginRes.Body.String(), "role", "teacher")

	resetRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/reset-password", map[string]any{
		"newPassword": "updated123",
	})
	require.Equal(t, http.StatusOK, resetRes.Code)
	requireBodyField(t, resetRes.Body.String(), "status", "active")

	disabledRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/disable", nil)
	require.Equal(t, http.StatusOK, disabledRes.Code)
	requireBodyField(t, disabledRes.Body.String(), "status", "disabled")

	disabledLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-reset",
		"password": "updated123",
	})
	require.Equal(t, http.StatusForbidden, disabledLoginRes.Code)

	enabledRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/enable", nil)
	require.Equal(t, http.StatusOK, enabledRes.Code)
	requireBodyField(t, enabledRes.Body.String(), "status", "active")

	newLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": "teacher-reset",
		"password": "updated123",
	})
	require.Equal(t, http.StatusOK, newLoginRes.Code)
}

func TestAdminCanViewOverviewAndManageStudents(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()
	adminToken := loginTeacherToken(t, handler, "admin", "admin12345")
	teacherToken := registerTeacher(t, handler, "teacher-students-admin", "secret123")
	studentID := createStudent(t, handler, teacherToken, "student-admin-1", "小蓝", "stud1234")

	overviewRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodGet, "/api/admin/overview", nil)
	require.Equal(t, http.StatusOK, overviewRes.Code)
	require.EqualValues(t, 1, requireInt64Field(t, overviewRes.Body.String(), "adminCount"))
	require.EqualValues(t, 1, requireInt64Field(t, overviewRes.Body.String(), "teacherCount"))
	require.EqualValues(t, 1, requireInt64Field(t, overviewRes.Body.String(), "activeTeacherCount"))
	require.EqualValues(t, 0, requireInt64Field(t, overviewRes.Body.String(), "disabledTeacherCount"))
	require.EqualValues(t, 1, requireInt64Field(t, overviewRes.Body.String(), "studentCount"))
	require.EqualValues(t, 1, requireInt64Field(t, overviewRes.Body.String(), "activeStudentCount"))
	require.EqualValues(t, 0, requireInt64Field(t, overviewRes.Body.String(), "disabledStudentCount"))

	listRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodGet, "/api/admin/students", nil)
	require.Equal(t, http.StatusOK, listRes.Code)
	requireJSONArrayLen(t, listRes.Body.String(), "items", 1)
	requireBodyField(t, listRes.Body.String(), "items.0.username", "student-admin-1")
	requireBodyField(t, listRes.Body.String(), "items.0.displayName", "小蓝")
	requireBodyField(t, listRes.Body.String(), "items.0.teacherUsername", "teacher-students-admin")
	requireBodyField(t, listRes.Body.String(), "items.0.status", "active")

	resetRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/students/"+strconv.FormatInt(studentID, 10)+"/reset-password", map[string]any{
		"newPassword": "renewed123",
	})
	require.Equal(t, http.StatusOK, resetRes.Code)
	requireBodyField(t, resetRes.Body.String(), "status", "active")

	disableRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/students/"+strconv.FormatInt(studentID, 10)+"/disable", nil)
	require.Equal(t, http.StatusOK, disableRes.Code)
	requireBodyField(t, disableRes.Body.String(), "status", "disabled")

	disabledLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-admin-1",
		"password":   "renewed123",
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusForbidden, disabledLoginRes.Code)

	overviewAfterDisableRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodGet, "/api/admin/overview", nil)
	require.Equal(t, http.StatusOK, overviewAfterDisableRes.Code)
	require.EqualValues(t, 0, requireInt64Field(t, overviewAfterDisableRes.Body.String(), "activeStudentCount"))
	require.EqualValues(t, 1, requireInt64Field(t, overviewAfterDisableRes.Body.String(), "disabledStudentCount"))

	enableRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/students/"+strconv.FormatInt(studentID, 10)+"/enable", nil)
	require.Equal(t, http.StatusOK, enableRes.Code)
	requireBodyField(t, enableRes.Body.String(), "status", "active")

	enabledLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-admin-1",
		"password":   "renewed123",
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusOK, enabledLoginRes.Code)
}

func TestAdminCanCreateStudentForTeacher(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()
	adminToken := loginTeacherToken(t, handler, "admin", "admin12345")

	createTeacherRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers", map[string]any{
		"username":        "teacher-student-owner",
		"initialPassword": "secret123",
	})
	require.Equal(t, http.StatusCreated, createTeacherRes.Code)
	teacherID := requireInt64Field(t, createTeacherRes.Body.String(), "id")

	createStudentRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/students", map[string]any{
		"teacherId":       teacherID,
		"username":        "student-created-by-admin",
		"displayName":     "小绿",
		"initialPassword": "stud1234",
	})
	require.Equal(t, http.StatusCreated, createStudentRes.Code)
	requireBodyField(t, createStudentRes.Body.String(), "username", "student-created-by-admin")
	requireBodyField(t, createStudentRes.Body.String(), "displayName", "小绿")
	requireBodyField(t, createStudentRes.Body.String(), "teacherUsername", "teacher-student-owner")
	requireBodyField(t, createStudentRes.Body.String(), "status", "active")

	studentLoginRes := performJSONRequest(t, handler, http.MethodPost, "/api/student/login", map[string]any{
		"username":   "student-created-by-admin",
		"password":   "stud1234",
		"clientType": "desktop",
	})
	require.Equal(t, http.StatusOK, studentLoginRes.Code)
	requireBodyField(t, studentLoginRes.Body.String(), "studentName", "小绿")
}

func TestAdminCanPromoteAndDemoteTeacherRole(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()
	adminToken := loginTeacherToken(t, handler, "admin", "admin12345")

	createTeacherRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers", map[string]any{
		"username":        "teacher-promoted",
		"initialPassword": "secret123",
	})
	require.Equal(t, http.StatusCreated, createTeacherRes.Code)
	teacherID := requireInt64Field(t, createTeacherRes.Body.String(), "id")

	promoteRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/role", map[string]any{
		"role": "admin",
	})
	require.Equal(t, http.StatusOK, promoteRes.Code)
	requireBodyField(t, promoteRes.Body.String(), "role", "admin")

	promotedToken := loginTeacherToken(t, handler, "teacher-promoted", "secret123")
	promotedOverviewRes := performAuthedJSONRequest(t, handler, promotedToken, http.MethodGet, "/api/admin/overview", nil)
	require.Equal(t, http.StatusOK, promotedOverviewRes.Code)

	demoteRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/role", map[string]any{
		"role": "teacher",
	})
	require.Equal(t, http.StatusOK, demoteRes.Code)
	requireBodyField(t, demoteRes.Body.String(), "role", "teacher")

	demotedOverviewRes := performAuthedJSONRequest(t, handler, promotedToken, http.MethodGet, "/api/admin/overview", nil)
	require.Equal(t, http.StatusForbidden, demotedOverviewRes.Code)
}

func TestAdminAuditLogsCaptureSensitiveOperations(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()
	adminToken := loginTeacherToken(t, handler, "admin", "admin12345")

	createTeacherRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers", map[string]any{
		"username":        "teacher-audit",
		"initialPassword": "secret123",
	})
	require.Equal(t, http.StatusCreated, createTeacherRes.Code)
	teacherID := requireInt64Field(t, createTeacherRes.Body.String(), "id")

	createStudentRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/students", map[string]any{
		"teacherId":       teacherID,
		"username":        "student-audit",
		"displayName":     "小审计",
		"initialPassword": "stud1234",
	})
	require.Equal(t, http.StatusCreated, createStudentRes.Code)
	studentID := requireInt64Field(t, createStudentRes.Body.String(), "id")

	disableStudentRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/students/"+strconv.FormatInt(studentID, 10)+"/disable", nil)
	require.Equal(t, http.StatusOK, disableStudentRes.Code)

	promoteTeacherRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodPost, "/api/admin/teachers/"+strconv.FormatInt(teacherID, 10)+"/role", map[string]any{
		"role": "admin",
	})
	require.Equal(t, http.StatusOK, promoteTeacherRes.Code)

	logsRes := performAuthedJSONRequest(t, handler, adminToken, http.MethodGet, "/api/admin/audit-logs", nil)
	require.Equal(t, http.StatusOK, logsRes.Code)
	requireJSONArrayLen(t, logsRes.Body.String(), "items", 4)
	requireBodyField(t, logsRes.Body.String(), "items.0.actorUsername", "admin")
	requireBodyField(t, logsRes.Body.String(), "items.0.action", "teacher.role_change")
	requireBodyField(t, logsRes.Body.String(), "items.0.targetType", "teacher")
	requireBodyField(t, logsRes.Body.String(), "items.0.targetUsername", "teacher-audit")
	requireBodyField(t, logsRes.Body.String(), "items.0.before.role", "teacher")
	requireBodyField(t, logsRes.Body.String(), "items.0.after.role", "admin")
	requireBodyField(t, logsRes.Body.String(), "items.1.action", "student.disable")
	requireBodyField(t, logsRes.Body.String(), "items.1.targetType", "student")
	requireBodyField(t, logsRes.Body.String(), "items.1.targetUsername", "student-audit")
	requireBodyField(t, logsRes.Body.String(), "items.1.before.status", "active")
	requireBodyField(t, logsRes.Body.String(), "items.1.after.status", "disabled")
	requireBodyField(t, logsRes.Body.String(), "items.2.action", "student.create")
	requireBodyField(t, logsRes.Body.String(), "items.2.after.teacherUsername", "teacher-audit")
	requireBodyField(t, logsRes.Body.String(), "items.3.action", "teacher.create")
	requireBodyField(t, logsRes.Body.String(), "items.3.after.username", "teacher-audit")
}

func TestTeacherCannotAccessAdminRoutes(t *testing.T) {
	t.Setenv("ADMIN_BOOTSTRAP_USERNAME", "admin")
	t.Setenv("ADMIN_BOOTSTRAP_PASSWORD", "admin12345")

	handler := newTestHandler()

	teacherToken := registerTeacher(t, handler, "teacher-no-admin", "secret123")

	listRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/admin/teachers", nil)
	require.Equal(t, http.StatusForbidden, listRes.Code)

	studentsRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/admin/students", nil)
	require.Equal(t, http.StatusForbidden, studentsRes.Code)

	createStudentRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodPost, "/api/admin/students", map[string]any{
		"teacherId":       1,
		"username":        "student-denied",
		"displayName":     "小紫",
		"initialPassword": "stud1234",
	})
	require.Equal(t, http.StatusForbidden, createStudentRes.Code)

	overviewRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/admin/overview", nil)
	require.Equal(t, http.StatusForbidden, overviewRes.Code)

	auditLogsRes := performAuthedJSONRequest(t, handler, teacherToken, http.MethodGet, "/api/admin/audit-logs", nil)
	require.Equal(t, http.StatusForbidden, auditLogsRes.Code)

	disableSelfRes := performAuthedJSONRequest(t, handler, loginTeacherToken(t, handler, "admin", "admin12345"), http.MethodPost, "/api/admin/teachers/1/disable", nil)
	require.Equal(t, http.StatusConflict, disableSelfRes.Code)

	demoteSelfRes := performAuthedJSONRequest(t, handler, loginTeacherToken(t, handler, "admin", "admin12345"), http.MethodPost, "/api/admin/teachers/1/role", map[string]any{
		"role": "teacher",
	})
	require.Equal(t, http.StatusConflict, demoteSelfRes.Code)
}

func loginTeacherToken(t *testing.T, handler http.Handler, username string, password string) string {
	t.Helper()

	loginRes := performJSONRequest(t, handler, http.MethodPost, "/api/teacher/login", map[string]any{
		"username": username,
		"password": password,
	})
	require.Equal(t, http.StatusOK, loginRes.Code)
	return requireStringField(t, loginRes.Body.String(), "token")
}
