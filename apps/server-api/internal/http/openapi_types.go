package http

import (
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/assignment"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/student"
)

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request body"`
}

type TeacherCredentialsRequestDoc struct {
	Username string `json:"username" binding:"required" example:"teacher01"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

type StudentLoginRequestDoc struct {
	Username   string `json:"username" binding:"required" example:"student01"`
	Password   string `json:"password" binding:"required" example:"abc12345"`
	ClientType string `json:"clientType" binding:"required" example:"desktop"`
}

type ProgressRoleSnapshotDoc struct {
	RoleName string   `json:"roleName" example:"Cat"`
	RoleType string   `json:"roleType" example:"sprite"`
	Blocks   []string `json:"blocks"`
}

type ProgressSnapshotDoc struct {
	CurrentRoleName string                    `json:"currentRoleName" example:"Cat"`
	Roles           []ProgressRoleSnapshotDoc `json:"roles"`
}

type ProgressReportRequestDoc struct {
	CurrentTarget    string              `json:"currentTarget" binding:"required" example:"让 Cat 角色移动起来"`
	StepSummary      string              `json:"stepSummary" binding:"required" example:"已经把事件积木接上了"`
	LocalProjectHash string              `json:"localProjectHash" example:"hash-1"`
	ReportedAt       string              `json:"reportedAt" example:"2026-05-26T12:00:00Z"`
	Snapshot         ProgressSnapshotDoc `json:"snapshot"`
}

type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

type TeacherMeResponse struct {
	TeacherID   int64  `json:"teacherId" example:"1"`
	TeacherName string `json:"teacherName" example:"teacher01"`
}

type StudentMeResponse struct {
	StudentID   int64  `json:"studentId" example:"1"`
	StudentName string `json:"studentName" example:"小明"`
	Username    string `json:"username" example:"student01"`
}

type TeacherStudentsResponse struct {
	Items []student.StudentItem `json:"items"`
}

type TeacherAssignmentsResponse struct {
	Items []assignment.TeacherAssignmentItem `json:"items"`
}

type StudentAssignmentsResponse struct {
	Items []assignment.StudentAssignmentItem `json:"items"`
}

type AssignmentUploadResponse struct {
	ID             int64  `json:"id" example:"1"`
	Title          string `json:"title" example:"Space Game"`
	Goal           string `json:"goal" example:"让角色移动起来"`
	Description    string `json:"description" example:"第一阶段任务"`
	Status         string `json:"status" example:"draft"`
	AnalysisStatus string `json:"analysisStatus" example:"pending"`
}

type AssignmentAnalysisResponse struct {
	AssignmentID         int64          `json:"assignmentId" example:"1"`
	AnalysisStatus       string         `json:"analysisStatus" example:"ready"`
	AnalysisErrorMessage string         `json:"analysisErrorMessage" example:""`
	RoleNames            []string       `json:"roleNames"`
	ScriptCounts         map[string]int `json:"scriptCounts"`
	BlockCounts          map[string]int `json:"blockCounts"`
	CategoryCounts       map[string]int `json:"categoryCounts"`
	BroadcastMessages    []string       `json:"broadcastMessages"`
	VariableNames        []string       `json:"variableNames"`
	ListNames            []string       `json:"listNames"`
	Extensions           []string       `json:"extensions"`
	TeachingPoints       []string       `json:"teachingPoints"`
}

type AssignmentStudentsResponse struct {
	AssignmentID  int64   `json:"assignmentId" example:"1"`
	StudentIDs    []int64 `json:"studentIds"`
	AssignedCount int     `json:"assignedCount" example:"2"`
}

type AssignmentStatusResponse struct {
	ID             int64  `json:"id" example:"1"`
	Title          string `json:"title" example:"Space Game"`
	Status         string `json:"status" example:"published"`
	AnalysisStatus string `json:"analysisStatus,omitempty" example:"ready"`
}

type ProgressCreatedResponse struct {
	ID            int64  `json:"id" example:"1"`
	AssignmentID  int64  `json:"assignmentId" example:"1"`
	CurrentTarget string `json:"currentTarget" example:"让角色左右移动"`
	StepSummary   string `json:"stepSummary" example:"已经接上方向键事件"`
	ReportedAt    string `json:"reportedAt" example:"2026-05-26T12:00:00Z"`
}

type HintCreatedResponse struct {
	ID           int64  `json:"id" example:"1"`
	AssignmentID int64  `json:"assignmentId" example:"1"`
	HintText     string `json:"hintText" example:"先把 Cat 的事件和移动积木串起来。"`
	ProviderName string `json:"providerName" example:"deepseek"`
}
