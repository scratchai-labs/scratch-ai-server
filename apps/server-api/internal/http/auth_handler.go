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
	Username string `json:"username"`
	Password string `json:"password"`
}

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
		writeJSONError(c, 500, "teacher login failed")
		return
	}

	writeJSON(c, 200, session)
}

func (h *authHandler) handleTeacherMe(c *gin.Context) {
	teacherRecord := mustTeacher(c)
	if teacherRecord.ID == 0 {
		return
	}

	writeJSON(c, 200, map[string]any{
		"teacherId":   teacherRecord.ID,
		"teacherName": teacherRecord.Username,
	})
}

func (h *authHandler) handleTeacherLogout(c *gin.Context) {
	if err := h.service.Logout(c.GetHeader("Authorization")); err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			writeJSONError(c, 401, err.Error())
			return
		}
		writeJSONError(c, 500, "teacher logout failed")
		return
	}

	writeJSON(c, 200, map[string]any{
		"status": "ok",
	})
}
