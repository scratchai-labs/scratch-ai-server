package assignment

import (
	"context"
	"errors"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/sb3"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var ErrAssignmentNotFound = errors.New("assignment not found")
var ErrAssignmentNotReady = errors.New("assignment analysis not ready")
var ErrStudentNotFound = errors.New("student not found")
var ErrAssignmentUnavailable = errors.New("assignment is not available to the student")
var ErrInvalidSB3File = errors.New("invalid sb3 file")
var ErrInvalidSB3MIME = errors.New("unsupported sb3 mime type")
var ErrSB3TooLarge = errors.New("sb3 file is too large")

const MaxSB3Bytes = 16 << 20
const maxAnalysisAttempts = 3

type UploadInput struct {
	Title       string
	Goal        string
	Description string
	FileName    string
	ContentType string
	SB3Data     []byte
}

type StudentAssignmentItem struct {
	ID                int64          `json:"id"`
	Title             string         `json:"title"`
	Goal              string         `json:"goal"`
	Description       string         `json:"description"`
	Status            string         `json:"status"`
	AnalysisStatus    string         `json:"analysisStatus"`
	RoleNames         []string       `json:"roleNames"`
	ScriptCounts      map[string]int `json:"scriptCounts"`
	BlockCounts       map[string]int `json:"blockCounts"`
	CategoryCounts    map[string]int `json:"categoryCounts"`
	BroadcastMessages []string       `json:"broadcastMessages"`
	VariableNames     []string       `json:"variableNames"`
	ListNames         []string       `json:"listNames"`
	Extensions        []string       `json:"extensions"`
	TeachingPoints    []string       `json:"teachingPoints"`
}

type TeacherAssignmentItem struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Goal           string `json:"goal"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	AnalysisStatus string `json:"analysisStatus"`
	StudentCount   int    `json:"studentCount"`
	UpdatedAt      string `json:"updatedAt"`
}

type AssignmentStudentItem struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
}

type AssignmentProgressSummary struct {
	CurrentTarget   string `json:"currentTarget"`
	StepSummary     string `json:"stepSummary"`
	CurrentRoleName string `json:"currentRoleName,omitempty"`
	ReportedAt      string `json:"reportedAt"`
}

type AssignmentHintSummary struct {
	HintText     string `json:"hintText"`
	ProviderName string `json:"providerName"`
	CreatedAt    string `json:"createdAt"`
}

type TeacherAssignmentDetail struct {
	ID                int64                   `json:"id"`
	Title             string                  `json:"title"`
	Goal              string                  `json:"goal"`
	Description       string                  `json:"description"`
	Status            string                  `json:"status"`
	AnalysisStatus    string                  `json:"analysisStatus"`
	RoleNames         []string                `json:"roleNames"`
	ScriptCounts      map[string]int          `json:"scriptCounts"`
	BlockCounts       map[string]int          `json:"blockCounts"`
	CategoryCounts    map[string]int          `json:"categoryCounts"`
	BroadcastMessages []string                `json:"broadcastMessages"`
	VariableNames     []string                `json:"variableNames"`
	ListNames         []string                `json:"listNames"`
	Extensions        []string                `json:"extensions"`
	TeachingPoints    []string                `json:"teachingPoints"`
	AssignedStudents  []AssignmentStudentItem `json:"assignedStudents"`
	UpdatedAt         string                  `json:"updatedAt"`
}

type StudentAssignmentDetail struct {
	ID                int64                      `json:"id"`
	Title             string                     `json:"title"`
	Goal              string                     `json:"goal"`
	Description       string                     `json:"description"`
	Status            string                     `json:"status"`
	AnalysisStatus    string                     `json:"analysisStatus"`
	RoleNames         []string                   `json:"roleNames"`
	ScriptCounts      map[string]int             `json:"scriptCounts"`
	BlockCounts       map[string]int             `json:"blockCounts"`
	CategoryCounts    map[string]int             `json:"categoryCounts"`
	BroadcastMessages []string                   `json:"broadcastMessages"`
	VariableNames     []string                   `json:"variableNames"`
	ListNames         []string                   `json:"listNames"`
	Extensions        []string                   `json:"extensions"`
	TeachingPoints    []string                   `json:"teachingPoints"`
	LatestProgress    *AssignmentProgressSummary `json:"latestProgress,omitempty"`
	LatestHint        *AssignmentHintSummary     `json:"latestHint,omitempty"`
}

type Service struct {
	store    *memory.Store
	analyzer *sb3.Analyzer
	storage  sb3.Storage
}

func NewService(store *memory.Store, analyzer *sb3.Analyzer, storage sb3.Storage) *Service {
	service := &Service{
		store:    store,
		analyzer: analyzer,
		storage:  storage,
	}
	go service.resumePendingAnalyses(context.Background())
	return service
}

func (s *Service) Upload(ctx context.Context, teacherID int64, input UploadInput) (memory.Assignment, error) {
	if err := validateUpload(input); err != nil {
		return memory.Assignment{}, err
	}

	sb3FilePath, err := s.storage.Save(ctx, input.FileName, input.SB3Data)
	if err != nil {
		return memory.Assignment{}, err
	}

	assignment, err := s.store.CreateAssignment(teacherID, memory.CreateAssignmentInput{
		Title:       input.Title,
		Goal:        input.Goal,
		Description: input.Description,
		FileName:    input.FileName,
		SB3FilePath: sb3FilePath,
		SB3Data:     input.SB3Data,
	})
	if err != nil {
		return memory.Assignment{}, err
	}

	go s.runAnalysisWithRetry(assignment.ID, func(context.Context) ([]byte, error) {
		return input.SB3Data, nil
	})
	return assignment, nil
}

func (s *Service) GetAnalysis(teacherID int64, assignmentID int64) (memory.Assignment, error) {
	assignment, ok := s.store.GetAssignmentByTeacher(teacherID, assignmentID)
	if !ok {
		return memory.Assignment{}, ErrAssignmentNotFound
	}
	return assignment, nil
}

func (s *Service) GetForTeacher(teacherID int64, assignmentID int64) (TeacherAssignmentDetail, error) {
	assignmentRecord, ok := s.store.GetAssignmentByTeacher(teacherID, assignmentID)
	if !ok {
		return TeacherAssignmentDetail{}, ErrAssignmentNotFound
	}

	students := s.store.ListAssignedStudents(assignmentID)
	assignedStudents := make([]AssignmentStudentItem, 0, len(students))
	for _, studentRecord := range students {
		assignedStudents = append(assignedStudents, AssignmentStudentItem{
			ID:          studentRecord.ID,
			Username:    studentRecord.Username,
			DisplayName: studentRecord.DisplayName,
			Status:      studentRecord.Status,
		})
	}

	return TeacherAssignmentDetail{
		ID:                assignmentRecord.ID,
		Title:             assignmentRecord.Title,
		Goal:              assignmentRecord.Goal,
		Description:       assignmentRecord.Description,
		Status:            assignmentRecord.Status,
		AnalysisStatus:    assignmentRecord.AnalysisStatus,
		RoleNames:         assignmentRecord.Analysis.RoleNames,
		ScriptCounts:      assignmentRecord.Analysis.ScriptCounts,
		BlockCounts:       assignmentRecord.Analysis.BlockCounts,
		CategoryCounts:    assignmentRecord.Analysis.CategoryCounts,
		BroadcastMessages: assignmentRecord.Analysis.BroadcastMessages,
		VariableNames:     assignmentRecord.Analysis.VariableNames,
		ListNames:         assignmentRecord.Analysis.ListNames,
		Extensions:        assignmentRecord.Analysis.Extensions,
		TeachingPoints:    assignmentRecord.Analysis.TeachingPoints,
		AssignedStudents:  assignedStudents,
		UpdatedAt:         assignmentRecord.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *Service) AssignStudents(teacherID int64, assignmentID int64, studentIDs []int64) error {
	err := s.store.AssignStudents(teacherID, assignmentID, studentIDs)
	if errors.Is(err, memory.ErrAssignmentNotFound) {
		return ErrAssignmentNotFound
	}
	if errors.Is(err, memory.ErrStudentNotFound) {
		return ErrStudentNotFound
	}
	return err
}

func (s *Service) Publish(teacherID int64, assignmentID int64) (memory.Assignment, error) {
	record, err := s.store.PublishAssignment(teacherID, assignmentID)
	if errors.Is(err, memory.ErrAssignmentNotFound) {
		return memory.Assignment{}, ErrAssignmentNotFound
	}
	if errors.Is(err, memory.ErrAssignmentNotReady) {
		return memory.Assignment{}, ErrAssignmentNotReady
	}
	return record, err
}

func (s *Service) Archive(teacherID int64, assignmentID int64) (memory.Assignment, error) {
	record, err := s.store.ArchiveAssignment(teacherID, assignmentID)
	if errors.Is(err, memory.ErrAssignmentNotFound) {
		return memory.Assignment{}, ErrAssignmentNotFound
	}
	return record, err
}

func (s *Service) ListForStudent(studentID int64) []StudentAssignmentItem {
	assignments := s.store.ListAssignmentsByStudent(studentID)
	items := make([]StudentAssignmentItem, 0, len(assignments))
	for _, assignment := range assignments {
		items = append(items, toStudentAssignmentItem(assignment))
	}
	return items
}

func (s *Service) ListForTeacher(teacherID int64) []TeacherAssignmentItem {
	assignments := s.store.ListAssignmentsByTeacher(teacherID)
	items := make([]TeacherAssignmentItem, 0, len(assignments))
	for _, assignmentRecord := range assignments {
		items = append(items, TeacherAssignmentItem{
			ID:             assignmentRecord.ID,
			Title:          assignmentRecord.Title,
			Goal:           assignmentRecord.Goal,
			Description:    assignmentRecord.Description,
			Status:         assignmentRecord.Status,
			AnalysisStatus: assignmentRecord.AnalysisStatus,
			StudentCount:   len(s.store.ListAssignedStudents(assignmentRecord.ID)),
			UpdatedAt:      assignmentRecord.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return items
}

func (s *Service) GetForStudent(studentID int64, assignmentID int64) (memory.Assignment, error) {
	record, ok := s.store.GetAssignmentForStudent(studentID, assignmentID)
	if !ok || record.Status != "published" {
		return memory.Assignment{}, ErrAssignmentUnavailable
	}
	return record, nil
}

func (s *Service) GetDetailForStudent(studentID int64, assignmentID int64) (StudentAssignmentDetail, error) {
	record, err := s.GetForStudent(studentID, assignmentID)
	if err != nil {
		return StudentAssignmentDetail{}, err
	}

	var latestProgress *AssignmentProgressSummary
	if progressRecord, ok := s.store.LatestProgress(studentID, assignmentID); ok {
		latestProgress = &AssignmentProgressSummary{
			CurrentTarget:   progressRecord.CurrentTarget,
			StepSummary:     progressRecord.StepSummary,
			CurrentRoleName: snapshotString(progressRecord.Snapshot, "currentRoleName"),
			ReportedAt:      progressRecord.ReportedAt,
		}
	}

	var latestHint *AssignmentHintSummary
	if hintRecord, ok := s.store.LatestHint(studentID, assignmentID); ok {
		latestHint = &AssignmentHintSummary{
			HintText:     hintRecord.HintText,
			ProviderName: hintRecord.ProviderName,
			CreatedAt:    hintRecord.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return StudentAssignmentDetail{
		ID:                record.ID,
		Title:             record.Title,
		Goal:              record.Goal,
		Description:       record.Description,
		Status:            record.Status,
		AnalysisStatus:    record.AnalysisStatus,
		RoleNames:         record.Analysis.RoleNames,
		ScriptCounts:      record.Analysis.ScriptCounts,
		BlockCounts:       record.Analysis.BlockCounts,
		CategoryCounts:    record.Analysis.CategoryCounts,
		BroadcastMessages: record.Analysis.BroadcastMessages,
		VariableNames:     record.Analysis.VariableNames,
		ListNames:         record.Analysis.ListNames,
		Extensions:        record.Analysis.Extensions,
		TeachingPoints:    record.Analysis.TeachingPoints,
		LatestProgress:    latestProgress,
		LatestHint:        latestHint,
	}, nil
}

func (s *Service) resumePendingAnalyses(ctx context.Context) {
	assignments := s.store.ListAssignmentsPendingAnalysis()
	for _, assignmentRecord := range assignments {
		assignmentRecord := assignmentRecord
		if strings.TrimSpace(assignmentRecord.SB3FilePath) == "" {
			continue
		}

		go s.runAnalysisWithRetry(assignmentRecord.ID, func(ctx context.Context) ([]byte, error) {
			return s.storage.Read(ctx, assignmentRecord.SB3FilePath)
		})
	}
}

func (s *Service) runAnalysisWithRetry(assignmentID int64, loadSB3 func(context.Context) ([]byte, error)) {
	if err := s.store.SetAssignmentAnalysisProcessing(assignmentID); err != nil {
		return
	}

	var lastErr error
	for attempt := 1; attempt <= maxAnalysisAttempts; attempt++ {
		rawSB3, err := loadSB3(context.Background())
		if err == nil {
			analysis, analyzeErr := s.analyzer.Analyze(rawSB3)
			if analyzeErr == nil {
				if readyErr := s.store.SetAssignmentAnalysisReady(assignmentID, analysis); readyErr == nil {
					return
				} else {
					lastErr = readyErr
				}
			} else {
				lastErr = analyzeErr
			}
		} else {
			lastErr = err
		}

		if attempt < maxAnalysisAttempts {
			time.Sleep(time.Duration(attempt) * 25 * time.Millisecond)
		}
	}

	if lastErr != nil {
		_ = s.store.SetAssignmentAnalysisFailed(assignmentID, lastErr.Error())
	}
}

func toStudentAssignmentItem(record memory.Assignment) StudentAssignmentItem {
	return StudentAssignmentItem{
		ID:                record.ID,
		Title:             record.Title,
		Goal:              record.Goal,
		Description:       record.Description,
		Status:            record.Status,
		AnalysisStatus:    record.AnalysisStatus,
		RoleNames:         record.Analysis.RoleNames,
		ScriptCounts:      record.Analysis.ScriptCounts,
		BlockCounts:       record.Analysis.BlockCounts,
		CategoryCounts:    record.Analysis.CategoryCounts,
		BroadcastMessages: record.Analysis.BroadcastMessages,
		VariableNames:     record.Analysis.VariableNames,
		ListNames:         record.Analysis.ListNames,
		Extensions:        record.Analysis.Extensions,
		TeachingPoints:    record.Analysis.TeachingPoints,
	}
}

func snapshotString(snapshot map[string]any, key string) string {
	value, ok := snapshot[key].(string)
	if !ok {
		return ""
	}
	return value
}

func validateUpload(input UploadInput) error {
	if strings.ToLower(filepath.Ext(strings.TrimSpace(input.FileName))) != ".sb3" {
		return ErrInvalidSB3File
	}
	if !isAllowedSB3MIME(input.ContentType) {
		return ErrInvalidSB3MIME
	}
	if len(input.SB3Data) == 0 {
		return ErrInvalidSB3File
	}
	if len(input.SB3Data) > MaxSB3Bytes {
		return ErrSB3TooLarge
	}
	return nil
}

func isAllowedSB3MIME(raw string) bool {
	contentType := strings.TrimSpace(raw)
	if contentType == "" {
		return true
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	switch mediaType {
	case "application/zip", "application/x-zip-compressed", "application/octet-stream":
		return true
	default:
		return false
	}
}
