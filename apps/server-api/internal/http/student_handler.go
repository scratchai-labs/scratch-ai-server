package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/assignment"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/hint"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/progress"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/student"
)

type studentHandler struct {
	studentService    *student.Service
	assignmentService *assignment.Service
	progressService   *progress.Service
	hintService       *hint.Service
}

type resetPasswordRequest struct {
	NewPassword string `json:"newPassword"`
}

func (h *studentHandler) handleTeacherStudents(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	switch c.Request.Method {
	case "GET":
		writeJSON(c, 200, map[string]any{
			"items": h.studentService.List(teacherRecord.ID),
		})
	case "POST":
		var request student.BatchCreateInput
		if err := c.ShouldBindJSON(&request); err != nil {
			writeJSONError(c, 400, "invalid request body")
			return
		}

		result, err := h.studentService.BatchCreate(teacherRecord.ID, student.BatchCreateRequest{
			Students: []student.BatchCreateInput{request},
		})
		if err != nil {
			writeJSONError(c, 500, "student create failed")
			return
		}

		writeJSON(c, 201, result)
	default:
		writeMethodNotAllowed(c)
	}
}

func (h *studentHandler) handleTeacherStudentsBatch(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	var request student.BatchCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	result, err := h.studentService.BatchCreate(teacherRecord.ID, request)
	if err != nil {
		writeJSONError(c, 500, "student batch create failed")
		return
	}

	writeJSON(c, 201, result)
}

func (h *studentHandler) handleTeacherStudentPasswordReset(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	studentID, ok := parseIDParam(c, "id", "student")
	if !ok {
		return
	}

	var request resetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	updatedStudent, err := h.studentService.ResetPassword(teacherRecord.ID, studentID, request.NewPassword)
	if err != nil {
		if errors.Is(err, student.ErrStudentNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "student password reset failed")
		return
	}

	writeJSON(c, 200, updatedStudent)
}

func (h *studentHandler) handleStudentLogin(c *gin.Context) {
	var request student.LoginInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	session, err := h.studentService.Login(request)
	if err != nil {
		switch {
		case errors.Is(err, student.ErrInvalidClientType):
			writeJSONError(c, 400, err.Error())
		case errors.Is(err, student.ErrInvalidCredentials):
			writeJSONError(c, 401, err.Error())
		default:
			writeJSONError(c, 500, "student login failed")
		}
		return
	}

	writeJSON(c, 200, session)
}

func (h *studentHandler) handleStudentMe(c *gin.Context) {
	studentRecord := mustStudent(c)
	if studentRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, map[string]any{
		"studentId":   studentRecord.ID,
		"studentName": studentRecord.DisplayName,
		"username":    studentRecord.Username,
	})
}

func (h *studentHandler) handleStudentLogout(c *gin.Context) {
	if err := h.studentService.Logout(c.GetHeader("Authorization")); err != nil {
		if errors.Is(err, student.ErrUnauthorized) {
			writeJSONError(c, 401, err.Error())
			return
		}
		writeJSONError(c, 500, "student logout failed")
		return
	}

	writeJSON(c, 200, map[string]any{
		"status": "ok",
	})
}

func (h *studentHandler) handleStudentAssignments(c *gin.Context) {
	studentRecord := mustStudent(c)
	if studentRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, map[string]any{
		"items": h.assignmentService.ListForStudent(studentRecord.ID),
	})
}

func (h *studentHandler) handleStudentAssignmentDetail(c *gin.Context) {
	studentRecord := mustStudent(c)
	if studentRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	detail, err := h.assignmentService.GetDetailForStudent(studentRecord.ID, assignmentID)
	if err != nil {
		if errors.Is(err, assignment.ErrAssignmentUnavailable) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "student assignment detail query failed")
		return
	}

	writeJSON(c, 200, detail)
}

func (h *studentHandler) handleStudentAssignmentProgress(c *gin.Context) {
	studentRecord := mustStudent(c)
	if studentRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	var request progress.ReportInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	report, err := h.progressService.Report(studentRecord.ID, assignmentID, request)
	if err != nil {
		if errors.Is(err, progress.ErrAssignmentUnavailable) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "student progress report failed")
		return
	}

	writeJSON(c, 201, map[string]any{
		"id":            report.ID,
		"assignmentId":  report.AssignmentID,
		"currentTarget": report.CurrentTarget,
		"stepSummary":   report.StepSummary,
		"reportedAt":    report.ReportedAt,
	})
}

func (h *studentHandler) handleStudentAssignmentHint(c *gin.Context) {
	studentRecord := mustStudent(c)
	if studentRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	record, err := h.hintService.Request(c.Request.Context(), studentRecord.ID, assignmentID)
	if err != nil {
		switch {
		case errors.Is(err, hint.ErrAssignmentUnavailable):
			writeJSONError(c, 404, err.Error())
		case errors.Is(err, hint.ErrAssignmentNotReady):
			writeJSONError(c, 409, err.Error())
		case errors.Is(err, hint.ErrProgressRequired):
			writeJSONError(c, 409, err.Error())
		default:
			writeJSONError(c, 500, "student hint request failed")
		}
		return
	}

	writeJSON(c, 201, map[string]any{
		"id":           record.ID,
		"assignmentId": record.AssignmentID,
		"hintText":     record.HintText,
		"providerName": record.ProviderName,
	})
}

func parseIDParam(c *gin.Context, name string, resource string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		writeJSONError(c, 400, "invalid "+resource+" id")
		return 0, false
	}
	return value, true
}
