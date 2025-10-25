// Package openrouter provides OpenRouter API integration.
// OpenRouter is a unified interface to multiple LLM providers.
package openrouter

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
	defaultBaseURL = "https://openrouter.ai/api/v1"
)

// Provider implements the OpenRouter API provider.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	appName string // Optional app name for OpenRouter
	appURL  string // Optional app URL for OpenRouter
}

// New creates a new OpenRouter provider.
func New(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: 120 * time.Second, // Longer timeout for model routing
		},
		appName: "bplus",
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

// WithAppInfo sets the app name and URL for OpenRouter tracking.
func WithAppInfo(name, url string) Option {
	return func(p *Provider) {
		p.appName = name
		p.appURL = url
	}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "openrouter"
}

// ListModels returns popular models available via OpenRouter.
func (p *Provider) ListModels(ctx context.Context) ([]models.Model, error) {
	// Return a curated list of popular models
	// Full list can be fetched from /api/v1/models endpoint
	return []models.Model{
		{
			ID:            "anthropic/claude-3.5-sonnet",
			Name:          "Claude 3.5 Sonnet",
			Provider:      "openrouter",
			ContextWindow: 200000,
			MaxOutput:     8192,
			Pricing: models.Pricing{
				InputTokens:  3.00 / 1000000,
				OutputTokens: 15.00 / 1000000,
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "anthropic/claude-3-opus",
			Name:          "Claude 3 Opus",
			Provider:      "openrouter",
			ContextWindow: 200000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  15.00 / 1000000,
				OutputTokens: 75.00 / 1000000,
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "openai/gpt-4-turbo",
			Name:          "GPT-4 Turbo",
			Provider:      "openrouter",
			ContextWindow: 128000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  10.00 / 1000000,
				OutputTokens: 30.00 / 1000000,
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "openai/gpt-4o",
			Name:          "GPT-4o",
			Provider:      "openrouter",
			ContextWindow: 128000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  5.00 / 1000000,
				OutputTokens: 15.00 / 1000000,
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "google/gemini-pro-1.5",
			Name:          "Gemini Pro 1.5",
			Provider:      "openrouter",
			ContextWindow: 2097152,
			MaxOutput:     8192,
			Pricing: models.Pricing{
				InputTokens:  1.25 / 1000000,
				OutputTokens: 5.00 / 1000000,
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "meta-llama/llama-3.1-405b-instruct",
			Name:          "Llama 3.1 405B Instruct",
			Provider:      "openrouter",
			ContextWindow: 131072,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  3.00 / 1000000,
				OutputTokens: 3.00 / 1000000,
			},
			Capabilities: []string{"streaming", "tools"},
		},
		{
			ID:            "meta-llama/llama-3.1-70b-instruct",
			Name:          "Llama 3.1 70B Instruct",
			Provider:      "openrouter",
			ContextWindow: 131072,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  0.52 / 1000000,
				OutputTokens: 0.75 / 1000000,
			},
			Capabilities: []string{"streaming", "tools"},
		},
		{
			ID:            "mistralai/mistral-large",
			Name:          "Mistral Large",
			Provider:      "openrouter",
			ContextWindow: 128000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  2.00 / 1000000,
				OutputTokens: 6.00 / 1000000,
			},
			Capabilities: []string{"streaming", "tools"},
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
			Provider:  "openrouter",
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
			Provider:  "openrouter",
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

			// Track usage (OpenRouter provides this in streaming)
			if chunk.Usage != nil {
				totalUsage = &models.Usage{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
					TotalTokens:  chunk.Usage.TotalTokens,
					Cost:         calculateCostFromUsage(chunk.Usage),
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

	// Try to fetch models list as a connection test
	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed with status %d", resp.StatusCode)
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
				Description: "Model available via OpenRouter",
				Available:   true,
			}, nil
		}
	}

	// Model not in our curated list, but might still be available
	return &models.ModelInfo{
		Model: models.Model{
			ID:       modelID,
			Provider: "openrouter",
		},
		Description: "Model available via OpenRouter (pricing varies)",
		Available:   true,
	}, nil
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

	// OpenRouter-specific headers
	if p.appName != "" {
		req.Header.Set("X-Title", p.appName)
	}
	if p.appURL != "" {
		req.Header.Set("HTTP-Referer", p.appURL)
	}
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
			Cost:         calculateCostFromUsage(apiResp.Usage),
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

func calculateCostFromUsage(usage *usage) float64 {
	// OpenRouter provides generation_cost in the usage metadata
	// If not available, we'll calculate based on tokens
	if usage.GenerationCost > 0 {
		return usage.GenerationCost
	}

	// Fallback: estimate based on average pricing
	inputCost := 2.00 / 1000000  // Average input cost
	outputCost := 6.00 / 1000000 // Average output cost

	return (float64(usage.PromptTokens) * inputCost) + (float64(usage.CompletionTokens) * outputCost)
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
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	GenerationCost   float64 `json:"generation_cost,omitempty"` // OpenRouter-specific
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
