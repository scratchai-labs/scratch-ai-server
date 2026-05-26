package http

import (
	"errors"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/assignment"
)

type assignmentHandler struct {
	assignmentService *assignment.Service
}

type assignStudentsRequest struct {
	StudentIDs []int64 `json:"studentIds"`
}

func (h *assignmentHandler) handleTeacherAssignmentsList(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, map[string]any{
		"items": h.assignmentService.ListForTeacher(teacherRecord.ID),
	})
}

func (h *assignmentHandler) handleTeacherAssignments(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	if err := c.Request.ParseMultipartForm(16 << 20); err != nil {
		writeJSONError(c, 400, "invalid multipart form")
		return
	}

	file, fileHeader, err := c.Request.FormFile("sb3")
	if err != nil {
		writeJSONError(c, 400, "sb3 file is required")
		return
	}
	defer file.Close()

	rawSB3, err := io.ReadAll(file)
	if err != nil {
		writeJSONError(c, 500, "failed to read sb3 file")
		return
	}

	createdAssignment, err := h.assignmentService.Upload(c.Request.Context(), teacherRecord.ID, assignment.UploadInput{
		Title:       strings.TrimSpace(c.Request.FormValue("title")),
		Goal:        strings.TrimSpace(c.Request.FormValue("goal")),
		Description: strings.TrimSpace(c.Request.FormValue("description")),
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		SB3Data:     rawSB3,
	})
	if err != nil {
		switch {
		case errors.Is(err, assignment.ErrInvalidSB3File), errors.Is(err, assignment.ErrInvalidSB3MIME), errors.Is(err, assignment.ErrSB3TooLarge):
			writeJSONError(c, 400, err.Error())
		default:
			writeJSONError(c, 500, "assignment upload failed")
		}
		return
	}

	writeJSON(c, 201, map[string]any{
		"id":             createdAssignment.ID,
		"title":          createdAssignment.Title,
		"goal":           createdAssignment.Goal,
		"description":    createdAssignment.Description,
		"status":         createdAssignment.Status,
		"analysisStatus": createdAssignment.AnalysisStatus,
	})
}

func (h *assignmentHandler) handleTeacherAssignmentDetail(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	detail, err := h.assignmentService.GetForTeacher(teacherRecord.ID, assignmentID)
	if err != nil {
		if errors.Is(err, assignment.ErrAssignmentNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "assignment detail lookup failed")
		return
	}

	writeJSON(c, 200, detail)
}

func (h *assignmentHandler) handleTeacherAssignmentAnalysis(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	assignmentRecord, err := h.assignmentService.GetAnalysis(teacherRecord.ID, assignmentID)
	if err != nil {
		if errors.Is(err, assignment.ErrAssignmentNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "assignment analysis lookup failed")
		return
	}

	writeJSON(c, 200, map[string]any{
		"assignmentId":         assignmentRecord.ID,
		"analysisStatus":       assignmentRecord.AnalysisStatus,
		"analysisErrorMessage": assignmentRecord.AnalysisErrorMessage,
		"roleNames":            assignmentRecord.Analysis.RoleNames,
		"scriptCounts":         assignmentRecord.Analysis.ScriptCounts,
		"blockCounts":          assignmentRecord.Analysis.BlockCounts,
		"categoryCounts":       assignmentRecord.Analysis.CategoryCounts,
		"broadcastMessages":    assignmentRecord.Analysis.BroadcastMessages,
		"variableNames":        assignmentRecord.Analysis.VariableNames,
		"listNames":            assignmentRecord.Analysis.ListNames,
		"extensions":           assignmentRecord.Analysis.Extensions,
		"teachingPoints":       assignmentRecord.Analysis.TeachingPoints,
	})
}

func (h *assignmentHandler) handleTeacherAssignmentStudents(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	var request assignStudentsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	if err := h.assignmentService.AssignStudents(teacherRecord.ID, assignmentID, request.StudentIDs); err != nil {
		switch {
		case errors.Is(err, assignment.ErrAssignmentNotFound):
			writeJSONError(c, 404, err.Error())
		case errors.Is(err, assignment.ErrStudentNotFound):
			writeJSONError(c, 404, err.Error())
		default:
			writeJSONError(c, 500, "assignment student binding failed")
		}
		return
	}

	writeJSON(c, 200, map[string]any{
		"assignmentId":  assignmentID,
		"studentIds":    request.StudentIDs,
		"assignedCount": len(request.StudentIDs),
	})
}

func (h *assignmentHandler) handleTeacherAssignmentPublish(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	record, err := h.assignmentService.Publish(teacherRecord.ID, assignmentID)
	if err != nil {
		switch {
		case errors.Is(err, assignment.ErrAssignmentNotFound):
			writeJSONError(c, 404, err.Error())
		case errors.Is(err, assignment.ErrAssignmentNotReady):
			writeJSONError(c, 409, err.Error())
		default:
			writeJSONError(c, 500, "assignment publish failed")
		}
		return
	}

	writeJSON(c, 200, map[string]any{
		"id":             record.ID,
		"title":          record.Title,
		"status":         record.Status,
		"analysisStatus": record.AnalysisStatus,
	})
}

func (h *assignmentHandler) handleTeacherAssignmentArchive(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	assignmentID, ok := parseIDParam(c, "id", "assignment")
	if !ok {
		return
	}

	record, err := h.assignmentService.Archive(teacherRecord.ID, assignmentID)
	if err != nil {
		if errors.Is(err, assignment.ErrAssignmentNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "assignment archive failed")
		return
	}

	writeJSON(c, 200, map[string]any{
		"id":     record.ID,
		"title":  record.Title,
		"status": record.Status,
	})
}
