package sb3

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/store/memory"
)

var ErrProjectJSONNotFound = errors.New("project.json not found in sb3")

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(rawSB3 []byte) (memory.AssignmentAnalysis, error) {
	reader, err := zip.NewReader(bytes.NewReader(rawSB3), int64(len(rawSB3)))
	if err != nil {
		return memory.AssignmentAnalysis{}, err
	}

	projectJSON, err := readProjectJSON(reader)
	if err != nil {
		return memory.AssignmentAnalysis{}, err
	}

	return buildAnalysis(projectJSON)
}

type projectJSONDocument struct {
	Targets    []targetDocument `json:"targets"`
	Extensions []string         `json:"extensions"`
}

type targetDocument struct {
	IsStage    bool                          `json:"isStage"`
	Name       string                        `json:"name"`
	Blocks     map[string]blockDocument      `json:"blocks"`
	Broadcasts map[string]string             `json:"broadcasts"`
	Variables  map[string]namedValueDocument `json:"variables"`
	Lists      map[string]namedValueDocument `json:"lists"`
}

type blockDocument struct {
	Opcode   string `json:"opcode"`
	Next     string `json:"next"`
	Parent   string `json:"parent"`
	TopLevel bool   `json:"topLevel"`
}

type namedValueDocument []any

func readProjectJSON(reader *zip.Reader) ([]byte, error) {
	for _, file := range reader.File {
		if file.Name != "project.json" {
			continue
		}

		handle, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer handle.Close()

		return io.ReadAll(handle)
	}

	return nil, ErrProjectJSONNotFound
}

func buildAnalysis(rawProjectJSON []byte) (memory.AssignmentAnalysis, error) {
	var document projectJSONDocument
	if err := json.Unmarshal(rawProjectJSON, &document); err != nil {
		return memory.AssignmentAnalysis{}, err
	}

	roleNames := make([]string, 0, len(document.Targets))
	scriptCounts := map[string]int{}
	blockCounts := map[string]int{}
	categoryCounts := map[string]int{}
	broadcastMessages := make([]string, 0)
	variableNames := make([]string, 0)
	listNames := make([]string, 0)
	extensions := make([]string, 0, len(document.Extensions))

	broadcastSeen := map[string]struct{}{}
	variableSeen := map[string]struct{}{}
	listSeen := map[string]struct{}{}
	extensionSeen := map[string]struct{}{}

	for _, extension := range document.Extensions {
		appendUniqueString(&extensions, extensionSeen, extension)
	}

	for _, target := range document.Targets {
		roleNames = append(roleNames, target.Name)
		scriptCounts[target.Name] = countScripts(target.Blocks)
		blockCounts[target.Name] = len(target.Blocks)

		for _, message := range target.Broadcasts {
			appendUniqueString(&broadcastMessages, broadcastSeen, message)
		}
		for _, variable := range target.Variables {
			appendUniqueString(&variableNames, variableSeen, variable.Name())
		}
		for _, list := range target.Lists {
			appendUniqueString(&listNames, listSeen, list.Name())
		}
		for _, block := range target.Blocks {
			category := blockCategory(block.Opcode)
			categoryCounts[category]++
		}
	}

	return memory.AssignmentAnalysis{
		RoleNames:         roleNames,
		ScriptCounts:      scriptCounts,
		BlockCounts:       blockCounts,
		CategoryCounts:    categoryCounts,
		BroadcastMessages: broadcastMessages,
		VariableNames:     variableNames,
		ListNames:         listNames,
		Extensions:        extensions,
		TeachingPoints:    teachingPoints(categoryCounts, broadcastMessages, variableNames, listNames, extensions),
	}, nil
}

func countScripts(blocks map[string]blockDocument) int {
	if len(blocks) == 0 {
		return 0
	}

	inboundNext := make(map[string]struct{}, len(blocks))
	for _, block := range blocks {
		nextID := strings.TrimSpace(block.Next)
		if nextID != "" {
			inboundNext[nextID] = struct{}{}
		}
	}

	count := 0
	for blockID, block := range blocks {
		if strings.TrimSpace(block.Opcode) == "" {
			continue
		}
		if block.TopLevel {
			count++
			continue
		}
		if strings.TrimSpace(block.Parent) != "" {
			continue
		}
		if _, referenced := inboundNext[strings.TrimSpace(blockID)]; referenced {
			continue
		}
		count++
	}

	return count
}

func blockCategory(opcode string) string {
	switch {
	case strings.HasPrefix(opcode, "event_"):
		return "event"
	case strings.HasPrefix(opcode, "motion_"):
		return "motion"
	case strings.HasPrefix(opcode, "control_"):
		return "control"
	case strings.HasPrefix(opcode, "looks_"):
		return "looks"
	case strings.HasPrefix(opcode, "sound_"):
		return "sound"
	case strings.HasPrefix(opcode, "sensing_"):
		return "sensing"
	case strings.HasPrefix(opcode, "operator_"):
		return "operator"
	case strings.HasPrefix(opcode, "data_"):
		return "data"
	case strings.HasPrefix(opcode, "pen_"):
		return "pen"
	default:
		return "other"
	}
}

func (d namedValueDocument) Name() string {
	if len(d) == 0 {
		return ""
	}

	name, _ := d[0].(string)
	return strings.TrimSpace(name)
}

func appendUniqueString(items *[]string, seen map[string]struct{}, raw string) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return
	}
	if _, exists := seen[value]; exists {
		return
	}

	seen[value] = struct{}{}
	*items = append(*items, value)
}

func teachingPoints(categoryCounts map[string]int, broadcastMessages []string, variableNames []string, listNames []string, extensions []string) []string {
	points := make([]string, 0, len(categoryCounts)+4)
	if categoryCounts["event"] > 0 {
		points = append(points, "作品依赖事件触发来启动流程")
	}
	if categoryCounts["motion"] > 0 {
		points = append(points, "作品包含角色移动相关积木")
	}
	if len(broadcastMessages) > 0 {
		points = append(points, "作品使用广播消息协调不同角色")
	}
	if len(variableNames) > 0 {
		points = append(points, "作品使用变量记录关键状态")
	}
	if len(listNames) > 0 {
		points = append(points, "作品使用列表组织一组数据")
	}
	if len(extensions) > 0 {
		points = append(points, "作品启用了扩展积木，需要关注扩展能力是否正确衔接")
	}
	if len(points) == 0 {
		points = append(points, "先按角色和积木组织作品主流程")
	}
	return points
}
