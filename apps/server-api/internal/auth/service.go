package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var (
	ErrInvalidCredentials = errors.New("invalid teacher credentials")
	ErrTeacherConflict    = errors.New("teacher username already exists")
	ErrUnauthorized       = errors.New("missing or invalid bearer token")
)

type Session struct {
	Token       string `json:"token"`
	TeacherName string `json:"teacherName"`
}

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
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

	return s.issueSession(teacher.ID, teacher.Username)
}

func (s *Service) Login(username string, password string) (Session, error) {
	teacher, ok := s.store.FindTeacherByUsername(strings.TrimSpace(username))
	if !ok {
		return Session{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(teacher.PasswordHash), []byte(password)); err != nil {
		return Session{}, ErrInvalidCredentials
	}

	return s.issueSession(teacher.ID, teacher.Username)
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

	s.store.DeleteTeacherToken(token)
	return nil
}

func (s *Service) issueSession(teacherID int64, teacherName string) (Session, error) {
	token, err := randomToken()
	if err != nil {
		return Session{}, err
	}

	s.store.SaveTeacherToken(token, teacherID)
	return Session{
		Token:       token,
		TeacherName: teacherName,
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
