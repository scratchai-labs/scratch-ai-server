package student

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

type BatchCreateInput struct {
	Username        string `json:"username" binding:"required"`
	DisplayName     string `json:"displayName" binding:"required"`
	InitialPassword string `json:"initialPassword" binding:"required"`
}

type BatchCreateRequest struct {
	Students []BatchCreateInput `json:"students" binding:"required,min=1,dive"`
}

type LoginInput struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	ClientType string `json:"clientType" binding:"required"`
}

type BatchCreateResult struct {
	Created   []StudentItem `json:"created"`
	Conflicts []string      `json:"conflicts"`
}

type StudentSession struct {
	Token       string `json:"token"`
	StudentName string `json:"studentName"`
}

type StudentItem struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

var (
	ErrInvalidCredentials = errors.New("invalid student credentials")
	ErrInvalidClientType  = errors.New("student login only supports desktop client")
	ErrUnauthorized       = errors.New("missing or invalid bearer token")
	ErrStudentNotFound    = errors.New("student not found")
)

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) BatchCreate(teacherID int64, input BatchCreateRequest) (BatchCreateResult, error) {
	created := make([]StudentItem, 0, len(input.Students))
	conflicts := make([]string, 0)

	for _, student := range input.Students {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(student.InitialPassword), bcrypt.DefaultCost)
		if err != nil {
			return BatchCreateResult{}, err
		}

		createdStudent, err := s.store.CreateStudent(teacherID, memory.CreateStudentInput{
			Username:     strings.TrimSpace(student.Username),
			DisplayName:  strings.TrimSpace(student.DisplayName),
			PasswordHash: string(passwordHash),
		})
		if err != nil {
			if err == memory.ErrStudentConflict {
				conflicts = append(conflicts, strings.TrimSpace(student.Username))
				continue
			}
			return BatchCreateResult{}, err
		}

		created = append(created, toStudentItem(createdStudent))
	}

	return BatchCreateResult{
		Created:   created,
		Conflicts: conflicts,
	}, nil
}

func (s *Service) List(teacherID int64) []StudentItem {
	students := s.store.ListStudentsByTeacher(teacherID)
	result := make([]StudentItem, 0, len(students))
	for _, student := range students {
		result = append(result, toStudentItem(student))
	}
	return result
}

func (s *Service) Login(input LoginInput) (StudentSession, error) {
	if strings.TrimSpace(input.ClientType) != "desktop" {
		return StudentSession{}, ErrInvalidClientType
	}

	studentRecord, ok := s.store.FindStudentByUsername(strings.TrimSpace(input.Username))
	if !ok {
		return StudentSession{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(studentRecord.PasswordHash), []byte(input.Password)); err != nil {
		return StudentSession{}, ErrInvalidCredentials
	}

	token, err := randomToken()
	if err != nil {
		return StudentSession{}, err
	}

	if err := s.store.SaveStudentToken(token, studentRecord.ID); err != nil {
		return StudentSession{}, err
	}
	return StudentSession{
		Token:       token,
		StudentName: studentRecord.DisplayName,
	}, nil
}

func (s *Service) StudentFromBearer(authorizationHeader string) (memory.Student, error) {
	token, err := bearerToken(authorizationHeader)
	if err != nil {
		return memory.Student{}, ErrUnauthorized
	}

	studentRecord, ok := s.store.FindStudentByToken(token)
	if !ok {
		return memory.Student{}, ErrUnauthorized
	}

	return studentRecord, nil
}

func (s *Service) Logout(authorizationHeader string) error {
	token, err := bearerToken(authorizationHeader)
	if err != nil {
		return ErrUnauthorized
	}

	if _, ok := s.store.FindStudentByToken(token); !ok {
		return ErrUnauthorized
	}

	return s.store.DeleteStudentToken(token)
}

func (s *Service) ResetPassword(teacherID int64, studentID int64, newPassword string) (StudentItem, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return StudentItem{}, err
	}

	updatedStudent, err := s.store.UpdateStudentPassword(teacherID, studentID, string(passwordHash))
	if err != nil {
		if errors.Is(err, memory.ErrStudentNotFound) {
			return StudentItem{}, ErrStudentNotFound
		}
		return StudentItem{}, err
	}

	return toStudentItem(updatedStudent), nil
}

func toStudentItem(student memory.Student) StudentItem {
	return StudentItem{
		ID:          student.ID,
		Username:    student.Username,
		DisplayName: student.DisplayName,
		Status:      student.Status,
		CreatedAt:   student.CreatedAt.Format(time.RFC3339),
	}
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
