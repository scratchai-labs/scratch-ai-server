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

type AuditLogItem struct {
	ID             int64             `json:"id"`
	ActorUsername  string            `json:"actorUsername"`
	Action         string            `json:"action"`
	TargetType     string            `json:"targetType"`
	TargetID       int64             `json:"targetId"`
	TargetUsername string            `json:"targetUsername"`
	Before         map[string]string `json:"before"`
	After          map[string]string `json:"after"`
	CreatedAt      string            `json:"createdAt"`
}

type AuditLogsResponse struct {
	Items []AuditLogItem `json:"items"`
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

func (s *Service) ListAuditLogs() []AuditLogItem {
	records := s.store.ListAuditLogs()
	items := make([]AuditLogItem, 0, len(records))
	for _, record := range records {
		items = append(items, toAuditLogItem(record))
	}
	return items
}

func (s *Service) CreateStudent(actor memory.Teacher, input CreateStudentInput) (StudentItem, error) {
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

	createdStudent := s.toStudentItem(student)
	if err := s.recordAuditLog(actor, "student.create", "student", student.ID, student.Username, map[string]string{}, studentSnapshot(createdStudent)); err != nil {
		return StudentItem{}, err
	}

	return createdStudent, nil
}

func (s *Service) CreateTeacher(actor memory.Teacher, input CreateTeacherInput) (TeacherItem, error) {
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

	createdTeacher := toTeacherItem(teacher)
	if err := s.recordAuditLog(actor, "teacher.create", "teacher", teacher.ID, teacher.Username, map[string]string{}, teacherSnapshot(createdTeacher)); err != nil {
		return TeacherItem{}, err
	}

	return createdTeacher, nil
}

func (s *Service) ResetTeacherPassword(actor memory.Teacher, teacherID int64, newPassword string) (TeacherItem, error) {
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

	updatedTeacher := toTeacherItem(teacher)
	if err := s.recordAuditLog(actor, "teacher.password_reset", "teacher", teacher.ID, teacher.Username, map[string]string{}, map[string]string{
		"passwordStatus": "updated",
	}); err != nil {
		return TeacherItem{}, err
	}

	return updatedTeacher, nil
}

func (s *Service) DisableTeacher(actor memory.Teacher, teacherID int64) (TeacherItem, error) {
	if actor.ID == teacherID {
		return TeacherItem{}, ErrSelfProtected
	}
	beforeTeacher, ok := s.store.GetTeacherByID(teacherID)
	if !ok {
		return TeacherItem{}, ErrTeacherNotFound
	}

	teacher, err := s.store.UpdateTeacherStatus(teacherID, "disabled")
	if err != nil {
		if errors.Is(err, memory.ErrTeacherNotFound) {
			return TeacherItem{}, ErrTeacherNotFound
		}
		return TeacherItem{}, err
	}

	updatedTeacher := toTeacherItem(teacher)
	if err := s.recordAuditLog(actor, "teacher.disable", "teacher", teacher.ID, teacher.Username, map[string]string{
		"status": beforeTeacher.Status,
	}, map[string]string{
		"status": teacher.Status,
	}); err != nil {
		return TeacherItem{}, err
	}

	return updatedTeacher, nil
}

func (s *Service) EnableTeacher(actor memory.Teacher, teacherID int64) (TeacherItem, error) {
	beforeTeacher, ok := s.store.GetTeacherByID(teacherID)
	if !ok {
		return TeacherItem{}, ErrTeacherNotFound
	}

	teacher, err := s.store.UpdateTeacherStatus(teacherID, "active")
	if err != nil {
		if errors.Is(err, memory.ErrTeacherNotFound) {
			return TeacherItem{}, ErrTeacherNotFound
		}
		return TeacherItem{}, err
	}

	updatedTeacher := toTeacherItem(teacher)
	if err := s.recordAuditLog(actor, "teacher.enable", "teacher", teacher.ID, teacher.Username, map[string]string{
		"status": beforeTeacher.Status,
	}, map[string]string{
		"status": teacher.Status,
	}); err != nil {
		return TeacherItem{}, err
	}

	return updatedTeacher, nil
}

func (s *Service) UpdateTeacherRole(actor memory.Teacher, teacherID int64, role string) (TeacherItem, error) {
	nextRole := strings.TrimSpace(role)
	if nextRole != "teacher" && nextRole != "admin" {
		return TeacherItem{}, ErrInvalidTeacherRole
	}
	if actor.ID == teacherID && nextRole != "admin" {
		return TeacherItem{}, ErrSelfRoleProtected
	}
	beforeTeacher, ok := s.store.GetTeacherByID(teacherID)
	if !ok {
		return TeacherItem{}, ErrTeacherNotFound
	}

	teacher, err := s.store.UpdateTeacherRole(teacherID, nextRole)
	if err != nil {
		if errors.Is(err, memory.ErrTeacherNotFound) {
			return TeacherItem{}, ErrTeacherNotFound
		}
		return TeacherItem{}, err
	}

	updatedTeacher := toTeacherItem(teacher)
	if err := s.recordAuditLog(actor, "teacher.role_change", "teacher", teacher.ID, teacher.Username, map[string]string{
		"role": beforeTeacher.Role,
	}, map[string]string{
		"role": teacher.Role,
	}); err != nil {
		return TeacherItem{}, err
	}

	return updatedTeacher, nil
}

func (s *Service) ResetStudentPassword(actor memory.Teacher, studentID int64, newPassword string) (StudentItem, error) {
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

	updatedStudent := s.toStudentItem(student)
	if err := s.recordAuditLog(actor, "student.password_reset", "student", student.ID, student.Username, map[string]string{}, map[string]string{
		"passwordStatus": "updated",
	}); err != nil {
		return StudentItem{}, err
	}

	return updatedStudent, nil
}

func (s *Service) DisableStudent(actor memory.Teacher, studentID int64) (StudentItem, error) {
	beforeStudent, ok := s.store.GetStudentByID(studentID)
	if !ok {
		return StudentItem{}, ErrStudentNotFound
	}

	student, err := s.store.UpdateStudentStatus(studentID, "disabled")
	if err != nil {
		if errors.Is(err, memory.ErrStudentNotFound) {
			return StudentItem{}, ErrStudentNotFound
		}
		return StudentItem{}, err
	}

	updatedStudent := s.toStudentItem(student)
	if err := s.recordAuditLog(actor, "student.disable", "student", student.ID, student.Username, map[string]string{
		"status": beforeStudent.Status,
	}, map[string]string{
		"status": student.Status,
	}); err != nil {
		return StudentItem{}, err
	}

	return updatedStudent, nil
}

func (s *Service) EnableStudent(actor memory.Teacher, studentID int64) (StudentItem, error) {
	beforeStudent, ok := s.store.GetStudentByID(studentID)
	if !ok {
		return StudentItem{}, ErrStudentNotFound
	}

	student, err := s.store.UpdateStudentStatus(studentID, "active")
	if err != nil {
		if errors.Is(err, memory.ErrStudentNotFound) {
			return StudentItem{}, ErrStudentNotFound
		}
		return StudentItem{}, err
	}

	updatedStudent := s.toStudentItem(student)
	if err := s.recordAuditLog(actor, "student.enable", "student", student.ID, student.Username, map[string]string{
		"status": beforeStudent.Status,
	}, map[string]string{
		"status": student.Status,
	}); err != nil {
		return StudentItem{}, err
	}

	return updatedStudent, nil
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

func toAuditLogItem(record memory.AuditLog) AuditLogItem {
	return AuditLogItem{
		ID:             record.ID,
		ActorUsername:  record.ActorUsername,
		Action:         record.Action,
		TargetType:     record.TargetType,
		TargetID:       record.TargetID,
		TargetUsername: record.TargetUsername,
		Before:         cloneStringMap(record.BeforeState),
		After:          cloneStringMap(record.AfterState),
		CreatedAt:      record.CreatedAt.Format(time.RFC3339),
	}
}

func (s *Service) recordAuditLog(
	actor memory.Teacher,
	action string,
	targetType string,
	targetID int64,
	targetUsername string,
	before map[string]string,
	after map[string]string,
) error {
	_, err := s.store.CreateAuditLog(memory.CreateAuditLogInput{
		ActorTeacherID: actor.ID,
		ActorUsername:  actor.Username,
		Action:         action,
		TargetType:     targetType,
		TargetID:       targetID,
		TargetUsername: targetUsername,
		BeforeState:    before,
		AfterState:     after,
	})
	return err
}

func teacherSnapshot(teacher TeacherItem) map[string]string {
	return map[string]string{
		"username": teacher.Username,
		"role":     teacher.Role,
		"status":   teacher.Status,
	}
}

func studentSnapshot(student StudentItem) map[string]string {
	return map[string]string{
		"username":        student.Username,
		"displayName":     student.DisplayName,
		"status":          student.Status,
		"teacherUsername": student.TeacherUsername,
	}
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}
