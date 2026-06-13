package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/admin"
)

type adminHandler struct {
	service *admin.Service
}

// handleAdminTeachersList godoc
//
//	@Summary		List teachers
//	@Description	List all teacher/admin accounts for the current administrator.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	AdminTeachersResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Router			/api/admin/teachers [get]
func (h *adminHandler) handleAdminTeachersList(c *gin.Context) {
	if teacherRecord := mustTeacher(c); teacherRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, admin.TeachersResponse{
		Items: h.service.ListTeachers(),
	})
}

// handleAdminTeachersCreate godoc
//
//	@Summary		Create teacher
//	@Description	Create a managed teacher account with an initial password.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payload	body		AdminTeacherCreateRequestDoc	true	"Teacher create payload"
//	@Success		201		{object}	AdminTeacherItemResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Router			/api/admin/teachers [post]
func (h *adminHandler) handleAdminTeachersCreate(c *gin.Context) {
	if teacherRecord := mustTeacher(c); teacherRecord.ID == 0 {
		return
	}

	var request admin.CreateTeacherInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	createdTeacher, err := h.service.CreateTeacher(request)
	if err != nil {
		if errors.Is(err, admin.ErrTeacherConflict) {
			writeJSONError(c, 409, err.Error())
			return
		}
		writeJSONError(c, 500, "teacher create failed")
		return
	}

	writeJSON(c, 201, createdTeacher)
}

// handleAdminTeacherPasswordReset godoc
//
//	@Summary		Reset teacher password
//	@Description	Replace one managed teacher account password.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int64							true	"Teacher ID"
//	@Param			payload	body		AdminTeacherPasswordResetRequestDoc	true	"New password payload"
//	@Success		200		{object}	AdminTeacherItemResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/api/admin/teachers/{id}/reset-password [post]
func (h *adminHandler) handleAdminTeacherPasswordReset(c *gin.Context) {
	if teacherRecord := mustTeacher(c); teacherRecord.ID == 0 {
		return
	}

	teacherID, ok := parseIDParam(c, "id", "teacher")
	if !ok {
		return
	}

	var request admin.ResetTeacherPasswordInput
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	updatedTeacher, err := h.service.ResetTeacherPassword(teacherID, request.NewPassword)
	if err != nil {
		if errors.Is(err, admin.ErrTeacherNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "teacher password reset failed")
		return
	}

	writeJSON(c, 200, updatedTeacher)
}

// handleAdminTeacherDisable godoc
//
//	@Summary		Disable teacher
//	@Description	Disable one managed teacher account.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Teacher ID"
//	@Success		200	{object}	AdminTeacherItemResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		409	{object}	ErrorResponse
//	@Router			/api/admin/teachers/{id}/disable [post]
func (h *adminHandler) handleAdminTeacherDisable(c *gin.Context) {
	adminRecord := mustTeacher(c)
	if adminRecord.ID == 0 {
		return
	}

	teacherID, ok := parseIDParam(c, "id", "teacher")
	if !ok {
		return
	}

	updatedTeacher, err := h.service.DisableTeacher(adminRecord.ID, teacherID)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrTeacherNotFound):
			writeJSONError(c, 404, err.Error())
		case errors.Is(err, admin.ErrSelfProtected):
			writeJSONError(c, 409, err.Error())
		default:
			writeJSONError(c, 500, "teacher disable failed")
		}
		return
	}

	writeJSON(c, 200, updatedTeacher)
}

// handleAdminTeacherEnable godoc
//
//	@Summary		Enable teacher
//	@Description	Enable one managed teacher account.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int64	true	"Teacher ID"
//	@Success		200	{object}	AdminTeacherItemResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		403	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Router			/api/admin/teachers/{id}/enable [post]
func (h *adminHandler) handleAdminTeacherEnable(c *gin.Context) {
	if teacherRecord := mustTeacher(c); teacherRecord.ID == 0 {
		return
	}

	teacherID, ok := parseIDParam(c, "id", "teacher")
	if !ok {
		return
	}

	updatedTeacher, err := h.service.EnableTeacher(teacherID)
	if err != nil {
		if errors.Is(err, admin.ErrTeacherNotFound) {
			writeJSONError(c, 404, err.Error())
			return
		}
		writeJSONError(c, 500, "teacher enable failed")
		return
	}

	writeJSON(c, 200, updatedTeacher)
}
