// Package lmstudio provides LM Studio API integration.
// LM Studio is a local model hosting application with OpenAI-compatible API.
package lmstudio

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
	defaultBaseURL = "http://localhost:1234/v1"
)

// Provider implements the LM Studio API provider.
type Provider struct {
	baseURL string
	client  *http.Client
}

// New creates a new LM Studio provider.
func New(opts ...Option) *Provider {
	p := &Provider{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: 300 * time.Second, // Long timeout for local models
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
	return "lmstudio"
}

// ListModels returns available models from LM Studio.
func (p *Provider) ListModels(ctx context.Context) ([]models.Model, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list models: %s", string(body))
	}

	var apiResp listModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := make([]models.Model, 0, len(apiResp.Data))
	for _, m := range apiResp.Data {
		result = append(result, models.Model{
			ID:            m.ID,
			Name:          m.ID, // LM Studio uses ID as name
			Provider:      "lmstudio",
			ContextWindow: determineContextWindow(m.ID),
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  0, // Local models are free
				OutputTokens: 0,
			},
			Capabilities: []string{"streaming"},
		})
	}

	return result, nil
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
			Provider:  "lmstudio",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: resp.StatusCode >= 500,
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
			Provider:  "lmstudio",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: resp.StatusCode >= 500,
		}
	}

	tokens := make(chan models.StreamToken, 10)

	go func() {
		defer close(tokens)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var totalUsage *models.Usage
		var totalTokens int

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				if totalUsage == nil {
					// Create usage if not provided
					totalUsage = &models.Usage{
						OutputTokens: totalTokens,
						TotalTokens:  totalTokens,
						Cost:         0, // Local models are free
					}
				}
				tokens <- models.StreamToken{
					Done:  true,
					Usage: totalUsage,
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
					// Rough token estimation (1 token ≈ 4 characters)
					totalTokens += len(delta.Content) / 4
				}

				// Handle tool calls if present (limited support)
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

			// Track usage if provided
			if chunk.Usage != nil {
				totalUsage = &models.Usage{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
					TotalTokens:  chunk.Usage.TotalTokens,
					Cost:         0, // Local models are free
				}
			}
		}

		if err := scanner.Err(); err != nil {
			tokens <- models.StreamToken{Error: err}
		}
	}()

	return tokens, nil
}

// TestConnection tests the connection to LM Studio.
func (p *Provider) TestConnection(ctx context.Context) error {
	// Try to list models as a connection test
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("LM Studio not running or not accessible at %s: %w", p.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LM Studio returned status %d", resp.StatusCode)
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
				Description: "Local model running in LM Studio",
				Available:   true,
			}, nil
		}
	}

	return nil, fmt.Errorf("model %s not found in LM Studio", modelID)
}

// SupportsStreaming returns true.
func (p *Provider) SupportsStreaming() bool {
	return true
}

// SupportsTools returns false (limited tool support, model-dependent).
func (p *Provider) SupportsTools() bool {
	return false // Most local models don't support tool calling reliably
}

// Helper methods

func (p *Provider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
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

	// Note: Tool support is limited and model-dependent in LM Studio
	// We skip tool definitions for now

	return apiReq
}

func (p *Provider) convertResponse(apiResp *chatCompletionResponse) *models.CompletionResponse {
	resp := &models.CompletionResponse{
		Model: apiResp.Model,
	}

	if len(apiResp.Choices) > 0 {
		choice := apiResp.Choices[0]
		resp.Content = choice.Message.Content

		// Map finish reason
		switch choice.FinishReason {
		case "stop":
			resp.StopReason = "end_turn"
		case "length":
			resp.StopReason = "max_tokens"
		default:
			resp.StopReason = "end_turn"
		}

		// Handle tool calls if present (rare)
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
			Cost:         0, // Local models are free
		}
	}

	return resp
}

// determineContextWindow estimates context window based on model name.
func determineContextWindow(modelID string) int {
	modelLower := strings.ToLower(modelID)

	// Common context window patterns
	switch {
	case strings.Contains(modelLower, "32k"):
		return 32768
	case strings.Contains(modelLower, "16k"):
		return 16384
	case strings.Contains(modelLower, "8k"):
		return 8192
	case strings.Contains(modelLower, "llama-3"):
		return 8192
	case strings.Contains(modelLower, "mistral"):
		return 8192
	case strings.Contains(modelLower, "gemma"):
		return 8192
	default:
		return 4096 // Conservative default
	}
}

// API types

type listModelsResponse struct {
	Data []modelInfo `json:"data"`
}

type modelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type chatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
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
