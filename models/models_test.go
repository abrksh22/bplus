package models

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseModelName tests model name parsing.
func TestParseModelName(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantProvider  string
		wantModelID   string
		wantErr       bool
	}{
		{
			name:         "Valid format",
			input:        "anthropic/claude-sonnet-4-5",
			wantProvider: "anthropic",
			wantModelID:  "claude-sonnet-4-5",
			wantErr:      false,
		},
		{
			name:         "Ollama model",
			input:        "ollama/llama2:13b",
			wantProvider: "ollama",
			wantModelID:  "llama2:13b",
			wantErr:      false,
		},
		{
			name:         "OpenAI model",
			input:        "openai/gpt-4-turbo",
			wantProvider: "openai",
			wantModelID:  "gpt-4-turbo",
			wantErr:      false,
		},
		{
			name:    "Missing slash",
			input:   "anthropic-claude",
			wantErr: true,
		},
		{
			name:    "Empty provider",
			input:   "/claude-sonnet",
			wantErr: true,
		},
		{
			name:    "Empty model ID",
			input:   "anthropic/",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, modelID, err := ParseModelName(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantProvider, provider)
			assert.Equal(t, tt.wantModelID, modelID)
		})
	}
}

// TestFormatModelName tests model name formatting.
func TestFormatModelName(t *testing.T) {
	tests := []struct {
		provider string
		modelID  string
		want     string
	}{
		{"anthropic", "claude-sonnet-4-5", "anthropic/claude-sonnet-4-5"},
		{"ollama", "llama2", "ollama/llama2"},
		{"openai", "gpt-4", "openai/gpt-4"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatModelName(tt.provider, tt.modelID)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestValidateModelName tests model name validation.
func TestValidateModelName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid", "anthropic/claude-sonnet", false},
		{"Invalid - no slash", "anthropic-claude", true},
		{"Invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModelName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetProviderFromModel tests provider extraction.
func TestGetProviderFromModel(t *testing.T) {
	provider, err := GetProviderFromModel("anthropic/claude-sonnet-4-5")
	require.NoError(t, err)
	assert.Equal(t, "anthropic", provider)

	_, err = GetProviderFromModel("invalid")
	assert.Error(t, err)
}

// TestGetModelIDFromModel tests model ID extraction.
func TestGetModelIDFromModel(t *testing.T) {
	modelID, err := GetModelIDFromModel("anthropic/claude-sonnet-4-5")
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet-4-5", modelID)

	_, err = GetModelIDFromModel("invalid")
	assert.Error(t, err)
}

// Mock provider for testing
type mockProvider struct {
	name             string
	models           []Model
	supportsStream   bool
	supportsTools    bool
	testConnectError error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) ListModels(ctx context.Context) ([]Model, error) {
	return m.models, nil
}

func (m *mockProvider) CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{
		Content:    "Test response",
		Model:      req.Model,
		StopReason: "end_turn",
		Usage: Usage{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
			Cost:         0.001,
		},
	}, nil
}

func (m *mockProvider) StreamCompletion(ctx context.Context, req *CompletionRequest) (<-chan StreamToken, error) {
	tokens := make(chan StreamToken, 2)
	go func() {
		defer close(tokens)
		tokens <- StreamToken{Content: "Test"}
		tokens <- StreamToken{Done: true}
	}()
	return tokens, nil
}

func (m *mockProvider) TestConnection(ctx context.Context) error {
	return m.testConnectError
}

func (m *mockProvider) GetModelInfo(ctx context.Context, modelID string) (*ModelInfo, error) {
	for _, model := range m.models {
		if model.ID == modelID {
			return &ModelInfo{
				Model:       model,
				Description: "Test model",
				Available:   true,
			}, nil
		}
	}
	return nil, ErrModelNotFound
}

func (m *mockProvider) SupportsStreaming() bool {
	return m.supportsStream
}

func (m *mockProvider) SupportsTools() bool {
	return m.supportsTools
}

var ErrModelNotFound = &ProviderError{
	Provider: "test",
	Code:     "NOT_FOUND",
	Message:  "model not found",
}

// TestRegistry tests the provider registry.
func TestRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		registry := NewRegistry()

		provider := &mockProvider{
			name: "test",
			models: []Model{
				{ID: "test-model", Name: "Test Model", Provider: "test"},
			},
		}

		err := registry.Register(provider)
		require.NoError(t, err)

		retrieved, err := registry.Get("test")
		require.NoError(t, err)
		assert.Equal(t, "test", retrieved.Name())
	})

	t.Run("Duplicate registration", func(t *testing.T) {
		registry := NewRegistry()

		provider1 := &mockProvider{name: "test"}
		provider2 := &mockProvider{name: "test"}

		err := registry.Register(provider1)
		require.NoError(t, err)

		err = registry.Register(provider2)
		assert.Error(t, err)
	})

	t.Run("Get non-existent provider", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.Get("nonexistent")
		assert.Error(t, err)
	})

	t.Run("List providers", func(t *testing.T) {
		registry := NewRegistry()

		provider1 := &mockProvider{name: "provider1"}
		provider2 := &mockProvider{name: "provider2"}

		_ = registry.Register(provider1)
		_ = registry.Register(provider2)

		names := registry.List()
		assert.Len(t, names, 2)
		assert.Contains(t, names, "provider1")
		assert.Contains(t, names, "provider2")
	})

	t.Run("TestAll", func(t *testing.T) {
		registry := NewRegistry()

		provider1 := &mockProvider{name: "provider1"}
		provider2 := &mockProvider{name: "provider2", testConnectError: assert.AnError}

		_ = registry.Register(provider1)
		_ = registry.Register(provider2)

		results := registry.TestAll(context.Background())
		assert.Len(t, results, 2)
		assert.NoError(t, results["provider1"])
		assert.Error(t, results["provider2"])
	})
}

// TestProviderError tests the ProviderError type.
func TestProviderError(t *testing.T) {
	err := &ProviderError{
		Provider:  "test",
		Code:      "TEST_ERROR",
		Message:   "test error message",
		Retryable: true,
	}

	assert.Equal(t, "test: test error message", err.Error())
	assert.True(t, err.IsRetryable())

	err.Retryable = false
	assert.False(t, err.IsRetryable())
}

// TestCompletionRequest tests the CompletionRequest type.
func TestCompletionRequest(t *testing.T) {
	temp := 0.7
	topP := 0.9
	topK := 40

	req := &CompletionRequest{
		Model: "test/model",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		System:      "You are a helpful assistant",
		Temperature: &temp,
		TopP:        &topP,
		TopK:        &topK,
		MaxTokens:   1000,
		StopSequences: []string{"STOP"},
		Metadata: map[string]string{
			"session": "test-session",
		},
	}

	assert.Equal(t, "test/model", req.Model)
	assert.Len(t, req.Messages, 1)
	assert.Equal(t, 0.7, *req.Temperature)
	assert.Equal(t, 0.9, *req.TopP)
	assert.Equal(t, 40, *req.TopK)
	assert.Equal(t, 1000, req.MaxTokens)
	assert.Contains(t, req.StopSequences, "STOP")
}

// TestMockProvider tests the mock provider implementation.
func TestMockProvider(t *testing.T) {
	provider := &mockProvider{
		name: "test",
		models: []Model{
			{ID: "test-1", Name: "Test Model 1", Provider: "test"},
			{ID: "test-2", Name: "Test Model 2", Provider: "test"},
		},
		supportsStream: true,
		supportsTools:  true,
	}

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "test", provider.Name())
	})

	t.Run("ListModels", func(t *testing.T) {
		models, err := provider.ListModels(context.Background())
		require.NoError(t, err)
		assert.Len(t, models, 2)
	})

	t.Run("CreateCompletion", func(t *testing.T) {
		req := &CompletionRequest{
			Model: "test-1",
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			MaxTokens: 100,
		}

		resp, err := provider.CreateCompletion(context.Background(), req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
		assert.Equal(t, 30, resp.Usage.TotalTokens)
	})

	t.Run("StreamCompletion", func(t *testing.T) {
		req := &CompletionRequest{
			Model: "test-1",
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
		}

		tokens, err := provider.StreamCompletion(context.Background(), req)
		require.NoError(t, err)

		var count int
		for token := range tokens {
			count++
			if token.Error != nil {
				t.Fatal(token.Error)
			}
		}
		assert.Equal(t, 2, count)
	})

	t.Run("TestConnection", func(t *testing.T) {
		err := provider.TestConnection(context.Background())
		assert.NoError(t, err)
	})

	t.Run("GetModelInfo", func(t *testing.T) {
		info, err := provider.GetModelInfo(context.Background(), "test-1")
		require.NoError(t, err)
		assert.Equal(t, "test-1", info.ID)
		assert.True(t, info.Available)
	})

	t.Run("SupportsStreaming", func(t *testing.T) {
		assert.True(t, provider.SupportsStreaming())
	})

	t.Run("SupportsTools", func(t *testing.T) {
		assert.True(t, provider.SupportsTools())
	})
}

// Benchmarks

func BenchmarkParseModelName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseModelName("anthropic/claude-sonnet-4-5")
	}
}

func BenchmarkFormatModelName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FormatModelName("anthropic", "claude-sonnet-4-5")
	}
}
