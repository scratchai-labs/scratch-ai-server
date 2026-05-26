package hint

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

var ErrEmptyHintResponse = errors.New("deepseek returned an empty hint")

type Provider interface {
	Generate(context.Context, GenerateInput) (GeneratedHint, error)
}

type GenerateInput struct {
	PromptInput map[string]any
}

type GeneratedHint struct {
	Text         string
	ProviderName string
}

type DeepSeekProvider struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewDeepSeekProvider(cfg config.DeepSeekConfig) Provider {
	if !cfg.Enabled() {
		return nil
	}

	return &DeepSeekProvider{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (p *DeepSeekProvider) Generate(ctx context.Context, input GenerateInput) (GeneratedHint, error) {
	systemPrompt, userPrompt, err := buildPromptMessages(input.PromptInput)
	if err != nil {
		return GeneratedHint{}, err
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model: p.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: 0.2,
		Stream:      false,
	})
	if err != nil {
		return GeneratedHint{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return GeneratedHint{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return GeneratedHint{}, err
	}
	defer res.Body.Close()

	rawResponse, err := io.ReadAll(res.Body)
	if err != nil {
		return GeneratedHint{}, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return GeneratedHint{}, fmt.Errorf("deepseek request failed with status %d: %s", res.StatusCode, strings.TrimSpace(string(rawResponse)))
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(rawResponse, &parsed); err != nil {
		return GeneratedHint{}, err
	}

	hintText := strings.TrimSpace(parsed.FirstMessageContent())
	if hintText == "" {
		return GeneratedHint{}, ErrEmptyHintResponse
	}

	return GeneratedHint{
		Text:         hintText,
		ProviderName: "deepseek",
	}, nil
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	Stream      bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []chatChoice `json:"choices"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

func (r chatCompletionResponse) FirstMessageContent() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

func buildPromptMessages(promptInput map[string]any) (string, string, error) {
	rawPrompt, err := json.Marshal(promptInput)
	if err != nil {
		return "", "", err
	}

	systemPrompt := "你是 Scratch 课堂助教。请只给学生下一步最小提示，保持简短、具体，不要直接替学生写完整作品。"
	userPrompt := "请基于下面的教学任务、参考作品分析和学生当前进度，给出下一步提示。\n" + string(rawPrompt)
	return systemPrompt, userPrompt, nil
}
