// Package openai provides OpenAI API integration.
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/abrksh22/bplus/models"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
)

// Provider implements the OpenAI API provider.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new OpenAI provider.
func New(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Option is a functional option for configuring the provider.
type Option func(*Provider)

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) Option {
	return func(p *Provider) {
		p.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "openai"
}

// ListModels returns available OpenAI models.
func (p *Provider) ListModels(ctx context.Context) ([]models.Model, error) {
	// Return commonly used models
	return []models.Model{
		{
			ID:            "gpt-4-turbo",
			Name:          "GPT-4 Turbo",
			Provider:      "openai",
			ContextWindow: 128000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  10.00 / 1000000, // $10 per million
				OutputTokens: 30.00 / 1000000, // $30 per million
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "gpt-4o",
			Name:          "GPT-4o",
			Provider:      "openai",
			ContextWindow: 128000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  5.00 / 1000000,  // $5 per million
				OutputTokens: 15.00 / 1000000, // $15 per million
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "gpt-4o-mini",
			Name:          "GPT-4o Mini",
			Provider:      "openai",
			ContextWindow: 128000,
			MaxOutput:     16384,
			Pricing: models.Pricing{
				InputTokens:  0.15 / 1000000, // $0.15 per million
				OutputTokens: 0.60 / 1000000, // $0.60 per million
			},
			Capabilities: []string{"streaming", "tools"},
		},
		{
			ID:            "o1",
			Name:          "O1",
			Provider:      "openai",
			ContextWindow: 200000,
			MaxOutput:     100000,
			Pricing: models.Pricing{
				InputTokens:  15.00 / 1000000, // $15 per million
				OutputTokens: 60.00 / 1000000, // $60 per million
			},
			Capabilities: []string{"streaming"},
		},
		{
			ID:            "o1-mini",
			Name:          "O1 Mini",
			Provider:      "openai",
			ContextWindow: 128000,
			MaxOutput:     65536,
			Pricing: models.Pricing{
				InputTokens:  3.00 / 1000000,  // $3 per million
				OutputTokens: 12.00 / 1000000, // $12 per million
			},
			Capabilities: []string{"streaming"},
		},
	}, nil
}

// CreateCompletion creates a non-streaming completion.
func (p *Provider) CreateCompletion(ctx context.Context, req *models.CompletionRequest) (*models.CompletionResponse, error) {
	apiReq := p.convertRequest(req, false)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &models.ProviderError{
			Provider:  "openai",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: resp.StatusCode >= 500 || resp.StatusCode == 429,
		}
	}

	var apiResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertResponse(&apiResp), nil
}

// StreamCompletion creates a streaming completion.
func (p *Provider) StreamCompletion(ctx context.Context, req *models.CompletionRequest) (<-chan models.StreamToken, error) {
	apiReq := p.convertRequest(req, true)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(httpReq)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &models.ProviderError{
			Provider:  "openai",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: resp.StatusCode >= 500 || resp.StatusCode == 429,
		}
	}

	tokens := make(chan models.StreamToken, 10)

	go func() {
		defer close(tokens)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var totalUsage *models.Usage

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				if totalUsage != nil {
					tokens <- models.StreamToken{
						Done:  true,
						Usage: totalUsage,
					}
				}
				return
			}

			var chunk chatCompletionChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				tokens <- models.StreamToken{Error: err}
				return
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					tokens <- models.StreamToken{
						Content: delta.Content,
					}
				}

				// Handle tool calls if present
				if len(delta.ToolCalls) > 0 {
					for _, tc := range delta.ToolCalls {
						if tc.Function.Name != "" {
							var args map[string]interface{}
							json.Unmarshal([]byte(tc.Function.Arguments), &args)
							tokens <- models.StreamToken{
								ToolCall: &models.ToolCall{
									ID:        tc.ID,
									Name:      tc.Function.Name,
									Arguments: args,
								},
							}
						}
					}
				}
			}

			// Track usage
			if chunk.Usage != nil {
				totalUsage = &models.Usage{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
					TotalTokens:  chunk.Usage.TotalTokens,
					Cost:         calculateCost(req.Model, chunk.Usage.PromptTokens, chunk.Usage.CompletionTokens),
				}
			}
		}

		if err := scanner.Err(); err != nil {
			tokens <- models.StreamToken{Error: err}
		}
	}()

	return tokens, nil
}

// TestConnection tests the API connection.
func (p *Provider) TestConnection(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("API key not set")
	}
	return nil
}

// GetModelInfo returns information about a specific model.
func (p *Provider) GetModelInfo(ctx context.Context, modelID string) (*models.ModelInfo, error) {
	allModels, err := p.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	for _, model := range allModels {
		if model.ID == modelID {
			return &models.ModelInfo{
				Model:       model,
				Description: "OpenAI model",
				Available:   true,
			}, nil
		}
	}

	return nil, fmt.Errorf("model %s not found", modelID)
}

// SupportsStreaming returns true.
func (p *Provider) SupportsStreaming() bool {
	return true
}

// SupportsTools returns true.
func (p *Provider) SupportsTools() bool {
	return true
}

// Helper methods

func (p *Provider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
}

func (p *Provider) convertRequest(req *models.CompletionRequest, stream bool) *chatCompletionRequest {
	apiReq := &chatCompletionRequest{
		Model:    req.Model,
		Stream:   stream,
		Messages: make([]chatMessage, 0, len(req.Messages)+1),
	}

	// Add system message if present
	if req.System != "" {
		apiReq.Messages = append(apiReq.Messages, chatMessage{
			Role:    "system",
			Content: req.System,
		})
	}

	// Add conversation messages
	for _, msg := range req.Messages {
		apiReq.Messages = append(apiReq.Messages, chatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if req.MaxTokens > 0 {
		apiReq.MaxTokens = req.MaxTokens
	}

	if req.Temperature != nil {
		apiReq.Temperature = req.Temperature
	}

	if req.TopP != nil {
		apiReq.TopP = req.TopP
	}

	if len(req.StopSequences) > 0 {
		apiReq.Stop = req.StopSequences
	}

	// Add tools if present
	if len(req.Tools) > 0 {
		apiReq.Tools = make([]tool, len(req.Tools))
		for i, t := range req.Tools {
			apiReq.Tools[i] = tool{
				Type: "function",
				Function: functionDef{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  convertToolParams(t.Parameters),
				},
			}
		}
	}

	return apiReq
}

func (p *Provider) convertResponse(apiResp *chatCompletionResponse) *models.CompletionResponse {
	resp := &models.CompletionResponse{
		Model: apiResp.Model,
	}

	if len(apiResp.Choices) > 0 {
		choice := apiResp.Choices[0]
		resp.Content = choice.Message.Content
		resp.StopReason = choice.FinishReason

		// Handle tool calls
		if len(choice.Message.ToolCalls) > 0 {
			resp.ToolCalls = make([]models.ToolCall, len(choice.Message.ToolCalls))
			for i, tc := range choice.Message.ToolCalls {
				var args map[string]interface{}
				json.Unmarshal([]byte(tc.Function.Arguments), &args)
				resp.ToolCalls[i] = models.ToolCall{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: args,
				}
			}
			resp.StopReason = "tool_use"
		}
	}

	// Usage
	if apiResp.Usage != nil {
		resp.Usage = models.Usage{
			InputTokens:  apiResp.Usage.PromptTokens,
			OutputTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:  apiResp.Usage.TotalTokens,
			Cost:         calculateCost(apiResp.Model, apiResp.Usage.PromptTokens, apiResp.Usage.CompletionTokens),
		}
	}

	return resp
}

func convertToolParams(params []models.Parameter) map[string]interface{} {
	properties := make(map[string]interface{})
	required := []string{}

	for _, p := range params {
		properties[p.Name] = map[string]interface{}{
			"type":        p.Type,
			"description": p.Description,
		}
		if p.Required {
			required = append(required, p.Name)
		}
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func calculateCost(model string, promptTokens, completionTokens int) float64 {
	var inputCost, outputCost float64

	switch {
	case strings.Contains(model, "gpt-4-turbo"):
		inputCost = 10.00 / 1000000
		outputCost = 30.00 / 1000000
	case strings.Contains(model, "gpt-4o-mini"):
		inputCost = 0.15 / 1000000
		outputCost = 0.60 / 1000000
	case strings.Contains(model, "gpt-4o"):
		inputCost = 5.00 / 1000000
		outputCost = 15.00 / 1000000
	case strings.Contains(model, "o1-mini"):
		inputCost = 3.00 / 1000000
		outputCost = 12.00 / 1000000
	case strings.Contains(model, "o1"):
		inputCost = 15.00 / 1000000
		outputCost = 60.00 / 1000000
	default:
		inputCost = 5.00 / 1000000
		outputCost = 15.00 / 1000000
	}

	return (float64(promptTokens) * inputCost) + (float64(completionTokens) * outputCost)
}

// API types

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Tools       []tool        `json:"tools,omitempty"`
}

type chatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}

type tool struct {
	Type     string      `json:"type"`
	Function functionDef `json:"function"`
}

type functionDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolCallFunc `json:"function"`
}

type toolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatCompletionResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   *usage   `json:"usage"`
}

type choice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatCompletionChunk struct {
	ID      string        `json:"id"`
	Model   string        `json:"model"`
	Choices []choiceDelta `json:"choices"`
	Usage   *usage        `json:"usage"`
}

type choiceDelta struct {
	Index        int          `json:"index"`
	Delta        messageDelta `json:"delta"`
	FinishReason string       `json:"finish_reason"`
}

type messageDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}
