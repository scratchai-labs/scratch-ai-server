package http

import (
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/assignment"
	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/classroom"
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
	Role        string `json:"role" example:"teacher"`
}

type StudentMeResponse struct {
	StudentID   int64  `json:"studentId" example:"1"`
	StudentName string `json:"studentName" example:"小明"`
	Username    string `json:"username" example:"student01"`
}

type TeacherStudentsResponse struct {
	Items []student.StudentItem `json:"items"`
}

type TeacherClassroomsResponse struct {
	Items []classroom.Item `json:"items"`
}

type AdminTeacherCreateRequestDoc struct {
	Username        string `json:"username" binding:"required" example:"teacher01"`
	InitialPassword string `json:"initialPassword" binding:"required" example:"secret123"`
}

type AdminStudentCreateRequestDoc struct {
	TeacherID       int64  `json:"teacherId" binding:"required" example:"2"`
	Username        string `json:"username" binding:"required" example:"student01"`
	DisplayName     string `json:"displayName" binding:"required" example:"小明"`
	InitialPassword string `json:"initialPassword" binding:"required" example:"abc12345"`
}

type AdminTeacherPasswordResetRequestDoc struct {
	NewPassword string `json:"newPassword" binding:"required" example:"updated123"`
}

type AdminTeacherRoleUpdateRequestDoc struct {
	Role string `json:"role" binding:"required" example:"admin"`
}

type AdminStudentPasswordResetRequestDoc struct {
	NewPassword string `json:"newPassword" binding:"required" example:"updated123"`
}

type AdminOverviewResponse struct {
	AdminCount           int `json:"adminCount" example:"1"`
	TeacherCount         int `json:"teacherCount" example:"8"`
	ActiveTeacherCount   int `json:"activeTeacherCount" example:"7"`
	DisabledTeacherCount int `json:"disabledTeacherCount" example:"1"`
	StudentCount         int `json:"studentCount" example:"180"`
	ActiveStudentCount   int `json:"activeStudentCount" example:"172"`
	DisabledStudentCount int `json:"disabledStudentCount" example:"8"`
}

type AdminTeacherItemResponse struct {
	ID        int64  `json:"id" example:"1"`
	Username  string `json:"username" example:"teacher01"`
	Role      string `json:"role" example:"teacher"`
	Status    string `json:"status" example:"active"`
	CreatedAt string `json:"createdAt" example:"2026-06-13T12:00:00Z"`
}

type AdminStudentItemResponse struct {
	ID              int64  `json:"id" example:"1"`
	TeacherID       int64  `json:"teacherId" example:"2"`
	TeacherUsername string `json:"teacherUsername" example:"teacher01"`
	Username        string `json:"username" example:"student01"`
	DisplayName     string `json:"displayName" example:"小明"`
	Status          string `json:"status" example:"active"`
	CreatedAt       string `json:"createdAt" example:"2026-06-13T12:00:00Z"`
}

type AdminTeachersResponse struct {
	Items []AdminTeacherItemResponse `json:"items"`
}

type AdminStudentsResponse struct {
	Items []AdminStudentItemResponse `json:"items"`
}

type AdminAuditLogItemResponse struct {
	ID             int64             `json:"id" example:"4"`
	ActorUsername  string            `json:"actorUsername" example:"admin"`
	Action         string            `json:"action" example:"teacher.role_change"`
	TargetType     string            `json:"targetType" example:"teacher"`
	TargetID       int64             `json:"targetId" example:"2"`
	TargetUsername string            `json:"targetUsername" example:"teacher01"`
	Before         map[string]string `json:"before"`
	After          map[string]string `json:"after"`
	CreatedAt      string            `json:"createdAt" example:"2026-06-14T12:00:00Z"`
}

type AdminAuditLogsResponse struct {
	Items []AdminAuditLogItemResponse `json:"items"`
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
