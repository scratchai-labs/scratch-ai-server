package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/dashboard"
)

type dashboardHandler struct {
	dashboardService *dashboard.Service
}

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
