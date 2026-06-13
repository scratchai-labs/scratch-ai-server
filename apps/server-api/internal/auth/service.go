package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var (
	ErrInvalidCredentials = errors.New("invalid teacher credentials")
	ErrTeacherConflict    = errors.New("teacher username already exists")
	ErrUnauthorized       = errors.New("missing or invalid bearer token")
	ErrTeacherDisabled    = errors.New("teacher account is disabled")
	ErrForbidden          = errors.New("forbidden")
)

type Session struct {
	Token       string `json:"token"`
	TeacherName string `json:"teacherName"`
	Role        string `json:"role"`
}

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) EnsureBootstrapAdmin(cfg config.AdminBootstrapConfig) error {
	if strings.TrimSpace(cfg.Username) == "" || strings.TrimSpace(cfg.Password) == "" {
		return nil
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.store.EnsureTeacher(strings.TrimSpace(cfg.Username), string(passwordHash), "admin", "active")
	return err
}

func (s *Service) Register(username string, password string) (Session, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Session{}, err
	}

	teacher, err := s.store.CreateTeacher(strings.TrimSpace(username), string(passwordHash))
	if err != nil {
		if errors.Is(err, memory.ErrTeacherConflict) {
			return Session{}, ErrTeacherConflict
		}
		return Session{}, err
	}

	return s.issueSession(teacher.ID, teacher.Username, teacher.Role)
}

func (s *Service) Login(username string, password string) (Session, error) {
	teacher, ok := s.store.FindTeacherByUsername(strings.TrimSpace(username))
	if !ok {
		return Session{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(teacher.PasswordHash), []byte(password)); err != nil {
		return Session{}, ErrInvalidCredentials
	}
	if teacher.Status != "active" {
		return Session{}, ErrTeacherDisabled
	}

	return s.issueSession(teacher.ID, teacher.Username, teacher.Role)
}

func (s *Service) TeacherFromBearer(authorizationHeader string) (memory.Teacher, error) {
	token, err := bearerToken(authorizationHeader)
	if err != nil {
		return memory.Teacher{}, ErrUnauthorized
	}

	teacher, ok := s.store.FindTeacherByToken(token)
	if !ok {
		return memory.Teacher{}, ErrUnauthorized
	}
	if teacher.Status != "active" {
		return memory.Teacher{}, ErrUnauthorized
	}

	return teacher, nil
}

func (s *Service) Logout(authorizationHeader string) error {
	token, err := bearerToken(authorizationHeader)
	if err != nil {
		return ErrUnauthorized
	}

	if _, ok := s.store.FindTeacherByToken(token); !ok {
		return ErrUnauthorized
	}

	return s.store.DeleteTeacherToken(token)
}

func (s *Service) issueSession(teacherID int64, teacherName string, role string) (Session, error) {
	token, err := randomToken()
	if err != nil {
		return Session{}, err
	}

	if err := s.store.SaveTeacherToken(token, teacherID); err != nil {
		return Session{}, err
	}
	return Session{
		Token:       token,
		TeacherName: teacherName,
		Role:        role,
	}, nil
}

func randomToken() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func bearerToken(authorizationHeader string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorizationHeader, prefix) {
		return "", ErrUnauthorized
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, prefix))
	if token == "" {
		return "", ErrUnauthorized
	}

	return token, nil
}
