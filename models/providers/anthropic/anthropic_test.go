package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/abrksh22/bplus/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	p := New("test-api-key")
	assert.NotNil(t, p)
	assert.Equal(t, "test-api-key", p.apiKey)
	assert.Equal(t, defaultBaseURL, p.baseURL)
	assert.NotNil(t, p.client)
}

func TestNew_WithOptions(t *testing.T) {
	customClient := &http.Client{}
	customBaseURL := "https://custom.api.com"

	p := New("test-key",
		WithBaseURL(customBaseURL),
		WithHTTPClient(customClient),
	)

	assert.Equal(t, customBaseURL, p.baseURL)
	assert.Equal(t, customClient, p.client)
}

func TestProvider_Name(t *testing.T) {
	p := New("test-key")
	assert.Equal(t, "anthropic", p.Name())
}

func TestProvider_SupportsStreaming(t *testing.T) {
	p := New("test-key")
	assert.True(t, p.SupportsStreaming())
}

func TestProvider_SupportsTools(t *testing.T) {
	p := New("test-key")
	assert.True(t, p.SupportsTools())
}

func TestProvider_ListModels(t *testing.T) {
	p := New("test-key")
	modelslist, err := p.ListModels(context.Background())

	require.NoError(t, err)
	require.Len(t, modelslist, 3) // Opus, Sonnet, Haiku

	// Verify all models have required fields
	for _, model := range modelslist {
		assert.NotEmpty(t, model.ID)
		assert.NotEmpty(t, model.Name)
		assert.Equal(t, "anthropic", model.Provider)
		assert.Greater(t, model.ContextWindow, 0)
		assert.Greater(t, model.MaxOutput, 0)
		assert.Contains(t, model.Capabilities, "streaming")
		assert.Contains(t, model.Capabilities, "tools")
	}

	// Verify specific models
	opus := modelslist[0]
	assert.Equal(t, "claude-opus-4-1", opus.ID)
	assert.Equal(t, 200000, opus.ContextWindow)

	sonnet := modelslist[1]
	assert.Equal(t, "claude-sonnet-4-5", sonnet.ID)

	haiku := modelslist[2]
	assert.Equal(t, "claude-haiku-4-0", haiku.ID)
}

func TestProvider_GetModelInfo(t *testing.T) {
	p := New("test-key")

	t.Run("Valid model ID", func(t *testing.T) {
		info, err := p.GetModelInfo(context.Background(), "claude-sonnet-4-5")
		require.NoError(t, err)
		assert.Equal(t, "claude-sonnet-4-5", info.ID)
		assert.Equal(t, "Claude Sonnet 4.5", info.Name)
	})

	t.Run("Invalid model ID", func(t *testing.T) {
		_, err := p.GetModelInfo(context.Background(), "nonexistent-model")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model not found")
	})
}

func TestProvider_CreateCompletion(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/messages", r.URL.Path)
		assert.Equal(t, apiVersion, r.Header.Get("anthropic-version"))
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))

		// Send mock response
		response := map[string]interface{}{
			"id":    "msg_123",
			"type":  "message",
			"role":  "assistant",
			"model": "claude-sonnet-4-5",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Hello! How can I help you?",
				},
			},
			"usage": map[string]interface{}{
				"input_tokens":  10,
				"output_tokens": 20,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p := New("test-api-key", WithBaseURL(server.URL))

	req := &models.CompletionRequest{
		Model: "claude-sonnet-4-5",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 100,
	}

	resp, err := p.CreateCompletion(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Hello! How can I help you?", resp.Content)
	assert.Equal(t, 10, resp.Usage.InputTokens)
	assert.Equal(t, 20, resp.Usage.OutputTokens)
	assert.Equal(t, "claude-sonnet-4-5", resp.Model)
}

func TestProvider_CreateCompletion_WithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			System string `json:"system"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "You are a helpful assistant.", req.System)

		response := map[string]interface{}{
			"id":   "msg_123",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "text", "text": "I'm ready to help!"},
			},
			"usage": map[string]interface{}{
				"input_tokens":  5,
				"output_tokens": 10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p := New("test-api-key", WithBaseURL(server.URL))

	req := &models.CompletionRequest{
		Model:  "claude-sonnet-4-5",
		System: "You are a helpful assistant.",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 100,
	}

	resp, err := p.CreateCompletion(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "I'm ready to help!", resp.Content)
}

func TestProvider_CreateCompletion_ErrorHandling(t *testing.T) {
	t.Run("API error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"Invalid API key"}}`))
		}))
		defer server.Close()

		p := New("invalid-key", WithBaseURL(server.URL))

		req := &models.CompletionRequest{
			Model:    "claude-sonnet-4-5",
			Messages: []models.Message{{Role: "user", Content: "Hello"}},
		}

		_, err := p.CreateCompletion(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid API key")
	})

	t.Run("Missing API key", func(t *testing.T) {
		p := New("")
		req := &models.CompletionRequest{
			Model:    "claude-sonnet-4-5",
			Messages: []models.Message{{Role: "user", Content: "Hello"}},
		}

		_, err := p.CreateCompletion(context.Background(), req)
		assert.Error(t, err)
		// Accept any error related to authentication/API key
		assert.True(t, strings.Contains(err.Error(), "API key") ||
			strings.Contains(err.Error(), "authentication") ||
			strings.Contains(err.Error(), "x-api-key"))
	})

	t.Run("Context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			select {
			case <-r.Context().Done():
				return
			}
		}))
		defer server.Close()

		p := New("test-key", WithBaseURL(server.URL))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := &models.CompletionRequest{
			Model:    "claude-sonnet-4-5",
			Messages: []models.Message{{Role: "user", Content: "Hello"}},
		}

		_, err := p.CreateCompletion(ctx, req)
		assert.Error(t, err)
	})
}

func TestProvider_TestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"id":   "msg_test",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "text", "text": "test"},
			},
			"usage": map[string]interface{}{
				"input_tokens":  1,
				"output_tokens": 1,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p := New("test-key", WithBaseURL(server.URL))
	err := p.TestConnection(context.Background())
	assert.NoError(t, err)
}
