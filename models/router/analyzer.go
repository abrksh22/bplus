package router

import (
	"strings"

	"github.com/abrksh22/bplus/models"
)

// TaskAnalysis contains the results of analyzing a task/request.
// This is a scaffold for Phase 4.4 - full implementation will come in later phases.
type TaskAnalysis struct {
	Type           string   // "coding", "chat", "analysis", "creative"
	Complexity     int      // 1-10 scale
	EstTokens      int      // Estimated total tokens
	RequiresTools  bool     // Whether the task requires tool usage
	RequiresVision bool     // Whether the task requires vision capabilities
	Language       string   // Programming language (if coding task)
	Keywords       []string // Extracted keywords
}

// AnalyzeTask analyzes a completion request to determine its characteristics.
// This is a placeholder implementation - full logic will be added in later phases.
func AnalyzeTask(req *models.CompletionRequest) TaskAnalysis {
	analysis := TaskAnalysis{
		Type:           "chat", // Default
		Complexity:     5,      // Medium complexity default
		EstTokens:      estimateTokens(req),
		RequiresTools:  false,
		RequiresVision: false,
		Language:       "",
		Keywords:       []string{},
	}

	// Simple heuristic-based analysis (placeholder)
	// Full implementation will use more sophisticated analysis

	// Check for coding keywords
	codingKeywords := []string{"function", "class", "import", "def", "const", "let", "var", "struct", "interface"}
	messageText := extractMessageText(req)
	lowerText := strings.ToLower(messageText)

	for _, keyword := range codingKeywords {
		if strings.Contains(lowerText, keyword) {
			analysis.Type = "coding"
			analysis.Keywords = append(analysis.Keywords, keyword)
			break
		}
	}

	// Check for analysis keywords
	analysisKeywords := []string{"analyze", "explain", "compare", "evaluate", "assess"}
	for _, keyword := range analysisKeywords {
		if strings.Contains(lowerText, keyword) {
			analysis.Type = "analysis"
			analysis.Complexity = 7 // Analysis tasks are typically more complex
			break
		}
	}

	// Detect language (very basic - will improve in full implementation)
	if analysis.Type == "coding" {
		analysis.Language = detectLanguage(messageText)
	}

	// Estimate complexity based on message length and structure
	wordCount := len(strings.Fields(messageText))
	if wordCount > 200 {
		analysis.Complexity = 8
	} else if wordCount > 100 {
		analysis.Complexity = 6
	} else if wordCount > 50 {
		analysis.Complexity = 5
	} else {
		analysis.Complexity = 3
	}

	// Check if tools might be needed
	toolKeywords := []string{"file", "read", "write", "execute", "run", "test", "build"}
	for _, keyword := range toolKeywords {
		if strings.Contains(lowerText, keyword) {
			analysis.RequiresTools = true
			break
		}
	}

	return analysis
}

// String returns a human-readable representation of the task analysis.
func (t TaskAnalysis) String() string {
	var parts []string
	parts = append(parts, "Type: "+t.Type)
	parts = append(parts, "Complexity: "+complexityToString(t.Complexity))
	parts = append(parts, "Est. Tokens: "+string(rune(t.EstTokens)))

	if t.RequiresTools {
		parts = append(parts, "Requires: Tools")
	}
	if t.RequiresVision {
		parts = append(parts, "Requires: Vision")
	}
	if t.Language != "" {
		parts = append(parts, "Language: "+t.Language)
	}

	return strings.Join(parts, ", ")
}

// Helper functions (placeholders for future implementation)

func estimateTokens(req *models.CompletionRequest) int {
	// Placeholder - will use proper tokenization in full implementation
	// Rough estimate: 1 token ≈ 4 characters
	messageText := extractMessageText(req)
	return len(messageText) / 4
}

func extractMessageText(req *models.CompletionRequest) string {
	// Placeholder - extract text from messages
	var text string
	for _, msg := range req.Messages {
		text += msg.Content + " "
	}
	return text
}

func detectLanguage(text string) string {
	// Very basic language detection (placeholder)
	lowerText := strings.ToLower(text)

	languages := map[string][]string{
		"go":         {"package", "func", "goroutine", "chan"},
		"python":     {"def", "import", "class", "__init__"},
		"javascript": {"const", "let", "var", "async", "=>"},
		"typescript": {"interface", "type", "const", "async"},
		"rust":       {"fn", "impl", "trait", "pub"},
		"java":       {"public", "class", "static", "void"},
	}

	for lang, keywords := range languages {
		for _, keyword := range keywords {
			if strings.Contains(lowerText, keyword) {
				return lang
			}
		}
	}

	return ""
}

func complexityToString(complexity int) string {
	switch {
	case complexity <= 3:
		return "Low"
	case complexity <= 6:
		return "Medium"
	case complexity <= 8:
		return "High"
	default:
		return "Very High"
	}
}

// ShouldUseExpensiveModel returns whether an expensive model should be used.
// Placeholder for future implementation.
func (t TaskAnalysis) ShouldUseExpensiveModel() bool {
	// Simple heuristic - will be more sophisticated in full implementation
	return t.Complexity >= 7 || t.Type == "analysis"
}

// ShouldUseLocalModel returns whether a local model is sufficient.
// Placeholder for future implementation.
func (t TaskAnalysis) ShouldUseLocalModel() bool {
	// Simple heuristic - will be more sophisticated in full implementation
	return t.Complexity <= 4 && t.Type == "chat" && !t.RequiresTools
}

// EstimateMaxTokens estimates the maximum tokens that might be needed.
// Placeholder for future implementation.
func (t TaskAnalysis) EstimateMaxTokens() int {
	// Will use more sophisticated estimation in full implementation
	baseTokens := t.EstTokens

	// Add buffer based on complexity
	buffer := float64(baseTokens) * (float64(t.Complexity) / 10.0)

	return baseTokens + int(buffer)
}
