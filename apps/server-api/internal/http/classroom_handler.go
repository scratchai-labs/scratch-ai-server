package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/classroom"
)

type classroomHandler struct {
	service *classroom.Service
}

func (h *classroomHandler) handleTeacherClassesList(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	writeJSON(c, http.StatusOK, TeacherClassroomsResponse{
		Items: h.service.List(teacherRecord.ID),
	})
}

// handleTeacherClassesCreate godoc
//
//	@Summary		Create classroom
//	@Description	Create one classroom for the current teacher.
//	@Tags			classrooms
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payload	body		classroom.CreateInput	true	"Classroom create payload"
//	@Success		201		{object}	classroom.Item
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/classes [post]
func (h *classroomHandler) handleTeacherClassesCreate(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	var request classroom.CreateInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	record, err := h.service.Create(teacherRecord.ID, request)
	if err != nil {
		writeJSONError(c, http.StatusInternalServerError, "classroom create failed")
		return
	}

	writeJSON(c, http.StatusCreated, record)
}

// handleTeacherClassDetail godoc
//
//	@Summary		Get classroom detail
//	@Description	Read one classroom plus its student and project counts for the current teacher.
//	@Tags			classrooms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Classroom ID"
//	@Success		200	{object}	classroom.Detail
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/classes/{id} [get]
func (h *classroomHandler) handleTeacherClassDetail(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	classroomID, ok := parseIDParam(c, "id", "classroom")
	if !ok {
		return
	}

	record, err := h.service.Get(teacherRecord.ID, classroomID)
	if err != nil {
		if errors.Is(err, classroom.ErrClassroomNotFound) {
			writeJSONError(c, http.StatusNotFound, err.Error())
			return
		}
		writeJSONError(c, http.StatusInternalServerError, "classroom detail query failed")
		return
	}

	writeJSON(c, http.StatusOK, record)
}

// handleTeacherClassUpdate godoc
//
//	@Summary		Update classroom
//	@Description	Update one classroom name for the current teacher.
//	@Tags			classrooms
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int64					true	"Classroom ID"
//	@Param			payload	body		classroom.UpdateInput	true	"Classroom update payload"
//	@Success		200		{object}	classroom.Item
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/classes/{id} [patch]
func (h *classroomHandler) handleTeacherClassUpdate(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	classroomID, ok := parseIDParam(c, "id", "classroom")
	if !ok {
		return
	}

	var request classroom.UpdateInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	record, err := h.service.Update(teacherRecord.ID, classroomID, request)
	if err != nil {
		if errors.Is(err, classroom.ErrClassroomNotFound) {
			writeJSONError(c, http.StatusNotFound, err.Error())
			return
		}
		writeJSONError(c, http.StatusInternalServerError, "classroom update failed")
		return
	}

	writeJSON(c, http.StatusOK, record)
}

// handleTeacherClassDelete godoc
//
//	@Summary		Delete classroom
//	@Description	Delete one empty classroom for the current teacher.
//	@Tags			classrooms
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Classroom ID"
//	@Success		200	{object}	StatusResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		409	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/classes/{id} [delete]
func (h *classroomHandler) handleTeacherClassDelete(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	classroomID, ok := parseIDParam(c, "id", "classroom")
	if !ok {
		return
	}

	err := h.service.Delete(teacherRecord.ID, classroomID)
	if err != nil {
		switch {
		case errors.Is(err, classroom.ErrClassroomNotFound):
			writeJSONError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, classroom.ErrClassroomNotEmpty):
			writeJSONError(c, http.StatusConflict, err.Error())
		default:
			writeJSONError(c, http.StatusInternalServerError, "classroom delete failed")
		}
		return
	}

	writeJSON(c, http.StatusOK, StatusResponse{Status: "ok"})
}
