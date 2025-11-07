package context

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/abrksh22/bplus/internal/storage"
	"github.com/abrksh22/bplus/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProvider implements a mock LLM provider for testing
type MockProvider struct {
	responses []string
	callCount int
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) ListModels(ctx context.Context) ([]models.Model, error) {
	return []models.Model{}, nil
}

func (m *MockProvider) CreateCompletion(ctx context.Context, req models.CompletionRequest) (*models.CompletionResponse, error) {
	response := "This is a summary."
	if m.callCount < len(m.responses) {
		response = m.responses[m.callCount]
	}
	m.callCount++

	return &models.CompletionResponse{
		Choices: []models.Choice{
			{
				Message: models.Message{
					Role:    "assistant",
					Content: response,
				},
			},
		},
		Usage: models.Usage{
			InputTokens:  10,
			OutputTokens: 5,
			TotalTokens:  15,
		},
	}, nil
}

func (m *MockProvider) StreamCompletion(ctx context.Context, req models.CompletionRequest) (<-chan models.StreamChunk, error) {
	ch := make(chan models.StreamChunk)
	close(ch)
	return ch, nil
}

func (m *MockProvider) TestConnection(ctx context.Context) error {
	return nil
}

func (m *MockProvider) GetModelInfo(modelID string) (*models.ModelInfo, error) {
	return &models.ModelInfo{
		ID:            modelID,
		Name:          "Mock Model",
		ContextWindow: 128000,
	}, nil
}

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) *storage.SQLiteDB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := storage.NewSQLiteDB(dbPath)
	require.NoError(t, err)

	return db
}

func TestContextManager_AddItem(t *testing.T) {
	db := setupTestDB(t)
	provider := &MockProvider{}

	config := OptimizationConfig{
		MaxTokens:    10000,
		TargetTokens: 5000,
		AutoOptimize: false,
	}

	manager, err := NewManager(db, provider, config)
	require.NoError(t, err)

	sessionID := "test-session-1"

	item := ContextItem{
		Type:       TypeMessage,
		Content:    "Test message content",
		TokenCount: 100,
		Relevance:  1.0,
	}

	err = manager.AddItem(sessionID, item)
	assert.NoError(t, err)

	// Get context
	items, err := manager.GetContext(sessionID, 10000)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Test message content", items[0].Content)
}

func TestContextManager_OptimizeWithPruning(t *testing.T) {
	db := setupTestDB(t)
	provider := &MockProvider{}

	config := OptimizationConfig{
		Strategy:     StrategySelectivePruning,
		MaxTokens:    1000,
		TargetTokens: 500,
		MinRelevance: 0.5,
	}

	manager, err := NewManager(db, provider, config)
	require.NoError(t, err)

	sessionID := "test-session-2"

	// Add items with different relevance scores
	items := []ContextItem{
		{Type: TypeMessage, Content: "High relevance", TokenCount: 200, Relevance: 0.9},
		{Type: TypeMessage, Content: "Medium relevance", TokenCount: 200, Relevance: 0.6},
		{Type: TypeMessage, Content: "Low relevance", TokenCount: 200, Relevance: 0.3},
		{Type: TypeUserIntent, Content: "User intent", TokenCount: 200, Relevance: 1.0},
	}

	for _, item := range items {
		err := manager.AddItem(sessionID, item)
		require.NoError(t, err)
	}

	// Optimize
	metrics, err := manager.OptimizeContext(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	// Should have pruned low-relevance item
	optimized, err := manager.GetContext(sessionID, 10000)
	assert.NoError(t, err)
	assert.Less(t, len(optimized), len(items))
}

func TestContextManager_OptimizeWithTiering(t *testing.T) {
	db := setupTestDB(t)
	provider := &MockProvider{}

	config := OptimizationConfig{
		Strategy:     StrategyTieredEviction,
		MaxTokens:    10000,
		TargetTokens: 5000,
	}

	manager, err := NewManager(db, provider, config)
	require.NoError(t, err)

	sessionID := "test-session-3"

	// Add items
	items := []ContextItem{
		{Type: TypeUserIntent, Content: "Critical", TokenCount: 100, Relevance: 1.0},
		{Type: TypeMessage, Content: "Important", TokenCount: 200, Relevance: 0.8},
		{Type: TypeToolResult, Content: "Normal", TokenCount: 300, Relevance: 0.5},
		{Type: TypeMessage, Content: "Less important", TokenCount: 400, Relevance: 0.3},
	}

	for _, item := range items {
		err := manager.AddItem(sessionID, item)
		require.NoError(t, err)
	}

	// Optimize with tiering
	_, err = manager.OptimizeContext(context.Background(), sessionID)
	assert.NoError(t, err)

	// Check tiers
	optimized, err := manager.GetContext(sessionID, 10000)
	assert.NoError(t, err)

	// User intent should be in hot tier
	for _, item := range optimized {
		if item.Type == TypeUserIntent {
			assert.Equal(t, TierHot, item.Tier)
		}
	}
}

func TestSummarizer_Summarize(t *testing.T) {
	provider := &MockProvider{
		responses: []string{"This is a concise summary of the content."},
	}

	summarizer := NewSummarizer(provider, "mock-model")

	req := SummarizationRequest{
		Content:      "This is a very long piece of content that needs to be summarized. It contains lots of details and information that should be condensed into a shorter form while preserving the key points.",
		Type:         TypeMessage,
		TargetTokens: 20,
	}

	result, err := summarizer.Summarize(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.OriginalSize, result.SummarySize)
	assert.Greater(t, result.Compression, 0.0)
}

func TestCheckpointManager_CreateAndRestore(t *testing.T) {
	db := setupTestDB(t)
	manager := NewCheckpointManager(db)

	sessionID := "test-session-4"
	snapshot := &ContextSnapshot{
		SessionID:   sessionID,
		Timestamp:   time.Now(),
		TotalTokens: 5000,
		Items: []ContextItem{
			{Type: TypeMessage, Content: "Test", TokenCount: 100},
		},
	}

	messages := []models.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	// Create checkpoint
	checkpoint, err := manager.CreateCheckpoint(context.Background(), sessionID, "test-checkpoint", "Test description", snapshot, messages)
	assert.NoError(t, err)
	assert.NotEmpty(t, checkpoint.ID)

	// Restore checkpoint
	restored, err := manager.RestoreCheckpoint(context.Background(), checkpoint.ID)
	assert.NoError(t, err)
	assert.Equal(t, checkpoint.ID, restored.ID)
	assert.Equal(t, sessionID, restored.SessionID)
	assert.Equal(t, "test-checkpoint", restored.Name)
	assert.Len(t, restored.Messages, 2)
}

func TestCheckpointManager_ListAndDelete(t *testing.T) {
	db := setupTestDB(t)
	manager := NewCheckpointManager(db)

	sessionID := "test-session-5"
	snapshot := &ContextSnapshot{
		SessionID:   sessionID,
		Timestamp:   time.Now(),
		TotalTokens: 1000,
	}

	// Create multiple checkpoints
	for i := 0; i < 3; i++ {
		_, err := manager.CreateCheckpoint(context.Background(), sessionID, "checkpoint", "desc", snapshot, nil)
		require.NoError(t, err)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// List checkpoints
	checkpoints, err := manager.ListCheckpoints(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Len(t, checkpoints, 3)

	// Delete one
	err = manager.DeleteCheckpoint(context.Background(), checkpoints[0].ID)
	assert.NoError(t, err)

	// Verify deletion
	checkpoints, err = manager.ListCheckpoints(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Len(t, checkpoints, 2)
}

func TestSessionExporter_ExportAndImport(t *testing.T) {
	exporter := NewSessionExporter()

	sessionID := "test-session-6"
	snapshot := &ContextSnapshot{
		SessionID:   sessionID,
		Timestamp:   time.Now(),
		TotalTokens: 2000,
	}

	messages := []models.Message{
		{Role: "user", Content: "Export test"},
		{Role: "assistant", Content: "This is a test"},
	}

	opts := ExportOptions{
		IncludeMessages: true,
		IncludeContext:  true,
	}

	// Export
	export, err := exporter.ExportSession(context.Background(), sessionID, snapshot, messages, opts)
	assert.NoError(t, err)
	assert.NotNil(t, export)
	assert.Equal(t, "1.0", export.Version)
	assert.Len(t, export.Messages, 2)

	// Export to file
	tmpFile := filepath.Join(t.TempDir(), "export.json")
	err = exporter.ExportToFile(context.Background(), export, tmpFile)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tmpFile)
	assert.NoError(t, err)

	// Import from file
	imported, err := exporter.ImportFromFile(context.Background(), tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, sessionID, imported.Session.ID)
	assert.Len(t, imported.Messages, 2)
}

func TestContextMetrics(t *testing.T) {
	db := setupTestDB(t)
	provider := &MockProvider{}

	config := OptimizationConfig{
		MaxTokens:    10000,
		TargetTokens: 5000,
	}

	manager, err := NewManager(db, provider, config)
	require.NoError(t, err)

	sessionID := "test-session-7"

	// Add items
	for i := 0; i < 5; i++ {
		item := ContextItem{
			Type:       TypeMessage,
			Content:    "Test content",
			TokenCount: 500,
			Relevance:  0.8,
		}
		err := manager.AddItem(sessionID, item)
		require.NoError(t, err)
	}

	// Get metrics
	metrics, err := manager.GetMetrics(sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 2500, metrics.CurrentSize)
	assert.Equal(t, 10000, metrics.MaxSize)
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "Empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "Short text",
			text:     "Hello",
			expected: 1,
		},
		{
			name:     "Medium text",
			text:     "This is a test message",
			expected: 5,
		},
		{
			name:     "Long text",
			text:     "This is a much longer test message that should have more tokens estimated based on the character count.",
			expected: 27,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := estimateTokens(tt.text)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}
