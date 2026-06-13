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
	NewPassword string `json:"newPassword" binding:"required"`
}

// handleTeacherStudentsList godoc
//
//	@Summary		List students
//	@Description	List students for the current teacher.
//	@Tags			students
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	TeacherStudentsResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/students [get]
func (h *studentHandler) handleTeacherStudentsList(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, TeacherStudentsResponse{
		Items: h.studentService.List(teacherRecord.ID),
	})
}

// handleTeacherStudentCreate godoc
//
//	@Summary		Create one student
//	@Description	Create a single student for the current teacher, returning the same shape as batch creation.
//	@Tags			students
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payload	body		student.BatchCreateInput	true	"Single student create payload"
//	@Success		201		{object}	student.BatchCreateResult
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/students [post]
func (h *studentHandler) handleTeacherStudentCreate(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

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
}

// handleTeacherStudentsBatch godoc
//
//	@Summary		Batch create students
//	@Description	Create multiple students for the current teacher and report conflicts without failing the entire request.
//	@Tags			students
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payload	body		student.BatchCreateRequest	true	"Batch create payload"
//	@Success		201		{object}	student.BatchCreateResult
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/students/batch [post]
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

// handleTeacherStudentPasswordReset godoc
//
//	@Summary		Reset a student password
//	@Description	Replace the initial or current password for one student owned by the current teacher.
//	@Tags			students
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int64					true	"Student ID"
//	@Param			payload	body		resetPasswordRequest	true	"New password payload"
//	@Success		200		{object}	student.StudentItem
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/students/{id}/reset-password [post]
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

// handleStudentLogin godoc
//
//	@Summary		Login as student
//	@Description	Authenticate a student from the desktop client and return a bearer token.
//	@Tags			student-auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		StudentLoginRequestDoc	true	"Student login payload"
//	@Success		200		{object}	student.StudentSession
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/student/login [post]
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
		case errors.Is(err, student.ErrStudentDisabled):
			writeJSONError(c, 403, err.Error())
		case errors.Is(err, student.ErrInvalidCredentials):
			writeJSONError(c, 401, err.Error())
		default:
			writeJSONError(c, 500, "student login failed")
		}
		return
	}

	writeJSON(c, 200, session)
}

// handleStudentMe godoc
//
//	@Summary		Get current student profile
//	@Description	Read the current authenticated student session details.
//	@Tags			student-auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	StudentMeResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/student/me [get]
func (h *studentHandler) handleStudentMe(c *gin.Context) {
	studentRecord := mustStudent(c)
	if studentRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, StudentMeResponse{
		StudentID:   studentRecord.ID,
		StudentName: studentRecord.DisplayName,
		Username:    studentRecord.Username,
	})
}

// handleStudentLogout godoc
//
//	@Summary		Logout student
//	@Description	Invalidate the current student bearer token immediately.
//	@Tags			student-auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	StatusResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/student/logout [post]
func (h *studentHandler) handleStudentLogout(c *gin.Context) {
	if err := h.studentService.Logout(c.GetHeader("Authorization")); err != nil {
		if errors.Is(err, student.ErrUnauthorized) {
			writeJSONError(c, 401, err.Error())
			return
		}
		writeJSONError(c, 500, "student logout failed")
		return
	}

	writeJSON(c, 200, StatusResponse{Status: "ok"})
}

// handleStudentAssignments godoc
//
//	@Summary		List student assignments
//	@Description	List the assignments currently published and assigned to the authenticated student.
//	@Tags			student-assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	StudentAssignmentsResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/student/assignments [get]
func (h *studentHandler) handleStudentAssignments(c *gin.Context) {
	studentRecord := mustStudent(c)
	if studentRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, StudentAssignmentsResponse{
		Items: h.assignmentService.ListForStudent(studentRecord.ID),
	})
}

// handleStudentAssignmentDetail godoc
//
//	@Summary		Get student assignment detail
//	@Description	Read the assignment detail plus latest progress and hint summaries for the authenticated student.
//	@Tags			student-assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Assignment ID"
//	@Success		200	{object}	assignment.StudentAssignmentDetail
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/student/assignments/{id} [get]
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

// handleStudentAssignmentProgress godoc
//
//	@Summary		Report student progress
//	@Description	Upload the latest structured Scratch progress snapshot for one assigned assignment.
//	@Tags			student-assignments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int64				true	"Assignment ID"
//	@Param			payload	body		ProgressReportRequestDoc	true	"Progress payload"
//	@Success		201		{object}	ProgressCreatedResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/student/assignments/{id}/progress [post]
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

	writeJSON(c, 201, ProgressCreatedResponse{
		ID:            report.ID,
		AssignmentID:  report.AssignmentID,
		CurrentTarget: report.CurrentTarget,
		StepSummary:   report.StepSummary,
		ReportedAt:    report.ReportedAt,
	})
}

// handleStudentAssignmentHint godoc
//
//	@Summary		Request the next student hint
//	@Description	Generate the next hint from assignment analysis plus the student's latest reported progress.
//	@Tags			student-assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Assignment ID"
//	@Success		201	{object}	HintCreatedResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		409	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/student/assignments/{id}/hints [post]
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

	writeJSON(c, 201, HintCreatedResponse{
		ID:           record.ID,
		AssignmentID: record.AssignmentID,
		HintText:     record.HintText,
		ProviderName: record.ProviderName,
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
