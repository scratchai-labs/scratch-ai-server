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

type UpdateTeacherRoleInput struct {
	Role string `json:"role" binding:"required"`
}

type CreateStudentInput struct {
	TeacherID       int64  `json:"teacherId" binding:"required"`
	Username        string `json:"username" binding:"required"`
	DisplayName     string `json:"displayName" binding:"required"`
	InitialPassword string `json:"initialPassword" binding:"required"`
}

type ResetTeacherPasswordInput struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

type ResetStudentPasswordInput struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

type TeacherItem struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type StudentItem struct {
	ID              int64  `json:"id"`
	TeacherID       int64  `json:"teacherId"`
	TeacherUsername string `json:"teacherUsername"`
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	Status          string `json:"status"`
	CreatedAt       string `json:"createdAt"`
}

type Overview struct {
	AdminCount           int `json:"adminCount"`
	TeacherCount         int `json:"teacherCount"`
	ActiveTeacherCount   int `json:"activeTeacherCount"`
	DisabledTeacherCount int `json:"disabledTeacherCount"`
	StudentCount         int `json:"studentCount"`
	ActiveStudentCount   int `json:"activeStudentCount"`
	DisabledStudentCount int `json:"disabledStudentCount"`
}

type TeachersResponse struct {
	Items []TeacherItem `json:"items"`
}

type StudentsResponse struct {
	Items []StudentItem `json:"items"`
}

var (
	ErrTeacherConflict    = errors.New("teacher username already exists")
	ErrTeacherNotFound    = errors.New("teacher not found")
	ErrStudentConflict    = errors.New("student username already exists")
	ErrStudentNotFound    = errors.New("student not found")
	ErrSelfProtected      = errors.New("admin cannot disable self")
	ErrSelfRoleProtected  = errors.New("admin cannot change own role")
	ErrInvalidTeacherRole = errors.New("invalid teacher role")
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

func (s *Service) GetOverview() Overview {
	teachers := s.store.ListTeachers()
	students := s.store.ListStudents()

	overview := Overview{}
	for _, teacher := range teachers {
		switch teacher.Role {
		case "admin":
			overview.AdminCount++
		default:
			overview.TeacherCount++
			if teacher.Status == "disabled" {
				overview.DisabledTeacherCount++
			} else {
				overview.ActiveTeacherCount++
			}
		}
	}

	for _, student := range students {
		overview.StudentCount++
		if student.Status == "disabled" {
			overview.DisabledStudentCount++
		} else {
			overview.ActiveStudentCount++
		}
	}

	return overview
}

func (s *Service) ListStudents() []StudentItem {
	students := s.store.ListStudents()
	items := make([]StudentItem, 0, len(students))
	for _, student := range students {
		items = append(items, s.toStudentItem(student))
	}
	return items
}

func (s *Service) CreateStudent(input CreateStudentInput) (StudentItem, error) {
	teacher, ok := s.store.GetTeacherByID(input.TeacherID)
	if !ok || teacher.Role != "teacher" {
		return StudentItem{}, ErrTeacherNotFound
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.InitialPassword), bcrypt.DefaultCost)
	if err != nil {
		return StudentItem{}, err
	}

	student, err := s.store.CreateStudent(input.TeacherID, memory.CreateStudentInput{
		Username:     strings.TrimSpace(input.Username),
		DisplayName:  strings.TrimSpace(input.DisplayName),
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		if errors.Is(err, memory.ErrStudentConflict) {
			return StudentItem{}, ErrStudentConflict
		}
		return StudentItem{}, err
	}

	return s.toStudentItem(student), nil
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

func (s *Service) UpdateTeacherRole(adminID int64, teacherID int64, role string) (TeacherItem, error) {
	nextRole := strings.TrimSpace(role)
	if nextRole != "teacher" && nextRole != "admin" {
		return TeacherItem{}, ErrInvalidTeacherRole
	}
	if adminID == teacherID && nextRole != "admin" {
		return TeacherItem{}, ErrSelfRoleProtected
	}

	teacher, err := s.store.UpdateTeacherRole(teacherID, nextRole)
	if err != nil {
		if errors.Is(err, memory.ErrTeacherNotFound) {
			return TeacherItem{}, ErrTeacherNotFound
		}
		return TeacherItem{}, err
	}

	return toTeacherItem(teacher), nil
}

func (s *Service) ResetStudentPassword(studentID int64, newPassword string) (StudentItem, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return StudentItem{}, err
	}

	student, err := s.store.UpdateStudentPasswordByID(studentID, string(passwordHash))
	if err != nil {
		if errors.Is(err, memory.ErrStudentNotFound) {
			return StudentItem{}, ErrStudentNotFound
		}
		return StudentItem{}, err
	}

	return s.toStudentItem(student), nil
}

func (s *Service) DisableStudent(studentID int64) (StudentItem, error) {
	student, err := s.store.UpdateStudentStatus(studentID, "disabled")
	if err != nil {
		if errors.Is(err, memory.ErrStudentNotFound) {
			return StudentItem{}, ErrStudentNotFound
		}
		return StudentItem{}, err
	}

	return s.toStudentItem(student), nil
}

func (s *Service) EnableStudent(studentID int64) (StudentItem, error) {
	student, err := s.store.UpdateStudentStatus(studentID, "active")
	if err != nil {
		if errors.Is(err, memory.ErrStudentNotFound) {
			return StudentItem{}, ErrStudentNotFound
		}
		return StudentItem{}, err
	}

	return s.toStudentItem(student), nil
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

func (s *Service) toStudentItem(student memory.Student) StudentItem {
	teacherUsername := ""
	if teacher, ok := s.store.GetTeacherByID(student.TeacherID); ok {
		teacherUsername = teacher.Username
	}

	return StudentItem{
		ID:              student.ID,
		TeacherID:       student.TeacherID,
		TeacherUsername: teacherUsername,
		Username:        student.Username,
		DisplayName:     student.DisplayName,
		Status:          student.Status,
		CreatedAt:       student.CreatedAt.Format(time.RFC3339),
	}
}
