package router

import (
	"fmt"
	"sync"
	"time"
)

// CostTracker tracks API costs and enforces budget limits.
// This is a scaffold for Phase 4.4 - full implementation will come in later phases.
type CostTracker struct {
	daily   float64
	monthly float64
	spent   map[string]float64 // model -> cost
	budgets BudgetLimits
	mu      sync.RWMutex
}

// BudgetLimits defines budget constraints.
type BudgetLimits struct {
	DailyMax   float64
	MonthlyMax float64
	WarnAt     float64 // Percentage at which to warn (e.g., 0.8 for 80%)
}

// NewCostTracker creates a new cost tracker.
func NewCostTracker(limits BudgetLimits) *CostTracker {
	return &CostTracker{
		daily:   0,
		monthly: 0,
		spent:   make(map[string]float64),
		budgets: limits,
	}
}

// RecordCost records the cost of a model API call.
// Placeholder for future implementation.
func (c *CostTracker) RecordCost(modelID string, inputTokens, outputTokens int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Placeholder - will calculate actual cost based on model pricing
	// For now, just track that the method exists
	cost := 0.0 // Will calculate: inputTokens * inputPrice + outputTokens * outputPrice

	c.spent[modelID] += cost
	c.daily += cost
	c.monthly += cost

	// Check if budget exceeded
	if c.budgets.DailyMax > 0 && c.daily > c.budgets.DailyMax {
		return fmt.Errorf("daily budget exceeded: $%.2f / $%.2f", c.daily, c.budgets.DailyMax)
	}
	if c.budgets.MonthlyMax > 0 && c.monthly > c.budgets.MonthlyMax {
		return fmt.Errorf("monthly budget exceeded: $%.2f / $%.2f", c.monthly, c.budgets.MonthlyMax)
	}

	return nil
}

// GetSpent returns the amount spent on a specific model.
func (c *CostTracker) GetSpent(modelID string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spent[modelID]
}

// GetTotalSpent returns the total amount spent.
func (c *CostTracker) GetTotalSpent() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.monthly
}

// GetDailySpent returns the amount spent today.
func (c *CostTracker) GetDailySpent() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.daily
}

// Exceeded returns whether any budget limit has been exceeded.
func (c *CostTracker) Exceeded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.budgets.DailyMax > 0 && c.daily >= c.budgets.DailyMax {
		return true
	}
	if c.budgets.MonthlyMax > 0 && c.monthly >= c.budgets.MonthlyMax {
		return true
	}
	return false
}

// ProjectCost estimates the cost of a request.
// Placeholder for future implementation.
func (c *CostTracker) ProjectCost(modelID string, estimatedTokens int) float64 {
	// Placeholder - will calculate based on model pricing and estimated tokens
	// For now, return 0
	return 0.0
}

// ShouldWarn returns whether we should warn about approaching budget limits.
func (c *CostTracker) ShouldWarn() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.budgets.WarnAt <= 0 {
		c.budgets.WarnAt = 0.8 // Default to 80%
	}

	if c.budgets.DailyMax > 0 && c.daily >= c.budgets.DailyMax*c.budgets.WarnAt {
		return true
	}
	if c.budgets.MonthlyMax > 0 && c.monthly >= c.budgets.MonthlyMax*c.budgets.WarnAt {
		return true
	}
	return false
}

// Reset resets the cost tracker (called daily/monthly).
// Placeholder for future implementation with proper time-based reset logic.
func (c *CostTracker) Reset(scope string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch scope {
	case "daily":
		c.daily = 0
	case "monthly":
		c.monthly = 0
		c.spent = make(map[string]float64)
	}
}

// GetBudgetStatus returns the current budget status.
func (c *CostTracker) GetBudgetStatus() BudgetStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := BudgetStatus{
		DailySpent:       c.daily,
		DailyLimit:       c.budgets.DailyMax,
		MonthlySpent:     c.monthly,
		MonthlyLimit:     c.budgets.MonthlyMax,
		DailyRemaining:   c.budgets.DailyMax - c.daily,
		MonthlyRemaining: c.budgets.MonthlyMax - c.monthly,
		LastUpdated:      time.Now(),
	}

	if c.budgets.DailyMax > 0 {
		status.DailyPercentage = (c.daily / c.budgets.DailyMax) * 100
	}
	if c.budgets.MonthlyMax > 0 {
		status.MonthlyPercentage = (c.monthly / c.budgets.MonthlyMax) * 100
	}

	return status
}

// BudgetStatus contains budget status information.
type BudgetStatus struct {
	DailySpent        float64
	DailyLimit        float64
	DailyRemaining    float64
	DailyPercentage   float64
	MonthlySpent      float64
	MonthlyLimit      float64
	MonthlyRemaining  float64
	MonthlyPercentage float64
	LastUpdated       time.Time
}

// String returns a human-readable budget status.
func (b BudgetStatus) String() string {
	return fmt.Sprintf(
		"Daily: $%.2f / $%.2f (%.1f%%) | Monthly: $%.2f / $%.2f (%.1f%%)",
		b.DailySpent, b.DailyLimit, b.DailyPercentage,
		b.MonthlySpent, b.MonthlyLimit, b.MonthlyPercentage,
	)
}
