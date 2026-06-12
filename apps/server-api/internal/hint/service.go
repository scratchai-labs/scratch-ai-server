package hint

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var ErrAssignmentUnavailable = errors.New("assignment is not available to the student")
var ErrAssignmentNotReady = errors.New("assignment analysis is not ready")
var ErrProgressRequired = errors.New("student progress is required before requesting a hint")

type Service struct {
	store    *memory.Store
	provider Provider
}

func NewService(store *memory.Store, provider Provider) *Service {
	return &Service{
		store:    store,
		provider: provider,
	}
}

func (s *Service) Request(ctx context.Context, studentID int64, assignmentID int64) (memory.HintRecord, error) {
	assignmentRecord, ok := s.store.GetAssignmentForStudent(studentID, assignmentID)
	if !ok || assignmentRecord.Status != "published" {
		return memory.HintRecord{}, ErrAssignmentUnavailable
	}
	if assignmentRecord.AnalysisStatus != "ready" {
		return memory.HintRecord{}, ErrAssignmentNotReady
	}

	progressRecord, ok := s.store.LatestProgress(studentID, assignmentID)
	if !ok {
		return memory.HintRecord{}, ErrProgressRequired
	}

	latestHint, hasLatestHint := s.store.LatestHint(studentID, assignmentID)
	promptInput := buildPromptInput(assignmentRecord, progressRecord, latestHint, hasLatestHint)

	if s.provider != nil {
		generatedHint, err := s.provider.Generate(ctx, GenerateInput{PromptInput: promptInput})
		if err == nil && strings.TrimSpace(generatedHint.Text) != "" {
			return s.store.CreateHint(memory.CreateHintInput{
				AssignmentID:     assignmentID,
				StudentID:        studentID,
				ProgressReportID: progressRecord.ID,
				PromptInput:      promptInput,
				HintText:         generatedHint.Text,
				ProviderName:     generatedHint.ProviderName,
			})
		}
	}

	hintText := buildFallbackHint(assignmentRecord, progressRecord)
	return s.store.CreateHint(memory.CreateHintInput{
		AssignmentID:     assignmentID,
		StudentID:        studentID,
		ProgressReportID: progressRecord.ID,
		PromptInput:      promptInput,
		HintText:         hintText,
		ProviderName:     "fallback",
	})
}

func buildPromptInput(assignmentRecord memory.Assignment, progressRecord memory.ProgressReport, latestHint memory.HintRecord, hasLatestHint bool) map[string]any {
	promptInput := map[string]any{
		"assignmentTitle":       assignmentRecord.Title,
		"assignmentGoal":        assignmentRecord.Goal,
		"assignmentDescription": assignmentRecord.Description,
		"analysis": map[string]any{
			"roleNames":         assignmentRecord.Analysis.RoleNames,
			"scriptCounts":      assignmentRecord.Analysis.ScriptCounts,
			"blockCounts":       assignmentRecord.Analysis.BlockCounts,
			"categoryCounts":    assignmentRecord.Analysis.CategoryCounts,
			"broadcastMessages": assignmentRecord.Analysis.BroadcastMessages,
			"variableNames":     assignmentRecord.Analysis.VariableNames,
			"listNames":         assignmentRecord.Analysis.ListNames,
			"extensions":        assignmentRecord.Analysis.Extensions,
			"teachingPoints":    assignmentRecord.Analysis.TeachingPoints,
		},
		"studentProgress": map[string]any{
			"currentTarget":    progressRecord.CurrentTarget,
			"stepSummary":      progressRecord.StepSummary,
			"localProjectHash": progressRecord.LocalProjectHash,
			"reportedAt":       progressRecord.ReportedAt,
			"snapshot":         progressRecord.Snapshot,
		},
		"hintStyle": []string{
			"短",
			"具体",
			"只给下一步",
			"不要直接替学生写完整作品",
		},
	}

	if hasLatestHint {
		promptInput["recentHint"] = map[string]any{
			"hintText":     latestHint.HintText,
			"providerName": latestHint.ProviderName,
			"createdAt":    latestHint.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return promptInput
}

func buildFallbackHint(assignmentRecord memory.Assignment, progressRecord memory.ProgressReport) string {
	currentRoleName := strings.TrimSpace(snapshotString(progressRecord.Snapshot, "currentRoleName"))
	if currentRoleName == "" && len(assignmentRecord.Analysis.RoleNames) > 0 {
		currentRoleName = assignmentRecord.Analysis.RoleNames[0]
	}

	roleBlocks := blocksForRole(progressRecord.Snapshot, currentRoleName)
	nextTeachingPoint := assignmentRecord.Goal
	if len(assignmentRecord.Analysis.TeachingPoints) > 0 {
		nextTeachingPoint = assignmentRecord.Analysis.TeachingPoints[0]
	}

	if currentRoleName == "" {
		return fmt.Sprintf("先按当前任务目标继续推进：%s。", nextTeachingPoint)
	}

	if len(roleBlocks) == 0 {
		return fmt.Sprintf("先聚焦 %s 角色，把它需要的事件和动作积木补齐，再对照任务目标：%s。", currentRoleName, nextTeachingPoint)
	}

	return fmt.Sprintf("继续完善 %s 角色。你已经有 %s，下一步把这些积木串成完整流程，并对照任务目标：%s。", currentRoleName, strings.Join(roleBlocks, "、"), nextTeachingPoint)
}

func snapshotString(snapshot map[string]any, key string) string {
	value, ok := snapshot[key].(string)
	if !ok {
		return ""
	}
	return value
}

func blocksForRole(snapshot map[string]any, currentRoleName string) []string {
	roles, ok := snapshot["roles"].([]any)
	if !ok {
		return nil
	}

	for _, rawRole := range roles {
		roleRecord, ok := rawRole.(map[string]any)
		if !ok {
			continue
		}
		roleName, _ := roleRecord["roleName"].(string)
		if strings.TrimSpace(roleName) != currentRoleName {
			continue
		}

		rawBlocks, ok := roleRecord["blocks"].([]any)
		if !ok {
			return nil
		}

		blocks := make([]string, 0, len(rawBlocks))
		for _, rawBlock := range rawBlocks {
			blockName, ok := rawBlock.(string)
			if ok && strings.TrimSpace(blockName) != "" {
				blocks = append(blocks, blockName)
			}
		}
		return blocks
	}

	return nil
}
