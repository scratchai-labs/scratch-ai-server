package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/dashboard"
)

type dashboardHandler struct {
	dashboardService *dashboard.Service
}

// handleTeacherLiveDashboard godoc
//
//	@Summary		Get live assignment dashboard
//	@Description	Read the latest reported progress and hint state for all students assigned to one assignment.
//	@Tags			dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Assignment ID"
//	@Success		200	{object}	dashboard.LiveAssignmentResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/dashboard/assignments/{id}/live [get]
//	@Router			/api/teacher/projects/{id}/live [get]
func (h *dashboardHandler) handleTeacherLiveDashboard(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	response, err := h.dashboardService.LiveAssignment(teacherRecord.ID, assignmentID)
	if err != nil {
		if errors.Is(err, dashboard.ErrAssignmentNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "live dashboard query failed")
		return
	}

	writeJSON(c, 200, response)
}

// handleTeacherStudentHistory godoc
//
//	@Summary		Get teacher view of one student's assignment history
//	@Description	Read the latest progress and hint state for every assignment assigned to one student.
//	@Tags			dashboard
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Student ID"
//	@Success		200	{object}	dashboard.StudentHistoryResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/dashboard/students/{id}/history [get]
func (h *dashboardHandler) handleTeacherStudentHistory(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	studentID, ok := parseIDParam(c, "id", "student")
	if !ok {
		return
	}

	response, err := h.dashboardService.StudentHistory(teacherRecord.ID, studentID)
	if err != nil {
		if errors.Is(err, dashboard.ErrStudentNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "student history query failed")
		return
	}

	writeJSON(c, 200, response)
}
