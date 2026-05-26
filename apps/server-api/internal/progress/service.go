package progress

import (
	"errors"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var ErrAssignmentUnavailable = errors.New("assignment is not available to the student")

type ReportInput struct {
	CurrentTarget    string         `json:"currentTarget"`
	StepSummary      string         `json:"stepSummary"`
	LocalProjectHash string         `json:"localProjectHash"`
	ReportedAt       string         `json:"reportedAt"`
	Snapshot         map[string]any `json:"snapshot"`
}

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Report(studentID int64, assignmentID int64, input ReportInput) (memory.ProgressReport, error) {
	record, ok := s.store.GetAssignmentForStudent(studentID, assignmentID)
	if !ok || record.Status != "published" {
		return memory.ProgressReport{}, ErrAssignmentUnavailable
	}

	return s.store.CreateProgress(memory.CreateProgressInput{
		AssignmentID:     assignmentID,
		StudentID:        studentID,
		CurrentTarget:    input.CurrentTarget,
		StepSummary:      input.StepSummary,
		LocalProjectHash: input.LocalProjectHash,
		ReportedAt:       input.ReportedAt,
		Snapshot:         input.Snapshot,
	}), nil
}

func (s *Service) Latest(studentID int64, assignmentID int64) (memory.ProgressReport, bool) {
	return s.store.LatestProgress(studentID, assignmentID)
}
