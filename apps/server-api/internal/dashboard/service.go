package dashboard

import (
	"errors"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var ErrAssignmentNotFound = errors.New("assignment not found")
var ErrStudentNotFound = errors.New("student not found")

type StudentLiveItem struct {
	StudentID       int64  `json:"studentId"`
	StudentName     string `json:"studentName"`
	Status          string `json:"status"`
	CurrentTarget   string `json:"currentTarget,omitempty"`
	StepSummary     string `json:"stepSummary,omitempty"`
	CurrentRoleName string `json:"currentRoleName,omitempty"`
	LastReportedAt  string `json:"lastReportedAt,omitempty"`
	LastHintText    string `json:"lastHintText,omitempty"`
	LastHintAt      string `json:"lastHintAt,omitempty"`
}

type LiveAssignmentResponse struct {
	AssignmentID    int64             `json:"assignmentId"`
	AssignmentTitle string            `json:"assignmentTitle"`
	UpdatedAt       string            `json:"updatedAt"`
	Students        []StudentLiveItem `json:"students"`
}

type StudentHistoryItem struct {
	AssignmentID     int64  `json:"assignmentId"`
	AssignmentTitle  string `json:"assignmentTitle"`
	AssignmentStatus string `json:"assignmentStatus"`
	CurrentTarget    string `json:"currentTarget,omitempty"`
	StepSummary      string `json:"stepSummary,omitempty"`
	CurrentRoleName  string `json:"currentRoleName,omitempty"`
	ReportedAt       string `json:"reportedAt,omitempty"`
	HintText         string `json:"hintText,omitempty"`
	HintProvider     string `json:"hintProvider,omitempty"`
	HintCreatedAt    string `json:"hintCreatedAt,omitempty"`
}

type StudentHistoryResponse struct {
	StudentID   int64                `json:"studentId"`
	StudentName string               `json:"studentName"`
	Items       []StudentHistoryItem `json:"items"`
}

type Service struct {
	store *memory.Store
}

func NewService(store *memory.Store) *Service {
	return &Service{store: store}
}

func (s *Service) LiveAssignment(teacherID int64, assignmentID int64) (LiveAssignmentResponse, error) {
	assignmentRecord, ok := s.store.GetAssignmentByTeacher(teacherID, assignmentID)
	if !ok {
		return LiveAssignmentResponse{}, ErrAssignmentNotFound
	}

	students := s.store.ListAssignedStudents(assignmentID)
	items := make([]StudentLiveItem, 0, len(students))
	for _, studentRecord := range students {
		item := StudentLiveItem{
			StudentID:   studentRecord.ID,
			StudentName: studentRecord.DisplayName,
			Status:      "assigned",
		}

		if progressRecord, ok := s.store.LatestProgress(studentRecord.ID, assignmentID); ok {
			item.Status = "active"
			item.CurrentTarget = progressRecord.CurrentTarget
			item.StepSummary = progressRecord.StepSummary
			item.CurrentRoleName = snapshotString(progressRecord.Snapshot, "currentRoleName")
			item.LastReportedAt = progressRecord.ReportedAt
		}

		if hintRecord, ok := s.store.LatestHint(studentRecord.ID, assignmentID); ok {
			item.LastHintText = hintRecord.HintText
			item.LastHintAt = hintRecord.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		items = append(items, item)
	}

	return LiveAssignmentResponse{
		AssignmentID:    assignmentID,
		AssignmentTitle: assignmentRecord.Title,
		UpdatedAt:       assignmentRecord.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Students:        items,
	}, nil
}

func (s *Service) StudentHistory(teacherID int64, studentID int64) (StudentHistoryResponse, error) {
	studentRecord, ok := s.store.GetStudentByTeacher(teacherID, studentID)
	if !ok {
		return StudentHistoryResponse{}, ErrStudentNotFound
	}

	assignments := s.store.ListAssignedAssignmentsByStudent(studentID)
	items := make([]StudentHistoryItem, 0, len(assignments))
	for _, assignmentRecord := range assignments {
		item := StudentHistoryItem{
			AssignmentID:     assignmentRecord.ID,
			AssignmentTitle:  assignmentRecord.Title,
			AssignmentStatus: assignmentRecord.Status,
		}

		if progressRecord, ok := s.store.LatestProgress(studentID, assignmentRecord.ID); ok {
			item.CurrentTarget = progressRecord.CurrentTarget
			item.StepSummary = progressRecord.StepSummary
			item.CurrentRoleName = snapshotString(progressRecord.Snapshot, "currentRoleName")
			item.ReportedAt = progressRecord.ReportedAt
		}

		if hintRecord, ok := s.store.LatestHint(studentID, assignmentRecord.ID); ok {
			item.HintText = hintRecord.HintText
			item.HintProvider = hintRecord.ProviderName
			item.HintCreatedAt = hintRecord.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		items = append(items, item)
	}

	return StudentHistoryResponse{
		StudentID:   studentRecord.ID,
		StudentName: studentRecord.DisplayName,
		Items:       items,
	}, nil
}

func snapshotString(snapshot map[string]any, key string) string {
	value, ok := snapshot[key].(string)
	if !ok {
		return ""
	}
	return value
}
