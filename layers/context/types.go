package context

import (
	"time"

	"github.com/abrksh22/bplus/models"
)

// ContextTier represents the storage tier of context
type ContextTier int

const (
	// TierHot - in-memory, immediately available (0-50K tokens)
	TierHot ContextTier = iota
	// TierWarm - database, fast retrieval (50-200K tokens)
	TierWarm
	// TierCold - database, loaded only when needed (200K+ tokens)
	TierCold
)

// ContextItem represents a piece of context
type ContextItem struct {
	ID           string
	Type         ContextItemType
	Content      string
	Tier         ContextTier
	TokenCount   int
	LastAccessed time.Time
	Relevance    float64 // 0.0 to 1.0
	Metadata     map[string]interface{}
}

// ContextItemType categorizes context items
type ContextItemType string

const (
	TypeUserIntent     ContextItemType = "user_intent"
	TypePlan           ContextItemType = "plan"
	TypeFileContent    ContextItemType = "file_content"
	TypeValidation     ContextItemType = "validation"
	TypeMessage        ContextItemType = "message"
	TypeArchitecture   ContextItemType = "architecture"
	TypeToolResult     ContextItemType = "tool_result"
	TypeSummary        ContextItemType = "summary"
)

// ContextSnapshot represents a point-in-time snapshot of context
type ContextSnapshot struct {
	SessionID    string
	Timestamp    time.Time
	Items        []ContextItem
	TotalTokens  int
	Optimization float64 // Percentage of optimization applied
}

// ContextMetrics tracks context health and performance
type ContextMetrics struct {
	CurrentSize      int     // Current tokens
	MaxSize          int     // Maximum allowed tokens
	Efficiency       float64 // Optimization percentage
	Accuracy         float64 // Validated against actual state
	StalenessCount   int     // Number of outdated items
	HotTierSize      int     // Tokens in hot tier
	WarmTierSize     int     // Tokens in warm tier
	ColdTierSize     int     // Tokens in cold tier
	LastOptimized    time.Time
	OptimizationRuns int
}

// SummarizationRequest represents a request to summarize content
type SummarizationRequest struct {
	Content         string
	Type            ContextItemType
	TargetTokens    int // Desired output size
	PreserveDetails []string // Details to preserve
}

// SummarizationResult represents the result of summarization
type SummarizationResult struct {
	Original      string
	Summarized    string
	OriginalSize  int
	SummarySize   int
	Compression   float64 // Percentage reduction
	TokensSaved   int
	Details       map[string]string // Preserved details
}

// Checkpoint represents a saved state
type Checkpoint struct {
	ID          string
	SessionID   string
	Name        string
	Description string
	CreatedAt   time.Time
	Context     ContextSnapshot
	Messages    []models.Message
	FileStates  map[string]string // path -> content hash
	Metadata    map[string]interface{}
}

// OptimizationStrategy defines how to optimize context
type OptimizationStrategy string

const (
	StrategyAggressiveSummarization OptimizationStrategy = "aggressive_summarization"
	StrategySelectivePruning        OptimizationStrategy = "selective_pruning"
	StrategySemanticChunking        OptimizationStrategy = "semantic_chunking"
	StrategyTieredEviction          OptimizationStrategy = "tiered_eviction"
	StrategyBalanced                OptimizationStrategy = "balanced"
)

// OptimizationConfig configures context optimization
type OptimizationConfig struct {
	Strategy          OptimizationStrategy
	MaxTokens         int
	TargetTokens      int // Target after optimization
	MinRelevance      float64 // Items below this are candidates for removal
	PreserveTypes     []ContextItemType // Types to never remove
	SummarizationModel string // Model to use for summarization
	AutoOptimize      bool
	OptimizeThreshold float64 // Trigger optimization at this % of max
}
