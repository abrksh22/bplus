package context

import (
	"context"
	"sort"
)

// optimizeWithSummarization uses LLM to summarize verbose content
func (m *Manager) optimizeWithSummarization(ctx context.Context, items []ContextItem) ([]ContextItem, error) {
	optimized := make([]ContextItem, 0, len(items))
	currentTokens := m.calculateTotalTokens(items)

	// If under target, no optimization needed
	if currentTokens <= m.config.TargetTokens {
		return items, nil
	}

	// Summarize items that can be compressed
	for _, item := range items {
		// Skip items that should be preserved
		if m.shouldPreserve(item) {
			optimized = append(optimized, item)
			continue
		}

		// Summarize large items
		if item.TokenCount > 1000 && item.Type != TypeSummary {
			req := SummarizationRequest{
				Content:      item.Content,
				Type:         item.Type,
				TargetTokens: item.TokenCount / 2, // Aim for 50% reduction
			}

			result, err := m.summarizer.Summarize(ctx, req)
			if err != nil {
				m.logger.Warn("Summarization failed", "error", err, "item_id", item.ID)
				optimized = append(optimized, item)
				continue
			}

			// Create summarized item
			summarized := item
			summarized.Content = result.Summarized
			summarized.TokenCount = result.SummarySize
			summarized.Type = TypeSummary
			if summarized.Metadata == nil {
				summarized.Metadata = make(map[string]interface{})
			}
			summarized.Metadata["original_tokens"] = result.OriginalSize
			summarized.Metadata["compression"] = result.Compression

			optimized = append(optimized, summarized)
		} else {
			optimized = append(optimized, item)
		}
	}

	return optimized, nil
}

// optimizeWithPruning removes low-relevance items
func (m *Manager) optimizeWithPruning(items []ContextItem) ([]ContextItem, error) {
	// Sort by relevance score
	sorted := m.sortByRelevance(items)

	optimized := make([]ContextItem, 0, len(items))
	currentTokens := 0

	// Keep items until we hit target
	for _, item := range sorted {
		// Always preserve critical items
		if m.shouldPreserve(item) {
			optimized = append(optimized, item)
			currentTokens += item.TokenCount
			continue
		}

		// Skip low-relevance items if over target
		if currentTokens+item.TokenCount > m.config.TargetTokens {
			if item.Relevance < m.config.MinRelevance {
				m.logger.Debug("Pruning low-relevance item",
					"item_id", item.ID,
					"relevance", item.Relevance,
					"tokens", item.TokenCount)
				continue
			}
		}

		optimized = append(optimized, item)
		currentTokens += item.TokenCount

		if currentTokens >= m.config.TargetTokens {
			break
		}
	}

	return optimized, nil
}

// optimizeWithChunking groups related items together semantically
func (m *Manager) optimizeWithChunking(items []ContextItem) ([]ContextItem, error) {
	// Group items by type
	chunks := make(map[ContextItemType][]ContextItem)
	for _, item := range items {
		chunks[item.Type] = append(chunks[item.Type], item)
	}

	optimized := make([]ContextItem, 0, len(items))

	// Process each chunk
	for itemType, chunkItems := range chunks {
		// Sort chunk by relevance
		sort.Slice(chunkItems, func(i, j int) bool {
			return m.calculateRelevanceScore(chunkItems[i]) > m.calculateRelevanceScore(chunkItems[j])
		})

		// Keep most relevant items from each chunk
		for _, item := range chunkItems {
			if m.shouldPreserve(item) || item.Relevance >= m.config.MinRelevance {
				optimized = append(optimized, item)
			}
		}

		m.logger.Debug("Chunking result",
			"type", itemType,
			"original", len(chunkItems),
			"kept", len(optimized))
	}

	return optimized, nil
}

// optimizeWithTiering moves items between hot/warm/cold tiers
func (m *Manager) optimizeWithTiering(items []ContextItem) ([]ContextItem, error) {
	optimized := make([]ContextItem, len(items))
	copy(optimized, items)

	hotLimit := m.config.MaxTokens / 4  // 25% in hot
	warmLimit := m.config.MaxTokens / 2 // 50% in warm
	// Rest goes to cold

	hotTokens := 0
	warmTokens := 0

	// Sort by relevance
	sorted := m.sortByRelevance(optimized)

	// Assign tiers based on relevance and size
	for i := range sorted {
		item := &sorted[i]

		// Critical items stay in hot tier
		if m.shouldPreserve(*item) {
			item.Tier = TierHot
			hotTokens += item.TokenCount
			continue
		}

		// Assign based on available space
		if hotTokens+item.TokenCount <= hotLimit {
			item.Tier = TierHot
			hotTokens += item.TokenCount
		} else if warmTokens+item.TokenCount <= warmLimit {
			item.Tier = TierWarm
			warmTokens += item.TokenCount
		} else {
			item.Tier = TierCold
		}
	}

	m.logger.Debug("Tiering complete",
		"hot_tokens", hotTokens,
		"warm_tokens", warmTokens,
		"hot_limit", hotLimit,
		"warm_limit", warmLimit)

	return sorted, nil
}

// optimizeBalanced uses a combination of strategies
func (m *Manager) optimizeBalanced(ctx context.Context, items []ContextItem) ([]ContextItem, error) {
	currentTokens := m.calculateTotalTokens(items)

	// If under target, just optimize tiers
	if currentTokens <= m.config.TargetTokens {
		return m.optimizeWithTiering(items)
	}

	// Step 1: Tier items
	tiered, err := m.optimizeWithTiering(items)
	if err != nil {
		return nil, err
	}

	currentTokens = m.calculateTotalTokens(tiered)

	// Step 2: If still over target, prune low-relevance items
	if currentTokens > m.config.TargetTokens {
		pruned, err := m.optimizeWithPruning(tiered)
		if err != nil {
			return nil, err
		}
		tiered = pruned
		currentTokens = m.calculateTotalTokens(tiered)
	}

	// Step 3: If still over target, summarize large items
	if currentTokens > m.config.TargetTokens {
		summarized, err := m.optimizeWithSummarization(ctx, tiered)
		if err != nil {
			return nil, err
		}
		tiered = summarized
	}

	return tiered, nil
}

// shouldPreserve checks if an item should always be preserved
func (m *Manager) shouldPreserve(item ContextItem) bool {
	// Check if type is in preserve list
	for _, preserveType := range m.config.PreserveTypes {
		if item.Type == preserveType {
			return true
		}
	}

	// Always preserve user intent
	if item.Type == TypeUserIntent {
		return true
	}

	// Preserve high-relevance items
	if item.Relevance >= 0.9 {
		return true
	}

	return false
}
