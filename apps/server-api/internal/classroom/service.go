package classroom

import (
	"errors"
	"strings"
	"time"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var (
	ErrClassroomNotFound = errors.New("classroom not found")
	ErrClassroomNotEmpty = errors.New("classroom is not empty")
)

type CreateInput struct {
	Name string `json:"name" binding:"required"`
}

type UpdateInput struct {
	Name string `json:"name" binding:"required"`
}

type Item struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	StudentCount int    `json:"studentCount"`
	ProjectCount int    `json:"projectCount"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type Detail struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	StudentCount int    `json:"studentCount"`
	ProjectCount int    `json:"projectCount"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(teacherID int64, input CreateInput) (Item, error) {
	record, err := s.store.CreateClassroom(teacherID, memory.CreateClassroomInput{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return Item{}, err
	}
	return s.toItem(record), nil
}

func (s *Service) List(teacherID int64) []Item {
	records := s.store.ListClassroomsByTeacher(teacherID)
	items := make([]Item, 0, len(records))
	for _, record := range records {
		items = append(items, s.toItem(record))
	}
	return items
}

func (s *Service) Get(teacherID int64, classroomID int64) (Detail, error) {
	record, ok := s.store.GetClassroomByTeacher(teacherID, classroomID)
	if !ok {
		return Detail{}, ErrClassroomNotFound
	}
	return Detail{
		ID:           record.ID,
		Name:         record.Name,
		StudentCount: s.store.CountStudentsByClassroom(classroomID),
		ProjectCount: s.store.CountAssignmentsByClassroom(classroomID),
		CreatedAt:    record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    record.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) Update(teacherID int64, classroomID int64, input UpdateInput) (Item, error) {
	record, err := s.store.UpdateClassroom(teacherID, classroomID, strings.TrimSpace(input.Name))
	if err != nil {
		if errors.Is(err, memory.ErrClassroomNotFound) {
			return Item{}, ErrClassroomNotFound
		}
		return Item{}, err
	}
	return s.toItem(record), nil
}

func (s *Service) Delete(teacherID int64, classroomID int64) error {
	err := s.store.DeleteClassroom(teacherID, classroomID)
	if errors.Is(err, memory.ErrClassroomNotFound) {
		return ErrClassroomNotFound
	}
	if errors.Is(err, memory.ErrClassroomNotEmpty) {
		return ErrClassroomNotEmpty
	}
	return err
}

func (s *Service) EnsureDefault(teacherID int64) (memory.Classroom, error) {
	return s.store.EnsureDefaultClassroom(teacherID)
}

func (s *Service) toItem(record memory.Classroom) Item {
	return Item{
		ID:           record.ID,
		Name:         record.Name,
		StudentCount: s.store.CountStudentsByClassroom(record.ID),
		ProjectCount: s.store.CountAssignmentsByClassroom(record.ID),
		CreatedAt:    record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    record.UpdatedAt.Format(time.RFC3339),
	}
}
