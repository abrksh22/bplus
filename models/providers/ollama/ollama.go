// Package ollama provides Ollama local model integration.
package ollama

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
	defaultBaseURL = "http://localhost:11434"
)

// Provider implements the Ollama provider for local models.
type Provider struct {
	baseURL string
	client  *http.Client
}

// New creates a new Ollama provider.
func New(opts ...Option) *Provider {
	p := &Provider{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: 5 * time.Minute, // Longer timeout for local model loading
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
	return "ollama"
}

// ListModels returns available models from Ollama.
func (p *Provider) ListModels(ctx context.Context) ([]models.Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &models.ProviderError{
			Provider:  "ollama",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: false,
		}
	}

	var tagsResp tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := make([]models.Model, len(tagsResp.Models))
	for i, model := range tagsResp.Models {
		result[i] = models.Model{
			ID:            model.Name,
			Name:          model.Name,
			Provider:      "ollama",
			ContextWindow: 4096, // Default, actual value depends on model
			MaxOutput:     2048,
			Pricing: models.Pricing{
				InputTokens:  0, // Free for local models
				OutputTokens: 0,
			},
			Capabilities: []string{"streaming"},
		}
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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
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
		return nil, &models.ProviderError{
			Provider:  "ollama",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: false,
		}
	}

	var apiResp chatResponse
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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &models.ProviderError{
			Provider:  "ollama",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: false,
		}
	}

	tokens := make(chan models.StreamToken, 10)

	go func() {
		defer close(tokens)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var totalTokens int

		for scanner.Scan() {
			var event chatResponse
			if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
				tokens <- models.StreamToken{Error: err}
				return
			}

			if event.Message.Content != "" {
				tokens <- models.StreamToken{
					Content: event.Message.Content,
				}
			}

			if event.Done {
				totalTokens = event.PromptEvalCount + event.EvalCount
				tokens <- models.StreamToken{
					Done: true,
					Usage: &models.Usage{
						InputTokens:  event.PromptEvalCount,
						OutputTokens: event.EvalCount,
						TotalTokens:  totalTokens,
						Cost:         0, // Free for local models
					},
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

// TestConnection tests the Ollama connection.
func (p *Provider) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/version", nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return nil
}

// GetModelInfo returns information about a specific model.
func (p *Provider) GetModelInfo(ctx context.Context, modelID string) (*models.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/show", bytes.NewBufferString(fmt.Sprintf(`{"name":"%s"}`, modelID)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model %s not found", modelID)
	}

	var showResp showResponse
	if err := json.NewDecoder(resp.Body).Decode(&showResp); err != nil {
		return nil, err
	}

	return &models.ModelInfo{
		Model: models.Model{
			ID:            modelID,
			Name:          modelID,
			Provider:      "ollama",
			ContextWindow: 4096,
			MaxOutput:     2048,
			Pricing: models.Pricing{
				InputTokens:  0,
				OutputTokens: 0,
			},
			Capabilities: []string{"streaming"},
		},
		Description: showResp.ModelInfo,
		Available:   true,
	}, nil
}

// SupportsStreaming returns true.
func (p *Provider) SupportsStreaming() bool {
	return true
}

// SupportsTools returns false (Ollama doesn't support tool calling natively).
func (p *Provider) SupportsTools() bool {
	return false
}

// Helper methods

func (p *Provider) convertRequest(req *models.CompletionRequest, stream bool) *chatRequest {
	apiReq := &chatRequest{
		Model:    req.Model,
		Messages: make([]message, len(req.Messages)),
		Stream:   stream,
		Options:  &options{},
	}

	// Add system message if present
	if req.System != "" {
		apiReq.Messages = append(apiReq.Messages, message{
			Role:    "system",
			Content: req.System,
		})
	}

	for i, msg := range req.Messages {
		apiReq.Messages[i] = message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	if req.Temperature != nil {
		apiReq.Options.Temperature = *req.Temperature
	}

	if req.TopP != nil {
		apiReq.Options.TopP = *req.TopP
	}

	if req.TopK != nil {
		apiReq.Options.TopK = *req.TopK
	}

	if len(req.StopSequences) > 0 {
		apiReq.Options.Stop = req.StopSequences
	}

	return apiReq
}

func (p *Provider) convertResponse(apiResp *chatResponse) *models.CompletionResponse {
	return &models.CompletionResponse{
		Content:    apiResp.Message.Content,
		Model:      apiResp.Model,
		StopReason: "end_turn",
		Usage: models.Usage{
			InputTokens:  apiResp.PromptEvalCount,
			OutputTokens: apiResp.EvalCount,
			TotalTokens:  apiResp.PromptEvalCount + apiResp.EvalCount,
			Cost:         0, // Free for local models
		},
	}
}

// API types

type tagsResponse struct {
	Models []modelInfo `json:"models"`
}

type modelInfo struct {
	Name string `json:"name"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  *options  `json:"options,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type options struct {
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	TopK        int      `json:"top_k,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

type chatResponse struct {
	Model           string  `json:"model"`
	Message         message `json:"message"`
	Done            bool    `json:"done"`
	PromptEvalCount int     `json:"prompt_eval_count"`
	EvalCount       int     `json:"eval_count"`
}

type showResponse struct {
	ModelInfo string `json:"modelinfo"`
}
