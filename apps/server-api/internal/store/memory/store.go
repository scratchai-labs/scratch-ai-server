package memory

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

var (
	ErrTeacherConflict    = errors.New("teacher username already exists")
	ErrTeacherNotFound    = errors.New("teacher not found")
	ErrClassroomNotFound  = errors.New("classroom not found")
	ErrClassroomNotEmpty  = errors.New("classroom is not empty")
	ErrStudentConflict    = errors.New("student username already exists")
	ErrAssignmentNotFound = errors.New("assignment not found")
	ErrAssignmentNotReady = errors.New("assignment analysis not ready")
	ErrStudentNotFound    = errors.New("student not found")
)

type Teacher struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    time.Time
}

type Student struct {
	ID           int64
	TeacherID    int64
	ClassroomID  int64
	Username     string
	DisplayName  string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
}

type Classroom struct {
	ID        int64
	TeacherID int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AssignmentAnalysis struct {
	RoleNames         []string
	ScriptCounts      map[string]int
	BlockCounts       map[string]int
	CategoryCounts    map[string]int
	BroadcastMessages []string
	VariableNames     []string
	ListNames         []string
	Extensions        []string
	TeachingPoints    []string
}

type Assignment struct {
	ID                   int64
	TeacherID            int64
	ClassroomID          int64
	Title                string
	Goal                 string
	Description          string
	Status               string
	FileName             string
	SB3FilePath          string
	SB3Data              []byte
	AnalysisStatus       string
	AnalysisErrorMessage string
	Analysis             AssignmentAnalysis
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ProgressReport struct {
	ID               int64
	AssignmentID     int64
	StudentID        int64
	CurrentTarget    string
	StepSummary      string
	LocalProjectHash string
	ReportedAt       string
	Snapshot         map[string]any
	CreatedAt        time.Time
}

type HintRecord struct {
	ID               int64
	AssignmentID     int64
	StudentID        int64
	ProgressReportID int64
	PromptInput      map[string]any
	HintText         string
	ProviderName     string
	CreatedAt        time.Time
}

type AuditLog struct {
	ID             int64
	ActorTeacherID int64
	ActorUsername  string
	Action         string
	TargetType     string
	TargetID       int64
	TargetUsername string
	BeforeState    map[string]string
	AfterState     map[string]string
	CreatedAt      time.Time
}

type CreateStudentInput struct {
	ClassroomID  int64
	Username     string
	DisplayName  string
	PasswordHash string
}

type CreateClassroomInput struct {
	Name string
}

type CreateAssignmentInput struct {
	ClassroomID int64
	Title       string
	Goal        string
	Description string
	FileName    string
	SB3FilePath string
	SB3Data     []byte
}

type CreateProgressInput struct {
	AssignmentID     int64
	StudentID        int64
	CurrentTarget    string
	StepSummary      string
	LocalProjectHash string
	ReportedAt       string
	Snapshot         map[string]any
}

type CreateHintInput struct {
	AssignmentID     int64
	StudentID        int64
	ProgressReportID int64
	PromptInput      map[string]any
	HintText         string
	ProviderName     string
}

type CreateAuditLogInput struct {
	ActorTeacherID int64
	ActorUsername  string
	Action         string
	TargetType     string
	TargetID       int64
	TargetUsername string
	BeforeState    map[string]string
	AfterState     map[string]string
}

type Store struct {
	sql *sqlBackend

	mu sync.RWMutex

	nextTeacherID    int64
	nextClassroomID  int64
	nextStudentID    int64
	nextAssignmentID int64
	nextProgressID   int64
	nextHintID       int64
	nextAuditLogID   int64

	teachersByID       map[int64]Teacher
	teachersByUsername map[string]int64
	teacherTokens      map[string]int64

	classroomsByID      map[int64]Classroom
	classroomsByTeacher map[int64][]int64

	studentsByID        map[int64]Student
	studentsByUsername  map[string]int64
	studentsByTeacher   map[int64][]int64
	studentsByClassroom map[int64][]int64
	studentTokens       map[string]int64

	assignmentsByID        map[int64]Assignment
	assignmentsByTeacher   map[int64][]int64
	assignmentsByClassroom map[int64][]int64
	assignmentsByStudent   map[int64][]int64
	studentsByAssignment   map[int64][]int64

	progressByID                map[int64]ProgressReport
	progressByStudentAssignment map[string][]int64
	hintsByID                   map[int64]HintRecord
	hintsByStudentAssignment    map[string][]int64
	auditLogs                   []AuditLog
}

func NewStore(cfg config.Config) (*Store, error) {
	sqlBackend, err := newSQLBackend(cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		sql:                         sqlBackend,
		nextTeacherID:               1,
		nextClassroomID:             1,
		nextStudentID:               1,
		nextAssignmentID:            1,
		nextProgressID:              1,
		nextHintID:                  1,
		nextAuditLogID:              1,
		teachersByID:                map[int64]Teacher{},
		teachersByUsername:          map[string]int64{},
		teacherTokens:               map[string]int64{},
		classroomsByID:              map[int64]Classroom{},
		classroomsByTeacher:         map[int64][]int64{},
		studentsByID:                map[int64]Student{},
		studentsByUsername:          map[string]int64{},
		studentsByTeacher:           map[int64][]int64{},
		studentsByClassroom:         map[int64][]int64{},
		studentTokens:               map[string]int64{},
		assignmentsByID:             map[int64]Assignment{},
		assignmentsByTeacher:        map[int64][]int64{},
		assignmentsByClassroom:      map[int64][]int64{},
		assignmentsByStudent:        map[int64][]int64{},
		studentsByAssignment:        map[int64][]int64{},
		progressByID:                map[int64]ProgressReport{},
		progressByStudentAssignment: map[string][]int64{},
		hintsByID:                   map[int64]HintRecord{},
		hintsByStudentAssignment:    map[string][]int64{},
		auditLogs:                   []AuditLog{},
	}, nil
}

func (s *Store) CreateTeacher(username string, passwordHash string) (Teacher, error) {
	return s.CreateTeacherWithRole(username, passwordHash, "teacher", "active")
}

func (s *Store) CreateTeacherWithRole(username string, passwordHash string, role string, status string) (Teacher, error) {
	if s.sql != nil {
		return s.sql.CreateTeacherWithRole(username, passwordHash, role, status)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.teachersByUsername[username]; exists {
		return Teacher{}, ErrTeacherConflict
	}

	teacher := Teacher{
		ID:           s.nextTeacherID,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       status,
		CreatedAt:    time.Now().UTC(),
	}
	s.nextTeacherID++

	s.teachersByID[teacher.ID] = teacher
	s.teachersByUsername[teacher.Username] = teacher.ID
	return teacher, nil
}

func (s *Store) EnsureTeacher(username string, passwordHash string, role string, status string) (Teacher, error) {
	if s.sql != nil {
		return s.sql.EnsureTeacher(username, passwordHash, role, status)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if id, exists := s.teachersByUsername[username]; exists {
		teacher := s.teachersByID[id]
		teacher.PasswordHash = passwordHash
		teacher.Role = role
		teacher.Status = status
		s.teachersByID[id] = teacher
		return teacher, nil
	}

	teacher := Teacher{
		ID:           s.nextTeacherID,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       status,
		CreatedAt:    time.Now().UTC(),
	}
	s.nextTeacherID++

	s.teachersByID[teacher.ID] = teacher
	s.teachersByUsername[teacher.Username] = teacher.ID
	return teacher, nil
}

func (s *Store) FindTeacherByUsername(username string) (Teacher, bool) {
	if s.sql != nil {
		return s.sql.FindTeacherByUsername(username)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.teachersByUsername[username]
	if !ok {
		return Teacher{}, false
	}

	teacher, ok := s.teachersByID[id]
	return teacher, ok
}

func (s *Store) SaveTeacherToken(token string, teacherID int64) error {
	if s.sql != nil {
		return s.sql.SaveTeacherToken(token, teacherID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.teacherTokens[token] = teacherID
	return nil
}

func (s *Store) DeleteTeacherToken(token string) error {
	if s.sql != nil {
		return s.sql.DeleteTeacherToken(token)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.teacherTokens, token)
	return nil
}

func (s *Store) FindTeacherByToken(token string) (Teacher, bool) {
	if s.sql != nil {
		return s.sql.FindTeacherByToken(token)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	teacherID, ok := s.teacherTokens[token]
	if !ok {
		return Teacher{}, false
	}

	teacher, ok := s.teachersByID[teacherID]
	return teacher, ok
}

func (s *Store) ListTeachers() []Teacher {
	if s.sql != nil {
		return s.sql.ListTeachers()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	teachers := make([]Teacher, 0, len(s.teachersByID))
	for _, teacher := range s.teachersByID {
		teachers = append(teachers, teacher)
	}
	slices.SortFunc(teachers, func(a Teacher, b Teacher) int {
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})
	return teachers
}

func (s *Store) GetTeacherByID(teacherID int64) (Teacher, bool) {
	if s.sql != nil {
		return s.sql.GetTeacherByID(teacherID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	teacher, ok := s.teachersByID[teacherID]
	return teacher, ok
}

func (s *Store) UpdateTeacherPassword(teacherID int64, passwordHash string) (Teacher, error) {
	if s.sql != nil {
		return s.sql.UpdateTeacherPassword(teacherID, passwordHash)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	teacher, ok := s.teachersByID[teacherID]
	if !ok {
		return Teacher{}, ErrTeacherNotFound
	}

	teacher.PasswordHash = passwordHash
	s.teachersByID[teacherID] = teacher
	return teacher, nil
}

func (s *Store) UpdateTeacherStatus(teacherID int64, status string) (Teacher, error) {
	if s.sql != nil {
		return s.sql.UpdateTeacherStatus(teacherID, status)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	teacher, ok := s.teachersByID[teacherID]
	if !ok {
		return Teacher{}, ErrTeacherNotFound
	}

	teacher.Status = status
	s.teachersByID[teacherID] = teacher
	return teacher, nil
}

func (s *Store) UpdateTeacherRole(teacherID int64, role string) (Teacher, error) {
	if s.sql != nil {
		return s.sql.UpdateTeacherRole(teacherID, role)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	teacher, ok := s.teachersByID[teacherID]
	if !ok {
		return Teacher{}, ErrTeacherNotFound
	}

	teacher.Role = role
	s.teachersByID[teacherID] = teacher
	return teacher, nil
}

func (s *Store) CreateAuditLog(input CreateAuditLogInput) (AuditLog, error) {
	if s.sql != nil {
		return s.sql.CreateAuditLog(input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record := AuditLog{
		ID:             s.nextAuditLogID,
		ActorTeacherID: input.ActorTeacherID,
		ActorUsername:  input.ActorUsername,
		Action:         input.Action,
		TargetType:     input.TargetType,
		TargetID:       input.TargetID,
		TargetUsername: input.TargetUsername,
		BeforeState:    cloneStringMap(input.BeforeState),
		AfterState:     cloneStringMap(input.AfterState),
		CreatedAt:      time.Now().UTC(),
	}
	s.nextAuditLogID++
	s.auditLogs = append(s.auditLogs, record)
	return record, nil
}

func (s *Store) ListAuditLogs() []AuditLog {
	if s.sql != nil {
		return s.sql.ListAuditLogs()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	records := make([]AuditLog, 0, len(s.auditLogs))
	for index := len(s.auditLogs) - 1; index >= 0; index-- {
		record := s.auditLogs[index]
		record.BeforeState = cloneStringMap(record.BeforeState)
		record.AfterState = cloneStringMap(record.AfterState)
		records = append(records, record)
	}
	return records
}

func (s *Store) CreateClassroom(teacherID int64, input CreateClassroomInput) (Classroom, error) {
	if s.sql != nil {
		return s.sql.CreateClassroom(teacherID, input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	record := Classroom{
		ID:        s.nextClassroomID,
		TeacherID: teacherID,
		Name:      input.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.nextClassroomID++

	s.classroomsByID[record.ID] = record
	s.classroomsByTeacher[teacherID] = append(s.classroomsByTeacher[teacherID], record.ID)
	return record, nil
}

func (s *Store) EnsureDefaultClassroom(teacherID int64) (Classroom, error) {
	if s.sql != nil {
		return s.sql.EnsureDefaultClassroom(teacherID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, classroomID := range s.classroomsByTeacher[teacherID] {
		record, ok := s.classroomsByID[classroomID]
		if ok && record.Name == "默认班级" {
			return record, nil
		}
	}

	now := time.Now().UTC()
	record := Classroom{
		ID:        s.nextClassroomID,
		TeacherID: teacherID,
		Name:      "默认班级",
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.nextClassroomID++
	s.classroomsByID[record.ID] = record
	s.classroomsByTeacher[teacherID] = append(s.classroomsByTeacher[teacherID], record.ID)
	return record, nil
}

func (s *Store) ListClassroomsByTeacher(teacherID int64) []Classroom {
	if s.sql != nil {
		return s.sql.ListClassroomsByTeacher(teacherID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := slices.Clone(s.classroomsByTeacher[teacherID])
	items := make([]Classroom, 0, len(ids))
	for _, id := range ids {
		if record, ok := s.classroomsByID[id]; ok {
			items = append(items, record)
		}
	}
	return items
}

func (s *Store) GetClassroomByTeacher(teacherID int64, classroomID int64) (Classroom, bool) {
	if s.sql != nil {
		return s.sql.GetClassroomByTeacher(teacherID, classroomID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.classroomsByID[classroomID]
	if !ok || record.TeacherID != teacherID {
		return Classroom{}, false
	}
	return record, true
}

func (s *Store) UpdateClassroom(teacherID int64, classroomID int64, name string) (Classroom, error) {
	if s.sql != nil {
		return s.sql.UpdateClassroom(teacherID, classroomID, name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.classroomsByID[classroomID]
	if !ok || record.TeacherID != teacherID {
		return Classroom{}, ErrClassroomNotFound
	}
	record.Name = name
	record.UpdatedAt = time.Now().UTC()
	s.classroomsByID[classroomID] = record
	return record, nil
}

func (s *Store) DeleteClassroom(teacherID int64, classroomID int64) error {
	if s.sql != nil {
		return s.sql.DeleteClassroom(teacherID, classroomID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.classroomsByID[classroomID]
	if !ok || record.TeacherID != teacherID {
		return ErrClassroomNotFound
	}
	if len(s.studentsByClassroom[classroomID]) > 0 || len(s.assignmentsByClassroom[classroomID]) > 0 {
		return ErrClassroomNotEmpty
	}

	delete(s.classroomsByID, classroomID)
	s.classroomsByTeacher[teacherID] = removeInt64(s.classroomsByTeacher[teacherID], classroomID)
	delete(s.studentsByClassroom, classroomID)
	delete(s.assignmentsByClassroom, classroomID)
	return nil
}

func (s *Store) CountStudentsByClassroom(classroomID int64) int {
	if s.sql != nil {
		return s.sql.CountStudentsByClassroom(classroomID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.studentsByClassroom[classroomID])
}

func (s *Store) CountAssignmentsByClassroom(classroomID int64) int {
	if s.sql != nil {
		return s.sql.CountAssignmentsByClassroom(classroomID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.assignmentsByClassroom[classroomID])
}

func (s *Store) CreateStudent(teacherID int64, input CreateStudentInput) (Student, error) {
	if s.sql != nil {
		if input.ClassroomID == 0 {
			classroom, err := s.sql.EnsureDefaultClassroom(teacherID)
			if err != nil {
				return Student{}, err
			}
			input.ClassroomID = classroom.ID
		}
		return s.sql.CreateStudent(teacherID, input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.studentsByUsername[input.Username]; exists {
		return Student{}, ErrStudentConflict
	}

	if input.ClassroomID == 0 {
		classroom, err := s.ensureDefaultClassroomLocked(teacherID)
		if err != nil {
			return Student{}, err
		}
		input.ClassroomID = classroom.ID
	}

	classroom, ok := s.classroomsByID[input.ClassroomID]
	if !ok || classroom.TeacherID != teacherID {
		return Student{}, ErrClassroomNotFound
	}

	student := Student{
		ID:           s.nextStudentID,
		TeacherID:    teacherID,
		ClassroomID:  input.ClassroomID,
		Username:     input.Username,
		DisplayName:  input.DisplayName,
		PasswordHash: input.PasswordHash,
		Status:       "active",
		CreatedAt:    time.Now().UTC(),
	}
	s.nextStudentID++

	s.studentsByID[student.ID] = student
	s.studentsByUsername[student.Username] = student.ID
	s.studentsByTeacher[teacherID] = append(s.studentsByTeacher[teacherID], student.ID)
	s.studentsByClassroom[input.ClassroomID] = append(s.studentsByClassroom[input.ClassroomID], student.ID)
	return student, nil
}

func (s *Store) ListStudentsByTeacher(teacherID int64) []Student {
	if s.sql != nil {
		return s.sql.ListStudentsByTeacher(teacherID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := slices.Clone(s.studentsByTeacher[teacherID])
	students := make([]Student, 0, len(ids))
	for _, id := range ids {
		if student, ok := s.studentsByID[id]; ok {
			students = append(students, student)
		}
	}
	return students
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

func (s *Store) ListStudents() []Student {
	if s.sql != nil {
		return s.sql.ListStudents()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	students := make([]Student, 0, len(s.studentsByID))
	for _, student := range s.studentsByID {
		students = append(students, student)
	}
	slices.SortFunc(students, func(a Student, b Student) int {
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})
	return students
}

func (s *Store) FindStudentByUsername(username string) (Student, bool) {
	if s.sql != nil {
		return s.sql.FindStudentByUsername(username)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.studentsByUsername[username]
	if !ok {
		return Student{}, false
	}

	student, ok := s.studentsByID[id]
	return student, ok
}

func (s *Store) SaveStudentToken(token string, studentID int64) error {
	if s.sql != nil {
		return s.sql.SaveStudentToken(token, studentID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.studentTokens[token] = studentID
	return nil
}

func (s *Store) DeleteStudentToken(token string) error {
	if s.sql != nil {
		return s.sql.DeleteStudentToken(token)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.studentTokens, token)
	return nil
}

func (s *Store) FindStudentByToken(token string) (Student, bool) {
	if s.sql != nil {
		return s.sql.FindStudentByToken(token)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	studentID, ok := s.studentTokens[token]
	if !ok {
		return Student{}, false
	}

	student, ok := s.studentsByID[studentID]
	return student, ok
}

func (s *Store) GetStudentByTeacher(teacherID int64, studentID int64) (Student, bool) {
	if s.sql != nil {
		return s.sql.GetStudentByTeacher(teacherID, studentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	student, ok := s.studentsByID[studentID]
	if !ok || student.TeacherID != teacherID {
		return Student{}, false
	}
	return student, true
}

func (s *Store) GetStudentByClassroom(teacherID int64, classroomID int64, studentID int64) (Student, bool) {
	if s.sql != nil {
		return s.sql.GetStudentByClassroom(teacherID, classroomID, studentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	student, ok := s.studentsByID[studentID]
	if !ok || student.TeacherID != teacherID || student.ClassroomID != classroomID {
		return Student{}, false
	}
	return student, true
}

func (s *Store) ListStudentsByClassroom(teacherID int64, classroomID int64) []Student {
	if s.sql != nil {
		return s.sql.ListStudentsByClassroom(teacherID, classroomID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	classroom, ok := s.classroomsByID[classroomID]
	if !ok || classroom.TeacherID != teacherID {
		return nil
	}

	ids := slices.Clone(s.studentsByClassroom[classroomID])
	students := make([]Student, 0, len(ids))
	for _, id := range ids {
		if student, ok := s.studentsByID[id]; ok {
			students = append(students, student)
		}
	}
	return students
}

func (s *Store) GetStudentByID(studentID int64) (Student, bool) {
	if s.sql != nil {
		return s.sql.GetStudentByID(studentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	student, ok := s.studentsByID[studentID]
	return student, ok
}

func (s *Store) UpdateStudentPassword(teacherID int64, studentID int64, passwordHash string) (Student, error) {
	if s.sql != nil {
		return s.sql.UpdateStudentPassword(teacherID, studentID, passwordHash)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	student, ok := s.studentsByID[studentID]
	if !ok || student.TeacherID != teacherID {
		return Student{}, ErrStudentNotFound
	}

	student.PasswordHash = passwordHash
	s.studentsByID[studentID] = student
	return student, nil
}

func (s *Store) UpdateStudent(teacherID int64, classroomID int64, studentID int64, username string, displayName string) (Student, error) {
	if s.sql != nil {
		return s.sql.UpdateStudent(teacherID, classroomID, studentID, username, displayName)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	student, ok := s.studentsByID[studentID]
	if !ok || student.TeacherID != teacherID || student.ClassroomID != classroomID {
		return Student{}, ErrStudentNotFound
	}
	if existingID, exists := s.studentsByUsername[username]; exists && existingID != studentID {
		return Student{}, ErrStudentConflict
	}

	if student.Username != username {
		delete(s.studentsByUsername, student.Username)
		s.studentsByUsername[username] = studentID
	}
	student.Username = username
	student.DisplayName = displayName
	s.studentsByID[studentID] = student
	return student, nil
}

func (s *Store) DeleteStudent(teacherID int64, classroomID int64, studentID int64) error {
	if s.sql != nil {
		return s.sql.DeleteStudent(teacherID, classroomID, studentID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	student, ok := s.studentsByID[studentID]
	if !ok || student.TeacherID != teacherID || student.ClassroomID != classroomID {
		return ErrStudentNotFound
	}

	delete(s.studentsByUsername, student.Username)
	delete(s.studentsByID, studentID)
	s.studentsByTeacher[teacherID] = removeInt64(s.studentsByTeacher[teacherID], studentID)
	s.studentsByClassroom[classroomID] = removeInt64(s.studentsByClassroom[classroomID], studentID)
	delete(s.assignmentsByStudent, studentID)
	for assignmentID, studentIDs := range s.studentsByAssignment {
		s.studentsByAssignment[assignmentID] = removeInt64(studentIDs, studentID)
	}
	return nil
}

func (s *Store) UpdateStudentPasswordByID(studentID int64, passwordHash string) (Student, error) {
	if s.sql != nil {
		return s.sql.UpdateStudentPasswordByID(studentID, passwordHash)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	student, ok := s.studentsByID[studentID]
	if !ok {
		return Student{}, ErrStudentNotFound
	}

	student.PasswordHash = passwordHash
	s.studentsByID[studentID] = student
	return student, nil
}

func (s *Store) UpdateStudentStatus(studentID int64, status string) (Student, error) {
	if s.sql != nil {
		return s.sql.UpdateStudentStatus(studentID, status)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	student, ok := s.studentsByID[studentID]
	if !ok {
		return Student{}, ErrStudentNotFound
	}

	student.Status = status
	s.studentsByID[studentID] = student
	return student, nil
}

func (s *Store) CreateAssignment(teacherID int64, input CreateAssignmentInput) (Assignment, error) {
	if s.sql != nil {
		if input.ClassroomID == 0 {
			classroom, err := s.sql.EnsureDefaultClassroom(teacherID)
			if err != nil {
				return Assignment{}, err
			}
			input.ClassroomID = classroom.ID
		}
		return s.sql.CreateAssignment(teacherID, input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if input.ClassroomID == 0 {
		classroom, err := s.ensureDefaultClassroomLocked(teacherID)
		if err != nil {
			return Assignment{}, err
		}
		input.ClassroomID = classroom.ID
	}

	classroom, ok := s.classroomsByID[input.ClassroomID]
	if !ok || classroom.TeacherID != teacherID {
		return Assignment{}, ErrClassroomNotFound
	}

	now := time.Now().UTC()
	assignment := Assignment{
		ID:             s.nextAssignmentID,
		TeacherID:      teacherID,
		ClassroomID:    input.ClassroomID,
		Title:          input.Title,
		Goal:           input.Goal,
		Description:    input.Description,
		Status:         "draft",
		FileName:       input.FileName,
		SB3FilePath:    input.SB3FilePath,
		SB3Data:        slices.Clone(input.SB3Data),
		AnalysisStatus: "pending",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.nextAssignmentID++

	s.assignmentsByID[assignment.ID] = assignment
	s.assignmentsByTeacher[teacherID] = append(s.assignmentsByTeacher[teacherID], assignment.ID)
	s.assignmentsByClassroom[input.ClassroomID] = append(s.assignmentsByClassroom[input.ClassroomID], assignment.ID)
	return assignment, nil
}

func (s *Store) ensureDefaultClassroomLocked(teacherID int64) (Classroom, error) {
	for _, classroomID := range s.classroomsByTeacher[teacherID] {
		record, ok := s.classroomsByID[classroomID]
		if ok && record.Name == "默认班级" {
			return record, nil
		}
	}

	now := time.Now().UTC()
	record := Classroom{
		ID:        s.nextClassroomID,
		TeacherID: teacherID,
		Name:      "默认班级",
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.nextClassroomID++
	s.classroomsByID[record.ID] = record
	s.classroomsByTeacher[teacherID] = append(s.classroomsByTeacher[teacherID], record.ID)
	return record, nil
}

func (s *Store) GetAssignmentByTeacher(teacherID int64, assignmentID int64) (Assignment, bool) {
	if s.sql != nil {
		return s.sql.GetAssignmentByTeacher(teacherID, assignmentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok || assignment.TeacherID != teacherID {
		return Assignment{}, false
	}
	return assignment, true
}

func (s *Store) ListAssignmentsByTeacher(teacherID int64) []Assignment {
	if s.sql != nil {
		return s.sql.ListAssignmentsByTeacher(teacherID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := slices.Clone(s.assignmentsByTeacher[teacherID])
	assignments := make([]Assignment, 0, len(ids))
	for _, id := range ids {
		if assignment, ok := s.assignmentsByID[id]; ok {
			assignments = append(assignments, assignment)
		}
	}
	return assignments
}

func (s *Store) ListAssignmentsByClassroom(teacherID int64, classroomID int64) []Assignment {
	if s.sql != nil {
		return s.sql.ListAssignmentsByClassroom(teacherID, classroomID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	classroom, ok := s.classroomsByID[classroomID]
	if !ok || classroom.TeacherID != teacherID {
		return nil
	}

	ids := slices.Clone(s.assignmentsByClassroom[classroomID])
	assignments := make([]Assignment, 0, len(ids))
	for _, id := range ids {
		if assignment, ok := s.assignmentsByID[id]; ok {
			assignments = append(assignments, assignment)
		}
	}
	return assignments
}

func (s *Store) ListAssignmentsPendingAnalysis() []Assignment {
	if s.sql != nil {
		return s.sql.ListAssignmentsPendingAnalysis()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	assignments := make([]Assignment, 0)
	for _, assignment := range s.assignmentsByID {
		if assignment.AnalysisStatus == "pending" || assignment.AnalysisStatus == "processing" {
			assignments = append(assignments, assignment)
		}
	}
	slices.SortFunc(assignments, func(a Assignment, b Assignment) int {
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		default:
			return 0
		}
	})
	return assignments
}

func (s *Store) SetAssignmentAnalysisProcessing(assignmentID int64) error {
	if s.sql != nil {
		return s.sql.SetAssignmentAnalysisProcessing(assignmentID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok {
		return nil
	}

	assignment.AnalysisStatus = "processing"
	assignment.UpdatedAt = time.Now().UTC()
	s.assignmentsByID[assignmentID] = assignment
	return nil
}

func (s *Store) SetAssignmentAnalysisReady(assignmentID int64, analysis AssignmentAnalysis) error {
	if s.sql != nil {
		return s.sql.SetAssignmentAnalysisReady(assignmentID, analysis)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok {
		return nil
	}

	assignment.AnalysisStatus = "ready"
	assignment.Analysis = analysis
	assignment.AnalysisErrorMessage = ""
	assignment.UpdatedAt = time.Now().UTC()
	s.assignmentsByID[assignmentID] = assignment
	return nil
}

func (s *Store) SetAssignmentAnalysisFailed(assignmentID int64, message string) error {
	if s.sql != nil {
		return s.sql.SetAssignmentAnalysisFailed(assignmentID, message)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok {
		return nil
	}

	assignment.AnalysisStatus = "failed"
	assignment.AnalysisErrorMessage = message
	assignment.UpdatedAt = time.Now().UTC()
	s.assignmentsByID[assignmentID] = assignment
	return nil
}

func (s *Store) AssignStudents(teacherID int64, assignmentID int64, studentIDs []int64) error {
	if s.sql != nil {
		return s.sql.AssignStudents(teacherID, assignmentID, studentIDs)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok || assignment.TeacherID != teacherID {
		return ErrAssignmentNotFound
	}

	for _, studentID := range studentIDs {
		student, ok := s.studentsByID[studentID]
		if !ok || student.TeacherID != teacherID || student.ClassroomID != assignment.ClassroomID {
			return ErrStudentNotFound
		}

		if !slices.Contains(s.studentsByAssignment[assignmentID], studentID) {
			s.studentsByAssignment[assignmentID] = append(s.studentsByAssignment[assignmentID], studentID)
		}
		if !slices.Contains(s.assignmentsByStudent[studentID], assignmentID) {
			s.assignmentsByStudent[studentID] = append(s.assignmentsByStudent[studentID], assignmentID)
		}
	}

	return nil
}

func (s *Store) PublishAssignment(teacherID int64, assignmentID int64) (Assignment, error) {
	if s.sql != nil {
		return s.sql.PublishAssignment(teacherID, assignmentID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok || assignment.TeacherID != teacherID {
		return Assignment{}, ErrAssignmentNotFound
	}
	if assignment.AnalysisStatus != "ready" {
		return Assignment{}, ErrAssignmentNotReady
	}

	assignment.Status = "published"
	assignment.UpdatedAt = time.Now().UTC()
	s.assignmentsByID[assignmentID] = assignment
	return assignment, nil
}

func (s *Store) ArchiveAssignment(teacherID int64, assignmentID int64) (Assignment, error) {
	if s.sql != nil {
		return s.sql.ArchiveAssignment(teacherID, assignmentID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok || assignment.TeacherID != teacherID {
		return Assignment{}, ErrAssignmentNotFound
	}

	assignment.Status = "archived"
	assignment.UpdatedAt = time.Now().UTC()
	s.assignmentsByID[assignmentID] = assignment
	return assignment, nil
}

func (s *Store) GetAssignmentForStudent(studentID int64, assignmentID int64) (Assignment, bool) {
	if s.sql != nil {
		return s.sql.GetAssignmentForStudent(studentID, assignmentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if !slices.Contains(s.assignmentsByStudent[studentID], assignmentID) {
		return Assignment{}, false
	}

	assignment, ok := s.assignmentsByID[assignmentID]
	if !ok {
		return Assignment{}, false
	}
	return assignment, true
}

func (s *Store) ListAssignmentsByStudent(studentID int64) []Assignment {
	if s.sql != nil {
		return s.sql.ListAssignmentsByStudent(studentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := slices.Clone(s.assignmentsByStudent[studentID])
	assignments := make([]Assignment, 0, len(ids))
	for _, id := range ids {
		assignment, ok := s.assignmentsByID[id]
		if !ok || assignment.Status != "published" {
			continue
		}
		assignments = append(assignments, assignment)
	}
	return assignments
}

func (s *Store) ListAssignedAssignmentsByStudent(studentID int64) []Assignment {
	if s.sql != nil {
		return s.sql.ListAssignedAssignmentsByStudent(studentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := slices.Clone(s.assignmentsByStudent[studentID])
	assignments := make([]Assignment, 0, len(ids))
	for _, id := range ids {
		if assignment, ok := s.assignmentsByID[id]; ok {
			assignments = append(assignments, assignment)
		}
	}
	return assignments
}

func (s *Store) CreateProgress(input CreateProgressInput) (ProgressReport, error) {
	if s.sql != nil {
		return s.sql.CreateProgress(input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	reportedAt := input.ReportedAt
	if reportedAt == "" {
		reportedAt = time.Now().UTC().Format(time.RFC3339)
	}

	report := ProgressReport{
		ID:               s.nextProgressID,
		AssignmentID:     input.AssignmentID,
		StudentID:        input.StudentID,
		CurrentTarget:    input.CurrentTarget,
		StepSummary:      input.StepSummary,
		LocalProjectHash: input.LocalProjectHash,
		ReportedAt:       reportedAt,
		Snapshot:         cloneMap(input.Snapshot),
		CreatedAt:        time.Now().UTC(),
	}
	s.nextProgressID++

	s.progressByID[report.ID] = report
	key := relationKey(input.StudentID, input.AssignmentID)
	s.progressByStudentAssignment[key] = append(s.progressByStudentAssignment[key], report.ID)
	return report, nil
}

func (s *Store) LatestProgress(studentID int64, assignmentID int64) (ProgressReport, bool) {
	if s.sql != nil {
		return s.sql.LatestProgress(studentID, assignmentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.progressByStudentAssignment[relationKey(studentID, assignmentID)]
	if len(ids) == 0 {
		return ProgressReport{}, false
	}

	report, ok := s.progressByID[ids[len(ids)-1]]
	return report, ok
}

func (s *Store) CreateHint(input CreateHintInput) (HintRecord, error) {
	if s.sql != nil {
		return s.sql.CreateHint(input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record := HintRecord{
		ID:               s.nextHintID,
		AssignmentID:     input.AssignmentID,
		StudentID:        input.StudentID,
		ProgressReportID: input.ProgressReportID,
		PromptInput:      cloneMap(input.PromptInput),
		HintText:         input.HintText,
		ProviderName:     input.ProviderName,
		CreatedAt:        time.Now().UTC(),
	}
	s.nextHintID++

	s.hintsByID[record.ID] = record
	key := relationKey(input.StudentID, input.AssignmentID)
	s.hintsByStudentAssignment[key] = append(s.hintsByStudentAssignment[key], record.ID)
	return record, nil
}

func (s *Store) Ping(ctx context.Context) error {
	if s.sql != nil {
		return s.sql.Ping(ctx)
	}
	return nil
}

func (s *Store) LatestHint(studentID int64, assignmentID int64) (HintRecord, bool) {
	if s.sql != nil {
		return s.sql.LatestHint(studentID, assignmentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := s.hintsByStudentAssignment[relationKey(studentID, assignmentID)]
	if len(ids) == 0 {
		return HintRecord{}, false
	}

	record, ok := s.hintsByID[ids[len(ids)-1]]
	return record, ok
}

func (s *Store) ListAssignedStudents(assignmentID int64) []Student {
	if s.sql != nil {
		return s.sql.ListAssignedStudents(assignmentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := slices.Clone(s.studentsByAssignment[assignmentID])
	students := make([]Student, 0, len(ids))
	for _, id := range ids {
		if student, ok := s.studentsByID[id]; ok {
			students = append(students, student)
		}
	}
	return students
}

func relationKey(studentID int64, assignmentID int64) string {
	return strconv.FormatInt(studentID, 10) + ":" + strconv.FormatInt(assignmentID, 10)
}

func removeInt64(items []int64, target int64) []int64 {
	if len(items) == 0 {
		return items
	}

	filtered := items[:0]
	for _, item := range items {
		if item != target {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}

	cloned := make(map[string]any, len(input))
	for key, value := range input {
		cloned[key] = cloneValue(value)
	}
	return cloned
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneMap(typed)
	case []any:
		cloned := make([]any, len(typed))
		for index, item := range typed {
			cloned[index] = cloneValue(item)
		}
		return cloned
	case []string:
		return slices.Clone(typed)
	default:
		return typed
	}
}
