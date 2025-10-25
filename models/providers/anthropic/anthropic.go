// Package anthropic provides Anthropic Claude API integration.
package anthropic

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
	defaultBaseURL = "https://api.anthropic.com/v1"
	apiVersion     = "2023-06-01"
)

// Provider implements the Anthropic Claude API provider.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new Anthropic provider.
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
	return "anthropic"
}

// ListModels returns available Claude models.
func (p *Provider) ListModels(ctx context.Context) ([]models.Model, error) {
	// Anthropic doesn't have a models endpoint, so we return known models
	return []models.Model{
		{
			ID:            "claude-opus-4-1",
			Name:          "Claude Opus 4.1",
			Provider:      "anthropic",
			ContextWindow: 200000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  15.00 / 1000000, // $15 per million tokens
				OutputTokens: 75.00 / 1000000, // $75 per million tokens
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "claude-sonnet-4-5",
			Name:          "Claude Sonnet 4.5",
			Provider:      "anthropic",
			ContextWindow: 200000,
			MaxOutput:     8192,
			Pricing: models.Pricing{
				InputTokens:  3.00 / 1000000,  // $3 per million tokens
				OutputTokens: 15.00 / 1000000, // $15 per million tokens
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "claude-haiku-4-0",
			Name:          "Claude Haiku 4.0",
			Provider:      "anthropic",
			ContextWindow: 200000,
			MaxOutput:     4096,
			Pricing: models.Pricing{
				InputTokens:  0.80 / 1000000, // $0.80 per million tokens
				OutputTokens: 4.00 / 1000000, // $4 per million tokens
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
	}, nil
}

// CreateCompletion creates a non-streaming completion.
func (p *Provider) CreateCompletion(ctx context.Context, req *models.CompletionRequest) (*models.CompletionResponse, error) {
	// Convert to Anthropic API format
	apiReq := p.convertRequest(req, false)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
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
			Provider:  "anthropic",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: resp.StatusCode >= 500 || resp.StatusCode == 429,
		}
	}

	var apiResp messageResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertResponse(&apiResp), nil
}

// StreamCompletion creates a streaming completion.
func (p *Provider) StreamCompletion(ctx context.Context, req *models.CompletionRequest) (<-chan models.StreamToken, error) {
	// Convert to Anthropic API format
	apiReq := p.convertRequest(req, true)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &models.ProviderError{
			Provider:  "anthropic",
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
		var usage *models.Usage

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var event streamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				tokens <- models.StreamToken{Error: err}
				return
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta != nil && event.Delta.Text != "" {
					tokens <- models.StreamToken{
						Content: event.Delta.Text,
					}
				}

			case "message_delta":
				if event.Usage != nil {
					usage = &models.Usage{
						InputTokens:  0, // Will be set in message_start
						OutputTokens: event.Usage.OutputTokens,
					}
				}

			case "message_start":
				if event.Message != nil && event.Message.Usage != nil {
					usage = &models.Usage{
						InputTokens: event.Message.Usage.InputTokens,
					}
				}

			case "message_stop":
				// Send final token with usage
				if usage != nil {
					usage.TotalTokens = usage.InputTokens + usage.OutputTokens
					// Calculate cost based on model pricing
					// TODO: Get actual model pricing
					usage.Cost = calculateCost(req.Model, usage.InputTokens, usage.OutputTokens)
				}
				tokens <- models.StreamToken{
					Done:  true,
					Usage: usage,
				}
				return
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
	// Simple test: try to list models (which is local, so just check API key)
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
				Description: "Claude model by Anthropic",
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
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
}

func (p *Provider) convertRequest(req *models.CompletionRequest, stream bool) *messageRequest {
	apiReq := &messageRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Stream:    stream,
		Messages:  make([]message, len(req.Messages)),
	}

	if req.System != "" {
		apiReq.System = req.System
	}

	for i, msg := range req.Messages {
		apiReq.Messages[i] = message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	if req.Temperature != nil {
		apiReq.Temperature = *req.Temperature
	}

	if req.TopP != nil {
		apiReq.TopP = *req.TopP
	}

	if req.TopK != nil {
		apiReq.TopK = *req.TopK
	}

	if len(req.StopSequences) > 0 {
		apiReq.StopSequences = req.StopSequences
	}

	return apiReq
}

func (p *Provider) convertResponse(apiResp *messageResponse) *models.CompletionResponse {
	resp := &models.CompletionResponse{
		Model:      apiResp.Model,
		StopReason: apiResp.StopReason,
	}

	// Extract text content
	if len(apiResp.Content) > 0 {
		var parts []string
		for _, content := range apiResp.Content {
			if content.Type == "text" {
				parts = append(parts, content.Text)
			}
		}
		resp.Content = strings.Join(parts, "")
	}

	// Usage
	if apiResp.Usage != nil {
		resp.Usage = models.Usage{
			InputTokens:  apiResp.Usage.InputTokens,
			OutputTokens: apiResp.Usage.OutputTokens,
			TotalTokens:  apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
			Cost:         calculateCost(apiResp.Model, apiResp.Usage.InputTokens, apiResp.Usage.OutputTokens),
		}
	}

	return resp
}

func calculateCost(model string, inputTokens, outputTokens int) float64 {
	// Simplified cost calculation - in production, this should use actual pricing
	var inputCost, outputCost float64

	switch {
	case strings.Contains(model, "opus"):
		inputCost = 15.00 / 1000000
		outputCost = 75.00 / 1000000
	case strings.Contains(model, "sonnet"):
		inputCost = 3.00 / 1000000
		outputCost = 15.00 / 1000000
	case strings.Contains(model, "haiku"):
		inputCost = 0.80 / 1000000
		outputCost = 4.00 / 1000000
	default:
		inputCost = 3.00 / 1000000
		outputCost = 15.00 / 1000000
	}

	return (float64(inputTokens) * inputCost) + (float64(outputTokens) * outputCost)
}

// API types

type messageRequest struct {
	Model         string    `json:"model"`
	Messages      []message `json:"messages"`
	System        string    `json:"system,omitempty"`
	MaxTokens     int       `json:"max_tokens"`
	Temperature   float64   `json:"temperature,omitempty"`
	TopP          float64   `json:"top_p,omitempty"`
	TopK          int       `json:"top_k,omitempty"`
	StopSequences []string  `json:"stop_sequences,omitempty"`
	Stream        bool      `json:"stream,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messageResponse struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []contentBlock `json:"content"`
	Model      string         `json:"model"`
	StopReason string         `json:"stop_reason"`
	Usage      *usageInfo     `json:"usage"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type usageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type streamEvent struct {
	Type    string           `json:"type"`
	Delta   *contentDelta    `json:"delta,omitempty"`
	Usage   *usageInfo       `json:"usage,omitempty"`
	Message *messageResponse `json:"message,omitempty"`
}

type contentDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
