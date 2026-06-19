package http

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/assignment"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

type assignmentHandler struct {
	assignmentService *assignment.Service
}

type assignStudentsRequest struct {
	StudentIDs []int64 `json:"studentIds" binding:"required,min=1"`
}

const maxSB3MultipartBodyBytes = assignment.MaxSB3Bytes + (1 << 20)

// handleTeacherAssignmentsList godoc
//
//	@Summary		List teacher assignments
//	@Description	List all assignments created by the current teacher.
//	@Tags			assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	TeacherAssignmentsResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/assignments [get]
func (h *assignmentHandler) handleTeacherAssignmentsList(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, TeacherAssignmentsResponse{
		Items: h.assignmentService.ListForTeacher(teacherRecord.ID),
	})
}

// handleTeacherAssignments godoc
//
//	@Summary		Upload a reference SB3 assignment
//	@Description	Create an assignment by uploading a reference Scratch project and queue analysis.
//	@Tags			assignments
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			title		formData	string	false	"Assignment title"
//	@Param			goal		formData	string	false	"Teaching goal"
//	@Param			description	formData	string	false	"Assignment description"
//	@Param			sb3			formData	file	true	"Scratch project .sb3 file"
//	@Success		201			{object}	AssignmentUploadResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/api/teacher/assignments [post]
func (h *assignmentHandler) handleTeacherAssignments(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSB3MultipartBodyBytes)
	if err := c.Request.ParseMultipartForm(assignment.MaxSB3Bytes); err != nil {
		writeJSONError(c, 400, "invalid multipart form")
		return
	}

	file, fileHeader, err := c.Request.FormFile("sb3")
	if err != nil {
		writeJSONError(c, 400, "sb3 file is required")
		return
	}
	defer file.Close()

	rawSB3, err := readSB3Upload(file, assignment.MaxSB3Bytes)
	if err != nil {
		if errors.Is(err, assignment.ErrSB3TooLarge) {
			writeJSONError(c, 400, err.Error())
			return
		}
		writeJSONError(c, 500, "failed to read sb3 file")
		return
	}

	createdAssignment, err := h.assignmentService.UploadLegacy(c.Request.Context(), teacherRecord.ID, assignment.UploadInput{
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

	writeJSON(c, 201, AssignmentUploadResponse{
		ID:             createdAssignment.ID,
		Title:          createdAssignment.Title,
		Goal:           createdAssignment.Goal,
		Description:    createdAssignment.Description,
		Status:         createdAssignment.Status,
		AnalysisStatus: createdAssignment.AnalysisStatus,
	})
}

// handleTeacherClassProjectsList godoc
//
//	@Summary		List class projects
//	@Description	List all projects created under one classroom.
//	@Tags			assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Classroom ID"
//	@Success		200	{object}	TeacherAssignmentsResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/classes/{id}/projects [get]
func (h *assignmentHandler) handleTeacherClassProjectsList(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	classroomID, ok := parseIDParam(c, "id", "classroom")
	if !ok {
		return
	}

	writeJSON(c, 200, TeacherAssignmentsResponse{
		Items: h.assignmentService.ListForTeacherByClassroom(teacherRecord.ID, classroomID),
	})
}

// handleTeacherClassProjectsCreate godoc
//
//	@Summary		Upload a class project
//	@Description	Create a project under one classroom by uploading a reference Scratch project.
//	@Tags			assignments
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int64	true	"Classroom ID"
//	@Param			title		formData	string	false	"Assignment title"
//	@Param			goal		formData	string	false	"Teaching goal"
//	@Param			description	formData	string	false	"Assignment description"
//	@Param			sb3			formData	file	true	"Scratch project .sb3 file"
//	@Success		201		{object}	AssignmentUploadResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/classes/{id}/projects [post]
func (h *assignmentHandler) handleTeacherClassProjectsCreate(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	classroomID, ok := parseIDParam(c, "id", "classroom")
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSB3MultipartBodyBytes)
	if err := c.Request.ParseMultipartForm(assignment.MaxSB3Bytes); err != nil {
		writeJSONError(c, 400, "invalid multipart form")
		return
	}

	file, fileHeader, err := c.Request.FormFile("sb3")
	if err != nil {
		writeJSONError(c, 400, "sb3 file is required")
		return
	}
	defer file.Close()

	rawSB3, err := readSB3Upload(file, assignment.MaxSB3Bytes)
	if err != nil {
		if errors.Is(err, assignment.ErrSB3TooLarge) {
			writeJSONError(c, 400, err.Error())
			return
		}
		writeJSONError(c, 500, "failed to read sb3 file")
		return
	}

	createdAssignment, err := h.assignmentService.Upload(c.Request.Context(), teacherRecord.ID, assignment.UploadInput{
		ClassroomID:  classroomID,
		Title:        strings.TrimSpace(c.Request.FormValue("title")),
		Goal:         strings.TrimSpace(c.Request.FormValue("goal")),
		Description:  strings.TrimSpace(c.Request.FormValue("description")),
		FileName:     fileHeader.Filename,
		ContentType:  fileHeader.Header.Get("Content-Type"),
		SB3Data:      rawSB3,
	})
	if err != nil {
		switch {
		case errors.Is(err, assignment.ErrInvalidSB3File), errors.Is(err, assignment.ErrInvalidSB3MIME), errors.Is(err, assignment.ErrSB3TooLarge):
			writeJSONError(c, 400, err.Error())
		case errors.Is(err, memory.ErrClassroomNotFound):
			writeJSONError(c, 404, "classroom not found")
		default:
			writeJSONError(c, 500, "project upload failed")
		}
		return
	}

	writeJSON(c, 201, AssignmentUploadResponse{
		ID:             createdAssignment.ID,
		Title:          createdAssignment.Title,
		Goal:           createdAssignment.Goal,
		Description:    createdAssignment.Description,
		Status:         createdAssignment.Status,
		AnalysisStatus: createdAssignment.AnalysisStatus,
	})
}

func readSB3Upload(reader io.Reader, maxBytes int64) ([]byte, error) {
	limitedReader := &io.LimitedReader{R: reader, N: maxBytes + 1}
	rawSB3, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if int64(len(rawSB3)) > maxBytes {
		return nil, assignment.ErrSB3TooLarge
	}
	return rawSB3, nil
}

// handleTeacherAssignmentDetail godoc
//
//	@Summary		Get teacher assignment detail
//	@Description	Read assignment metadata, analysis summary, and assigned students for the current teacher.
//	@Tags			assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Assignment ID"
//	@Success		200	{object}	assignment.TeacherAssignmentDetail
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/assignments/{id} [get]
//	@Router			/api/teacher/projects/{id} [get]
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

// handleTeacherAssignmentAnalysis godoc
//
//	@Summary		Get assignment analysis summary
//	@Description	Read the parsed SB3 analysis for a teacher assignment.
//	@Tags			assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Assignment ID"
//	@Success		200	{object}	AssignmentAnalysisResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/assignments/{id}/analysis [get]
//	@Router			/api/teacher/projects/{id}/analysis [get]
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

	writeJSON(c, 200, AssignmentAnalysisResponse{
		AssignmentID:         assignmentRecord.ID,
		AnalysisStatus:       assignmentRecord.AnalysisStatus,
		AnalysisErrorMessage: assignmentRecord.AnalysisErrorMessage,
		RoleNames:            assignmentRecord.Analysis.RoleNames,
		ScriptCounts:         assignmentRecord.Analysis.ScriptCounts,
		BlockCounts:          assignmentRecord.Analysis.BlockCounts,
		CategoryCounts:       assignmentRecord.Analysis.CategoryCounts,
		BroadcastMessages:    assignmentRecord.Analysis.BroadcastMessages,
		VariableNames:        assignmentRecord.Analysis.VariableNames,
		ListNames:            assignmentRecord.Analysis.ListNames,
		Extensions:           assignmentRecord.Analysis.Extensions,
		TeachingPoints:       assignmentRecord.Analysis.TeachingPoints,
	})
}

// handleTeacherAssignmentStudents godoc
//
//	@Summary		Assign students to an assignment
//	@Description	Bind one or more existing students to an assignment before publishing it.
//	@Tags			assignments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int64					true	"Assignment ID"
//	@Param			payload	body		assignStudentsRequest	true	"Student assignment payload"
//	@Success		200		{object}	AssignmentStudentsResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/assignments/{id}/assign-students [post]
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

	writeJSON(c, 200, AssignmentStudentsResponse{
		AssignmentID:  assignmentID,
		StudentIDs:    request.StudentIDs,
		AssignedCount: len(request.StudentIDs),
	})
}

// handleTeacherAssignmentPublish godoc
//
//	@Summary		Publish an assignment
//	@Description	Publish an analyzed assignment so students can access it.
//	@Tags			assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Assignment ID"
//	@Success		200	{object}	AssignmentStatusResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		409	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/assignments/{id}/publish [post]
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

	writeJSON(c, 200, AssignmentStatusResponse{
		ID:             record.ID,
		Title:          record.Title,
		Status:         record.Status,
		AnalysisStatus: record.AnalysisStatus,
	})
}

// handleTeacherAssignmentArchive godoc
//
//	@Summary		Archive an assignment
//	@Description	Archive an existing assignment so students no longer see it in active lists.
//	@Tags			assignments
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Assignment ID"
//	@Success		200	{object}	AssignmentStatusResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/assignments/{id}/archive [post]
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

	writeJSON(c, 200, AssignmentStatusResponse{
		ID:     record.ID,
		Title:  record.Title,
		Status: record.Status,
	})
}
