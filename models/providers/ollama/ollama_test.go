package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abrksh22/bplus/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	p := New()
	assert.NotNil(t, p)
	assert.Equal(t, defaultBaseURL, p.baseURL)
	assert.NotNil(t, p.client)
}

func TestNew_WithOptions(t *testing.T) {
	customClient := &http.Client{}
	customBaseURL := "http://custom:11434"

	p := New(
		WithBaseURL(customBaseURL),
		WithHTTPClient(customClient),
	)

	assert.Equal(t, customBaseURL, p.baseURL)
	assert.Equal(t, customClient, p.client)
}

func TestProvider_Name(t *testing.T) {
	p := New()
	assert.Equal(t, "ollama", p.Name())
}

func TestProvider_ListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/tags", r.URL.Path)

		response := map[string]interface{}{
			"models": []map[string]interface{}{
				{
					"name":        "llama3:latest",
					"modified_at": "2024-01-01T00:00:00Z",
					"size":        4000000000,
				},
				{
					"name":        "codellama:7b",
					"modified_at": "2024-01-01T00:00:00Z",
					"size":        3800000000,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p := New(WithBaseURL(server.URL))
	modelslist, err := p.ListModels(context.Background())

	require.NoError(t, err)
	require.Len(t, modelslist, 2)

	// Verify model fields
	assert.Equal(t, "llama3:latest", modelslist[0].ID)
	assert.Equal(t, "ollama", modelslist[0].Provider)
	assert.Greater(t, modelslist[0].ContextWindow, 0)

	assert.Equal(t, "codellama:7b", modelslist[1].ID)
}

func TestProvider_CreateCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/chat", r.URL.Path)

		var req struct {
			Model    string                   `json:"model"`
			Messages []map[string]interface{} `json:"messages"`
			Stream   bool                     `json:"stream"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "llama3:latest", req.Model)
		assert.False(t, req.Stream)
		assert.Len(t, req.Messages, 1)

		response := map[string]interface{}{
			"model": "llama3:latest",
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": "Hello! How can I help you today?",
			},
			"done": true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	p := New(WithBaseURL(server.URL))

	req := &models.CompletionRequest{
		Model: "llama3:latest",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 100,
	}

	resp, err := p.CreateCompletion(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Hello! How can I help you today?", resp.Content)
	assert.Equal(t, "llama3:latest", resp.Model)
}

func TestProvider_TestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/version" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"version": "0.1.0"})
		}
	}))
	defer server.Close()

	p := New(WithBaseURL(server.URL))
	err := p.TestConnection(context.Background())
	assert.NoError(t, err)
}

func TestProvider_ErrorHandling(t *testing.T) {
	t.Run("Server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		p := New(WithBaseURL(server.URL))

		req := &models.CompletionRequest{
			Model:    "llama3:latest",
			Messages: []models.Message{{Role: "user", Content: "Hello"}},
		}

		_, err := p.CreateCompletion(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("Context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				return
			}
		}))
		defer server.Close()

		p := New(WithBaseURL(server.URL))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		req := &models.CompletionRequest{
			Model:    "llama3:latest",
			Messages: []models.Message{{Role: "user", Content: "Hello"}},
		}

		_, err := p.CreateCompletion(ctx, req)
		assert.Error(t, err)
	})
}
