// Package models provides abstractions for LLM providers and model management.
package models

import (
	"context"
	"time"
)

// Provider represents an LLM provider (e.g., Anthropic, OpenAI, Ollama).
type Provider interface {
	// Name returns the provider's identifier (e.g., "anthropic", "openai", "ollama")
	Name() string

	// ListModels returns all available models for this provider
	ListModels(ctx context.Context) ([]Model, error)

	// CreateCompletion creates a completion with the given request
	CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// StreamCompletion creates a streaming completion
	StreamCompletion(ctx context.Context, req *CompletionRequest) (<-chan StreamToken, error)

	// TestConnection tests the provider's connectivity and authentication
	TestConnection(ctx context.Context) error

	// GetModelInfo returns detailed information about a specific model
	GetModelInfo(ctx context.Context, modelID string) (*ModelInfo, error)

	// SupportsStreaming returns true if the provider supports streaming
	SupportsStreaming() bool

	// SupportsTools returns true if the provider supports tool/function calling
	SupportsTools() bool
}

// Model represents an available LLM model.
type Model struct {
	ID            string    // Unique identifier (e.g., "claude-sonnet-4-5")
	Name          string    // Human-readable name
	Provider      string    // Provider name
	ContextWindow int       // Maximum context window in tokens
	MaxOutput     int       // Maximum output tokens
	Pricing       Pricing   // Cost information
	Capabilities  []string  // Supported features (e.g., "streaming", "tools", "vision")
	CreatedAt     time.Time // When the model was created/released
}

// ModelInfo provides detailed information about a model.
type ModelInfo struct {
	Model
	Description string            // Detailed description
	Available   bool              // Whether the model is currently available
	Metadata    map[string]string // Additional provider-specific metadata
}

// Pricing represents the cost structure for a model.
type Pricing struct {
	InputTokens  float64 // Cost per 1K input tokens (USD)
	OutputTokens float64 // Cost per 1K output tokens (USD)
	MinimumCost  float64 // Minimum charge per request (USD)
}

// CompletionRequest represents a request for text completion.
type CompletionRequest struct {
	// Model identifier
	Model string

	// Messages in the conversation
	Messages []Message

	// System prompt (optional)
	System string

	// Tool definitions for function calling (optional)
	Tools []Tool

	// Sampling parameters
	Temperature      *float64 // 0.0 to 1.0
	TopP             *float64 // 0.0 to 1.0
	TopK             *int     // Top-k sampling
	MaxTokens        int      // Maximum tokens to generate
	StopSequences    []string // Stop generation at these sequences
	FrequencyPenalty *float64 // -2.0 to 2.0
	PresencePenalty  *float64 // -2.0 to 2.0

	// Metadata
	Metadata map[string]string
}

// CompletionResponse represents a completion response.
type CompletionResponse struct {
	// Generated content
	Content string

	// Tool calls made by the model (if any)
	ToolCalls []ToolCall

	// Stop reason
	StopReason string // "end_turn", "max_tokens", "stop_sequence", "tool_use"

	// Usage information
	Usage Usage

	// Model used
	Model string

	// Response metadata
	Metadata map[string]string
}

// StreamToken represents a single token in a streaming response.
type StreamToken struct {
	Content  string    // Token content
	ToolCall *ToolCall // Tool call (if any)
	Done     bool      // True if this is the last token
	Usage    *Usage    // Usage info (sent with last token)
	Error    error     // Error (if any)
	Metadata map[string]string
}

// Message represents a conversation message.
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string // Message content
	Name    string // Optional name for multi-party conversations
}

// Tool represents a tool/function that the model can call.
type Tool struct {
	Name        string      // Tool name
	Description string      // Tool description
	Parameters  []Parameter // Tool parameters
	Required    []string    // Required parameter names
}

// Parameter represents a tool parameter.
type Parameter struct {
	Name        string   // Parameter name
	Type        string   // Parameter type (e.g., "string", "number", "boolean", "object")
	Description string   // Parameter description
	Required    bool     // Whether this parameter is required
	Enum        []string // Allowed values (optional)
}

// ToolCall represents a function/tool call made by the model.
type ToolCall struct {
	ID        string                 // Unique call ID
	Name      string                 // Tool name
	Arguments map[string]interface{} // Tool arguments
}

// Usage represents token usage and cost information.
type Usage struct {
	InputTokens  int     // Tokens in the prompt
	OutputTokens int     // Tokens in the completion
	TotalTokens  int     // Total tokens used
	Cost         float64 // Estimated cost in USD
}

// ProviderConfig represents configuration for a provider.
type ProviderConfig struct {
	Name    string            // Provider name
	APIKey  string            // API key (for cloud providers)
	BaseURL string            // Base URL (for custom endpoints)
	Timeout time.Duration     // Request timeout
	Retries int               // Number of retries
	Headers map[string]string // Custom HTTP headers
}

// Error types
type ProviderError struct {
	Provider  string
	Code      string
	Message   string
	Retryable bool
}

func (e *ProviderError) Error() string {
	return e.Provider + ": " + e.Message
}

// IsRetryable returns true if the error is retryable.
func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}
