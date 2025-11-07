package context

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
	"github.com/abrksh22/bplus/internal/storage"
	"github.com/abrksh22/bplus/models"
)

// Default optimization configuration values
const (
	DefaultMaxTokens         = 200000
	DefaultTargetTokens      = 100000
	DefaultMinRelevance      = 0.3
	DefaultOptimizeThreshold = 0.8
)

// Manager manages context optimization and persistence
type Manager struct {
	db            *storage.SQLiteDB
	logger        *logging.Logger
	config        OptimizationConfig
	summarizer    *Summarizer
	checkpointer  *CheckpointManager

	// In-memory context cache
	mu            sync.RWMutex
	hotContext    map[string][]ContextItem // sessionID -> items
	metrics       map[string]*ContextMetrics
}

// NewManager creates a new context manager
func NewManager(db *storage.SQLiteDB, provider models.Provider, config OptimizationConfig) (*Manager, error) {
	if config.MaxTokens == 0 {
		config.MaxTokens = DefaultMaxTokens
	}
	if config.TargetTokens == 0 {
		config.TargetTokens = config.MaxTokens / 2 // Default to 50%
	}
	if config.MinRelevance == 0 {
		config.MinRelevance = DefaultMinRelevance
	}
	if config.OptimizeThreshold == 0 {
		config.OptimizeThreshold = DefaultOptimizeThreshold
	}

	summarizer := NewSummarizer(provider, config.SummarizationModel)
	checkpointer := NewCheckpointManager(db)

	return &Manager{
		db:           db,
		logger:       logging.NewDefaultLogger().WithComponent("context_manager"),
		config:       config,
		summarizer:   summarizer,
		checkpointer: checkpointer,
		hotContext:   make(map[string][]ContextItem),
		metrics:      make(map[string]*ContextMetrics),
	}, nil
}

// AddItem adds a new context item
func (m *Manager) AddItem(sessionID string, item ContextItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if item.ID == "" {
		item.ID = generateItemID()
	}
	if item.LastAccessed.IsZero() {
		item.LastAccessed = time.Now()
	}
	if item.Relevance == 0 {
		item.Relevance = 1.0 // New items are highly relevant
	}

	// Add to hot context
	m.hotContext[sessionID] = append(m.hotContext[sessionID], item)

	// Update metrics
	m.updateMetrics(sessionID)

	// Check if optimization is needed
	if m.config.AutoOptimize {
		if err := m.checkAndOptimize(sessionID); err != nil {
			m.logger.Warn("Auto-optimization failed", "error", err)
		}
	}

	return nil
}

// GetContext retrieves the current context for a session
func (m *Manager) GetContext(sessionID string, maxTokens int) ([]ContextItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items, ok := m.hotContext[sessionID]
	if !ok {
		// Load from database
		var err error
		items, err = m.loadContext(sessionID)
		if err != nil {
			return nil, err
		}
	}

	// Sort by relevance and recency
	sorted := m.sortByRelevance(items)

	// Select items that fit within token limit
	selected := make([]ContextItem, 0)
	totalTokens := 0

	for _, item := range sorted {
		if totalTokens+item.TokenCount > maxTokens {
			break
		}
		selected = append(selected, item)
		totalTokens += item.TokenCount
	}

	// Update last accessed time
	m.updateAccessTime(sessionID, selected)

	return selected, nil
}

// OptimizeContext optimizes context for a session
func (m *Manager) OptimizeContext(ctx context.Context, sessionID string) (*ContextMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Optimizing context", "session", sessionID, "strategy", m.config.Strategy)

	items, ok := m.hotContext[sessionID]
	if !ok {
		return nil, errors.New(errors.ErrCodeInvalidInput, "session not found")
	}

	var optimized []ContextItem
	var err error

	switch m.config.Strategy {
	case StrategyAggressiveSummarization:
		optimized, err = m.optimizeWithSummarization(ctx, items)
	case StrategySelectivePruning:
		optimized, err = m.optimizeWithPruning(items)
	case StrategySemanticChunking:
		optimized, err = m.optimizeWithChunking(items)
	case StrategyTieredEviction:
		optimized, err = m.optimizeWithTiering(items)
	case StrategyBalanced:
		optimized, err = m.optimizeBalanced(ctx, items)
	default:
		optimized, err = m.optimizeBalanced(ctx, items)
	}

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "optimization failed")
	}

	// Update hot context
	m.hotContext[sessionID] = optimized

	// Update metrics
	metrics := m.updateMetrics(sessionID)

	m.logger.Info("Context optimized",
		"session", sessionID,
		"before_tokens", m.calculateTotalTokens(items),
		"after_tokens", m.calculateTotalTokens(optimized),
		"reduction", fmt.Sprintf("%.1f%%", metrics.Efficiency*100))

	return metrics, nil
}

// GetMetrics returns current context metrics for a session
func (m *Manager) GetMetrics(sessionID string) (*ContextMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, ok := m.metrics[sessionID]
	if !ok {
		return nil, errors.New(errors.ErrCodeInvalidInput, "session not found")
	}

	return metrics, nil
}

// CreateSnapshot creates a snapshot of current context
func (m *Manager) CreateSnapshot(sessionID string) (*ContextSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items, ok := m.hotContext[sessionID]
	if !ok {
		return nil, errors.New(errors.ErrCodeInvalidInput, "session not found")
	}

	snapshot := &ContextSnapshot{
		SessionID:    sessionID,
		Timestamp:    time.Now(),
		Items:        make([]ContextItem, len(items)),
		TotalTokens:  m.calculateTotalTokens(items),
		Optimization: m.metrics[sessionID].Efficiency,
	}

	copy(snapshot.Items, items)

	return snapshot, nil
}

// RestoreSnapshot restores context from a snapshot
func (m *Manager) RestoreSnapshot(snapshot *ContextSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hotContext[snapshot.SessionID] = snapshot.Items
	m.updateMetrics(snapshot.SessionID)

	m.logger.Info("Context restored from snapshot",
		"session", snapshot.SessionID,
		"tokens", snapshot.TotalTokens)

	return nil
}

// ClearSession clears context for a session
func (m *Manager) ClearSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.hotContext, sessionID)
	delete(m.metrics, sessionID)

	m.logger.Info("Session context cleared", "session", sessionID)
	return nil
}

// checkAndOptimize checks if optimization is needed and runs it
func (m *Manager) checkAndOptimize(sessionID string) error {
	metrics := m.metrics[sessionID]
	if metrics == nil {
		return nil
	}

	utilizationPct := float64(metrics.CurrentSize) / float64(m.config.MaxTokens)

	if utilizationPct >= m.config.OptimizeThreshold {
		m.logger.Info("Auto-optimization triggered",
			"session", sessionID,
			"utilization", fmt.Sprintf("%.1f%%", utilizationPct*100))

		_, err := m.OptimizeContext(context.Background(), sessionID)
		return err
	}

	return nil
}

// updateMetrics updates context metrics for a session
func (m *Manager) updateMetrics(sessionID string) *ContextMetrics {
	items, ok := m.hotContext[sessionID]
	if !ok {
		return nil
	}

	metrics := &ContextMetrics{
		CurrentSize: m.calculateTotalTokens(items),
		MaxSize:     m.config.MaxTokens,
		LastOptimized: time.Now(),
	}

	// Calculate tier sizes
	for _, item := range items {
		switch item.Tier {
		case TierHot:
			metrics.HotTierSize += item.TokenCount
		case TierWarm:
			metrics.WarmTierSize += item.TokenCount
		case TierCold:
			metrics.ColdTierSize += item.TokenCount
		}
	}

	// Calculate efficiency (how much we've optimized)
	if metrics.CurrentSize < m.config.MaxTokens {
		metrics.Efficiency = 1.0 - float64(metrics.CurrentSize)/float64(m.config.MaxTokens)
	}

	m.metrics[sessionID] = metrics
	return metrics
}

// sortByRelevance sorts context items by relevance and recency
func (m *Manager) sortByRelevance(items []ContextItem) []ContextItem {
	sorted := make([]ContextItem, len(items))
	copy(sorted, items)

	// Simple bubble sort by relevance score
	// Relevance score = relevance * 0.7 + recency * 0.3
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			scoreI := m.calculateRelevanceScore(sorted[i])
			scoreJ := m.calculateRelevanceScore(sorted[j])
			if scoreI < scoreJ {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// calculateRelevanceScore calculates a composite relevance score
func (m *Manager) calculateRelevanceScore(item ContextItem) float64 {
	// Recency score (newer = higher)
	hoursSinceAccess := time.Since(item.LastAccessed).Hours()
	recencyScore := 1.0 / (1.0 + hoursSinceAccess/24.0) // Decay over days

	// Combine relevance and recency
	return item.Relevance*0.7 + recencyScore*0.3
}

// calculateTotalTokens calculates total tokens in items
func (m *Manager) calculateTotalTokens(items []ContextItem) int {
	total := 0
	for _, item := range items {
		total += item.TokenCount
	}
	return total
}

// updateAccessTime updates last accessed time for items
func (m *Manager) updateAccessTime(sessionID string, items []ContextItem) {
	now := time.Now()
	itemMap := make(map[string]bool)
	for _, item := range items {
		itemMap[item.ID] = true
	}

	// Update in hot context
	if hotItems, ok := m.hotContext[sessionID]; ok {
		for i := range hotItems {
			if itemMap[hotItems[i].ID] {
				hotItems[i].LastAccessed = now
			}
		}
	}
}

// loadContext loads context from database
func (m *Manager) loadContext(sessionID string) ([]ContextItem, error) {
	// TODO: Implement loading from database
	// For now, return empty
	return make([]ContextItem, 0), nil
}

// generateItemID generates a unique item ID
func generateItemID() string {
	return fmt.Sprintf("item_%d", time.Now().UnixNano())
}
