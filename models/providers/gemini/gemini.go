// Package gemini provides Google Gemini API integration.
package gemini

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
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// Provider implements the Google Gemini API provider.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new Gemini provider.
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
	return "gemini"
}

// ListModels returns available Gemini models.
func (p *Provider) ListModels(ctx context.Context) ([]models.Model, error) {
	return []models.Model{
		{
			ID:            "gemini-2.0-flash-exp",
			Name:          "Gemini 2.0 Flash (Experimental)",
			Provider:      "gemini",
			ContextWindow: 1048576, // 1M tokens
			MaxOutput:     8192,
			Pricing: models.Pricing{
				InputTokens:  0.00, // Free during preview
				OutputTokens: 0.00,
			},
			Capabilities: []string{"streaming", "tools", "vision", "audio"},
		},
		{
			ID:            "gemini-1.5-pro",
			Name:          "Gemini 1.5 Pro",
			Provider:      "gemini",
			ContextWindow: 2097152, // 2M tokens
			MaxOutput:     8192,
			Pricing: models.Pricing{
				InputTokens:  1.25 / 1000000, // $1.25 per million (≤128k tokens)
				OutputTokens: 5.00 / 1000000, // $5.00 per million
			},
			Capabilities: []string{"streaming", "tools", "vision", "audio"},
		},
		{
			ID:            "gemini-1.5-flash",
			Name:          "Gemini 1.5 Flash",
			Provider:      "gemini",
			ContextWindow: 1048576, // 1M tokens
			MaxOutput:     8192,
			Pricing: models.Pricing{
				InputTokens:  0.075 / 1000000, // $0.075 per million (≤128k tokens)
				OutputTokens: 0.30 / 1000000,  // $0.30 per million
			},
			Capabilities: []string{"streaming", "tools", "vision"},
		},
		{
			ID:            "gemini-1.5-flash-8b",
			Name:          "Gemini 1.5 Flash-8B",
			Provider:      "gemini",
			ContextWindow: 1048576, // 1M tokens
			MaxOutput:     8192,
			Pricing: models.Pricing{
				InputTokens:  0.0375 / 1000000, // $0.0375 per million (≤128k tokens)
				OutputTokens: 0.15 / 1000000,   // $0.15 per million
			},
			Capabilities: []string{"streaming", "tools", "vision"},
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

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, req.Model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
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
			Provider:  "gemini",
			Code:      fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Message:   string(body),
			Retryable: resp.StatusCode >= 500 || resp.StatusCode == 429,
		}
	}

	var apiResp generateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertResponse(&apiResp, req.Model), nil
}

// StreamCompletion creates a streaming completion.
func (p *Provider) StreamCompletion(ctx context.Context, req *models.CompletionRequest) (<-chan models.StreamToken, error) {
	apiReq := p.convertRequest(req, true)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", p.baseURL, req.Model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
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
			Provider:  "gemini",
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
			if data == "" {
				continue
			}

			var chunk generateContentResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				tokens <- models.StreamToken{Error: err}
				return
			}

			if len(chunk.Candidates) > 0 {
				candidate := chunk.Candidates[0]
				if candidate.Content.Parts != nil && len(candidate.Content.Parts) > 0 {
					for _, part := range candidate.Content.Parts {
						if part.Text != "" {
							tokens <- models.StreamToken{
								Content: part.Text,
							}
						}

						// Handle function calls
						if part.FunctionCall != nil {
							tokens <- models.StreamToken{
								ToolCall: &models.ToolCall{
									ID:        fmt.Sprintf("call_%d", time.Now().UnixNano()),
									Name:      part.FunctionCall.Name,
									Arguments: part.FunctionCall.Args,
								},
							}
						}
					}
				}

				// Check if done
				if candidate.FinishReason != "" && candidate.FinishReason != "FINISH_REASON_UNSPECIFIED" {
					// Calculate usage
					if chunk.UsageMetadata != nil {
						totalUsage = &models.Usage{
							InputTokens:  chunk.UsageMetadata.PromptTokenCount,
							OutputTokens: chunk.UsageMetadata.CandidatesTokenCount,
							TotalTokens:  chunk.UsageMetadata.TotalTokenCount,
							Cost:         calculateCost(req.Model, chunk.UsageMetadata.PromptTokenCount, chunk.UsageMetadata.CandidatesTokenCount),
						}
					}

					if totalUsage != nil {
						tokens <- models.StreamToken{
							Done:  true,
							Usage: totalUsage,
						}
					}
					return
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

	// Try to list models as a connection test
	url := fmt.Sprintf("%s/models?key=%s", p.baseURL, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

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
				Description: "Google Gemini model",
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
}

func (p *Provider) convertRequest(req *models.CompletionRequest, stream bool) *generateContentRequest {
	apiReq := &generateContentRequest{
		Contents: make([]content, 0),
	}

	// Add system instruction if present
	if req.System != "" {
		apiReq.SystemInstruction = &content{
			Parts: []part{{Text: req.System}},
		}
	}

	// Add conversation messages
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		apiReq.Contents = append(apiReq.Contents, content{
			Role:  role,
			Parts: []part{{Text: msg.Content}},
		})
	}

	// Generation config
	apiReq.GenerationConfig = &generationConfig{}

	if req.MaxTokens > 0 {
		apiReq.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}

	if req.Temperature != nil {
		apiReq.GenerationConfig.Temperature = req.Temperature
	}

	if req.TopP != nil {
		apiReq.GenerationConfig.TopP = req.TopP
	}

	if len(req.StopSequences) > 0 {
		apiReq.GenerationConfig.StopSequences = req.StopSequences
	}

	// Add tools if present
	if len(req.Tools) > 0 {
		apiReq.Tools = make([]tool, 1)
		functionDeclarations := make([]functionDeclaration, len(req.Tools))

		for i, t := range req.Tools {
			functionDeclarations[i] = functionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  convertToolParams(t.Parameters),
			}
		}

		apiReq.Tools[0] = tool{
			FunctionDeclarations: functionDeclarations,
		}
	}

	return apiReq
}

func (p *Provider) convertResponse(apiResp *generateContentResponse, model string) *models.CompletionResponse {
	resp := &models.CompletionResponse{
		Model: model,
	}

	if len(apiResp.Candidates) > 0 {
		candidate := apiResp.Candidates[0]

		// Extract text content
		var contentParts []string
		var toolCalls []models.ToolCall

		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				contentParts = append(contentParts, part.Text)
			}

			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, models.ToolCall{
					ID:        fmt.Sprintf("call_%d", time.Now().UnixNano()),
					Name:      part.FunctionCall.Name,
					Arguments: part.FunctionCall.Args,
				})
			}
		}

		resp.Content = strings.Join(contentParts, "")
		resp.ToolCalls = toolCalls

		// Map finish reason
		switch candidate.FinishReason {
		case "STOP":
			resp.StopReason = "end_turn"
		case "MAX_TOKENS":
			resp.StopReason = "max_tokens"
		default:
			if len(toolCalls) > 0 {
				resp.StopReason = "tool_use"
			} else {
				resp.StopReason = "end_turn"
			}
		}
	}

	// Usage
	if apiResp.UsageMetadata != nil {
		resp.Usage = models.Usage{
			InputTokens:  apiResp.UsageMetadata.PromptTokenCount,
			OutputTokens: apiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  apiResp.UsageMetadata.TotalTokenCount,
			Cost:         calculateCost(model, apiResp.UsageMetadata.PromptTokenCount, apiResp.UsageMetadata.CandidatesTokenCount),
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
	case strings.Contains(model, "gemini-2.0-flash-exp"):
		inputCost = 0.00 // Free during preview
		outputCost = 0.00
	case strings.Contains(model, "gemini-1.5-pro"):
		inputCost = 1.25 / 1000000
		outputCost = 5.00 / 1000000
	case strings.Contains(model, "gemini-1.5-flash-8b"):
		inputCost = 0.0375 / 1000000
		outputCost = 0.15 / 1000000
	case strings.Contains(model, "gemini-1.5-flash"):
		inputCost = 0.075 / 1000000
		outputCost = 0.30 / 1000000
	default:
		inputCost = 0.075 / 1000000
		outputCost = 0.30 / 1000000
	}

	return (float64(promptTokens) * inputCost) + (float64(completionTokens) * outputCost)
}

// API types

type generateContentRequest struct {
	Contents          []content         `json:"contents"`
	SystemInstruction *content          `json:"systemInstruction,omitempty"`
	Tools             []tool            `json:"tools,omitempty"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text         string        `json:"text,omitempty"`
	FunctionCall *functionCall `json:"functionCall,omitempty"`
}

type functionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type tool struct {
	FunctionDeclarations []functionDeclaration `json:"functionDeclarations"`
}

type functionDeclaration struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type generationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type generateContentResponse struct {
	Candidates    []candidate    `json:"candidates"`
	UsageMetadata *usageMetadata `json:"usageMetadata,omitempty"`
}

type candidate struct {
	Content      content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

type usageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}
