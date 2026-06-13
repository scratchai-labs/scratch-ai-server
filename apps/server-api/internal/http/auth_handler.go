package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/auth"
)

type authHandler struct {
	service *auth.Service
}

type teacherCredentialsRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// handleTeacherRegister godoc
//
//	@Summary		Register a teacher account
//	@Description	Create a teacher session and return a bearer token for teacher APIs.
//	@Tags			teacher-auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		TeacherCredentialsRequestDoc	true	"Teacher credentials"
//	@Success		201		{object}	auth.Session
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/register [post]
func (h *authHandler) handleTeacherRegister(c *gin.Context) {
	var request teacherCredentialsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	session, err := h.service.Register(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, auth.ErrTeacherConflict) {
			writeJSONError(c, 409, err.Error())
			return
		}
		writeJSONError(c, 500, "teacher register failed")
		return
	}

	writeJSON(c, 201, session)
}

// handleTeacherLogin godoc
//
//	@Summary		Login as teacher
//	@Description	Authenticate a teacher and return a bearer token for teacher APIs.
//	@Tags			teacher-auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		TeacherCredentialsRequestDoc	true	"Teacher credentials"
//	@Success		200		{object}	auth.Session
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/teacher/login [post]
func (h *authHandler) handleTeacherLogin(c *gin.Context) {
	var request teacherCredentialsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeJSONError(c, 400, "invalid request body")
		return
	}

	session, err := h.service.Login(request.Username, request.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeJSONError(c, 401, err.Error())
			return
		}
		if errors.Is(err, auth.ErrTeacherDisabled) {
			writeJSONError(c, 403, err.Error())
			return
		}
		writeJSONError(c, 500, "teacher login failed")
		return
	}

	writeJSON(c, 200, session)
}

// handleTeacherMe godoc
//
//	@Summary		Get current teacher profile
//	@Description	Read the current authenticated teacher session details.
//	@Tags			teacher-auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	TeacherMeResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/me [get]
func (h *authHandler) handleTeacherMe(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, TeacherMeResponse{
		TeacherID:   teacherRecord.ID,
		TeacherName: teacherRecord.Username,
		Role:        teacherRecord.Role,
	})
}

// handleTeacherLogout godoc
//
//	@Summary		Logout teacher
//	@Description	Invalidate the current teacher bearer token immediately.
//	@Tags			teacher-auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	StatusResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/teacher/logout [post]
func (h *authHandler) handleTeacherLogout(c *gin.Context) {
	if err := h.service.Logout(c.GetHeader("Authorization")); err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			writeJSONError(c, 401, err.Error())
			return
		}
		writeJSONError(c, 500, "teacher logout failed")
		return
	}

	writeJSON(c, 200, StatusResponse{Status: "ok"})
}
