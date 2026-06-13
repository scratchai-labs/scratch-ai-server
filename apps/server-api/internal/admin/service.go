package admin

import (
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

type CreateTeacherInput struct {
	Username        string `json:"username" binding:"required"`
	InitialPassword string `json:"initialPassword" binding:"required"`
}

type ResetTeacherPasswordInput struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

type TeacherItem struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type TeachersResponse struct {
	Items []TeacherItem `json:"items"`
}

var (
	ErrTeacherConflict = errors.New("teacher username already exists")
	ErrTeacherNotFound = errors.New("teacher not found")
	ErrSelfProtected   = errors.New("admin cannot disable self")
)

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListTeachers() []TeacherItem {
	teachers := s.store.ListTeachers()
	items := make([]TeacherItem, 0, len(teachers))
	for _, teacher := range teachers {
		items = append(items, toTeacherItem(teacher))
	}
	return items
}

func (s *Service) CreateTeacher(input CreateTeacherInput) (TeacherItem, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.InitialPassword), bcrypt.DefaultCost)
	if err != nil {
		return TeacherItem{}, err
	}

	teacher, err := s.store.CreateTeacher(strings.TrimSpace(input.Username), string(passwordHash))
	if err != nil {
		if errors.Is(err, memory.ErrTeacherConflict) {
			return TeacherItem{}, ErrTeacherConflict
		}
		return TeacherItem{}, err
	}

	return toTeacherItem(teacher), nil
}

func (s *Service) ResetTeacherPassword(teacherID int64, newPassword string) (TeacherItem, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return TeacherItem{}, err
	}

	teacher, err := s.store.UpdateTeacherPassword(teacherID, string(passwordHash))
	if err != nil {
		if errors.Is(err, memory.ErrTeacherNotFound) {
			return TeacherItem{}, ErrTeacherNotFound
		}
		return TeacherItem{}, err
	}

	return toTeacherItem(teacher), nil
}

func (s *Service) DisableTeacher(adminID int64, teacherID int64) (TeacherItem, error) {
	if adminID == teacherID {
		return TeacherItem{}, ErrSelfProtected
	}

	teacher, err := s.store.UpdateTeacherStatus(teacherID, "disabled")
	if err != nil {
		if errors.Is(err, memory.ErrTeacherNotFound) {
			return TeacherItem{}, ErrTeacherNotFound
		}
		return TeacherItem{}, err
	}

	return toTeacherItem(teacher), nil
}

func (s *Service) EnableTeacher(teacherID int64) (TeacherItem, error) {
	teacher, err := s.store.UpdateTeacherStatus(teacherID, "active")
	if err != nil {
		if errors.Is(err, memory.ErrTeacherNotFound) {
			return TeacherItem{}, ErrTeacherNotFound
		}
		return TeacherItem{}, err
	}

	return toTeacherItem(teacher), nil
}

func toTeacherItem(teacher memory.Teacher) TeacherItem {
	return TeacherItem{
		ID:        teacher.ID,
		Username:  teacher.Username,
		Role:      teacher.Role,
		Status:    teacher.Status,
		CreatedAt: teacher.CreatedAt.Format(time.RFC3339),
	}
}
