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
	ErrStudentConflict    = errors.New("student username already exists")
	ErrAssignmentNotFound = errors.New("assignment not found")
	ErrAssignmentNotReady = errors.New("assignment analysis not ready")
	ErrStudentNotFound    = errors.New("student not found")
)

type Teacher struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Student struct {
	ID           int64
	TeacherID    int64
	Username     string
	DisplayName  string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
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

type CreateStudentInput struct {
	Username     string
	DisplayName  string
	PasswordHash string
}

type CreateAssignmentInput struct {
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

type Store struct {
	sql *sqlBackend

	mu sync.RWMutex

	nextTeacherID    int64
	nextStudentID    int64
	nextAssignmentID int64
	nextProgressID   int64
	nextHintID       int64

	teachersByID       map[int64]Teacher
	teachersByUsername map[string]int64
	teacherTokens      map[string]int64

	studentsByID       map[int64]Student
	studentsByUsername map[string]int64
	studentsByTeacher  map[int64][]int64
	studentTokens      map[string]int64

	assignmentsByID      map[int64]Assignment
	assignmentsByTeacher map[int64][]int64
	assignmentsByStudent map[int64][]int64
	studentsByAssignment map[int64][]int64

	progressByID                map[int64]ProgressReport
	progressByStudentAssignment map[string][]int64
	hintsByID                   map[int64]HintRecord
	hintsByStudentAssignment    map[string][]int64
}

func NewStore(cfg config.Config) (*Store, error) {
	sqlBackend, err := newSQLBackend(cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		sql:                         sqlBackend,
		nextTeacherID:               1,
		nextStudentID:               1,
		nextAssignmentID:            1,
		nextProgressID:              1,
		nextHintID:                  1,
		teachersByID:                map[int64]Teacher{},
		teachersByUsername:          map[string]int64{},
		teacherTokens:               map[string]int64{},
		studentsByID:                map[int64]Student{},
		studentsByUsername:          map[string]int64{},
		studentsByTeacher:           map[int64][]int64{},
		studentTokens:               map[string]int64{},
		assignmentsByID:             map[int64]Assignment{},
		assignmentsByTeacher:        map[int64][]int64{},
		assignmentsByStudent:        map[int64][]int64{},
		studentsByAssignment:        map[int64][]int64{},
		progressByID:                map[int64]ProgressReport{},
		progressByStudentAssignment: map[string][]int64{},
		hintsByID:                   map[int64]HintRecord{},
		hintsByStudentAssignment:    map[string][]int64{},
	}, nil
}

func (s *Store) CreateTeacher(username string, passwordHash string) (Teacher, error) {
	if s.sql != nil {
		return s.sql.CreateTeacher(username, passwordHash)
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

func (s *Store) CreateStudent(teacherID int64, input CreateStudentInput) (Student, error) {
	if s.sql != nil {
		return s.sql.CreateStudent(teacherID, input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.studentsByUsername[input.Username]; exists {
		return Student{}, ErrStudentConflict
	}

	student := Student{
		ID:           s.nextStudentID,
		TeacherID:    teacherID,
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

func (s *Store) CreateAssignment(teacherID int64, input CreateAssignmentInput) (Assignment, error) {
	if s.sql != nil {
		return s.sql.CreateAssignment(teacherID, input)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	assignment := Assignment{
		ID:             s.nextAssignmentID,
		TeacherID:      teacherID,
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
	return assignment, nil
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
		if !ok || student.TeacherID != teacherID {
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
