package execution

import (
	"fmt"
	"sync"
	"time"

	"github.com/abrksh22/bplus/models"
)

// CostTracker tracks token usage and costs for an agent session.
type CostTracker struct {
	mu              sync.RWMutex
	totalInput      int
	totalOutput     int
	totalCost       float64
	entries         []CostEntry
	sessionStart    time.Time
	lastReset       time.Time
	dailyBudget     float64
	dailySpent      float64
	budgetWarning   float64
	warningCallback func(float64, float64) // (spent, budget)
}

// CostEntry represents a single cost record.
type CostEntry struct {
	Timestamp    time.Time
	InputTokens  int
	OutputTokens int
	Cost         float64
	ModelName    string
	Operation    string // e.g., "completion", "streaming"
}

// NewCostTracker creates a new cost tracker.
func NewCostTracker() *CostTracker {
	now := time.Now()
	return &CostTracker{
		entries:      make([]CostEntry, 0, 100),
		sessionStart: now,
		lastReset:    now,
	}
}

// AddUsage records token usage and cost.
func (ct *CostTracker) AddUsage(usage models.Usage) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.totalInput += usage.InputTokens
	ct.totalOutput += usage.OutputTokens
	ct.totalCost += usage.Cost
	ct.dailySpent += usage.Cost

	entry := CostEntry{
		Timestamp:    time.Now(),
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
		Cost:         usage.Cost,
		Operation:    "completion",
	}
	ct.entries = append(ct.entries, entry)

	// Check budget warning
	if ct.dailyBudget > 0 && ct.budgetWarning > 0 {
		if ct.dailySpent >= ct.budgetWarning && ct.warningCallback != nil {
			go ct.warningCallback(ct.dailySpent, ct.dailyBudget)
		}
	}
}

// GetTotals returns total token usage and cost.
func (ct *CostTracker) GetTotals() (inputTokens, outputTokens int, cost float64) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.totalInput, ct.totalOutput, ct.totalCost
}

// GetDailySpent returns the amount spent today.
func (ct *CostTracker) GetDailySpent() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.dailySpent
}

// GetEntries returns all cost entries.
func (ct *CostTracker) GetEntries() []CostEntry {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// Return a copy to prevent external modification
	entries := make([]CostEntry, len(ct.entries))
	copy(entries, ct.entries)
	return entries
}

// SetDailyBudget sets the daily spending budget.
func (ct *CostTracker) SetDailyBudget(budget float64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.dailyBudget = budget
	ct.budgetWarning = budget * 0.8 // 80% threshold
}

// SetWarningCallback sets a callback for budget warnings.
func (ct *CostTracker) SetWarningCallback(callback func(spent, budget float64)) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.warningCallback = callback
}

// CheckBudget returns true if under budget, false if over.
func (ct *CostTracker) CheckBudget() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if ct.dailyBudget <= 0 {
		return true // No budget set
	}
	return ct.dailySpent < ct.dailyBudget
}

// RemainingBudget returns the remaining daily budget.
func (ct *CostTracker) RemainingBudget() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	if ct.dailyBudget <= 0 {
		return -1 // No budget set
	}
	return ct.dailyBudget - ct.dailySpent
}

// ResetDaily resets daily spending counters.
func (ct *CostTracker) ResetDaily() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.dailySpent = 0
	ct.lastReset = time.Now()
}

// Reset clears all tracking data.
func (ct *CostTracker) Reset() {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.totalInput = 0
	ct.totalOutput = 0
	ct.totalCost = 0
	ct.dailySpent = 0
	ct.entries = make([]CostEntry, 0, 100)
	ct.sessionStart = time.Now()
	ct.lastReset = time.Now()
}

// GetSessionDuration returns the duration since session start.
func (ct *CostTracker) GetSessionDuration() time.Duration {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return time.Since(ct.sessionStart)
}

// EstimateCost estimates the cost for a given number of tokens.
func EstimateCost(inputTokens, outputTokens int, pricing models.Pricing) float64 {
	inputCost := (float64(inputTokens) / 1000.0) * pricing.InputTokens
	outputCost := (float64(outputTokens) / 1000.0) * pricing.OutputTokens
	total := inputCost + outputCost

	if total < pricing.MinimumCost {
		return pricing.MinimumCost
	}
	return total
}

// FormatCost formats a cost value as a string.
func FormatCost(cost float64) string {
	if cost < 0.01 {
		return "<$0.01"
	}
	return fmt.Sprintf("$%.2f", cost)
}
