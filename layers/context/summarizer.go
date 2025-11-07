package context

import (
	"context"
	"fmt"
	"strings"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
	"github.com/abrksh22/bplus/models"
)

// Summarizer uses LLM to summarize context
type Summarizer struct {
	provider models.Provider
	model    string
	logger   *logging.Logger
}

// NewSummarizer creates a new summarizer
func NewSummarizer(provider models.Provider, model string) *Summarizer {
	if model == "" {
		model = "default" // Use provider's default model
	}

	return &Summarizer{
		provider: provider,
		model:    model,
		logger:   logging.NewDefaultLogger().WithComponent("summarizer"),
	}
}

// Summarize summarizes content using LLM
func (s *Summarizer) Summarize(ctx context.Context, req SummarizationRequest) (*SummarizationResult, error) {
	if req.Content == "" {
		return nil, errors.New(errors.ErrCodeInvalidInput, "content is required")
	}

	// Estimate token counts (rough approximation: 1 token ≈ 4 characters)
	originalTokens := estimateTokens(req.Content)

	if req.TargetTokens == 0 {
		req.TargetTokens = originalTokens / 2 // Default to 50% reduction
	}

	// If already small enough, no need to summarize
	if originalTokens <= req.TargetTokens {
		return &SummarizationResult{
			Original:     req.Content,
			Summarized:   req.Content,
			OriginalSize: originalTokens,
			SummarySize:  originalTokens,
			Compression:  0,
			TokensSaved:  0,
		}, nil
	}

	// Build summarization prompt
	prompt := s.buildPrompt(req)

	// Call LLM
	completion, err := s.provider.CreateCompletion(ctx, models.CompletionRequest{
		Messages: []models.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   req.TargetTokens,
		Temperature: 0.3, // Lower temperature for factual summarization
	})

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeExternal, "summarization LLM call failed")
	}

	if len(completion.Choices) == 0 {
		return nil, errors.New(errors.ErrCodeExternal, "no summary generated")
	}

	summary := completion.Choices[0].Message.Content
	summaryTokens := estimateTokens(summary)

	result := &SummarizationResult{
		Original:     req.Content,
		Summarized:   summary,
		OriginalSize: originalTokens,
		SummarySize:  summaryTokens,
		TokensSaved:  originalTokens - summaryTokens,
		Details:      make(map[string]string),
	}

	if originalTokens > 0 {
		result.Compression = float64(result.TokensSaved) / float64(originalTokens)
	}

	s.logger.Debug("Summarization complete",
		"original_tokens", originalTokens,
		"summary_tokens", summaryTokens,
		"compression", fmt.Sprintf("%.1f%%", result.Compression*100))

	return result, nil
}

// buildPrompt builds a summarization prompt based on content type
func (s *Summarizer) buildPrompt(req SummarizationRequest) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("Summarize the following ")

	switch req.Type {
	case TypeToolResult:
		promptBuilder.WriteString("tool execution output")
	case TypeMessage:
		promptBuilder.WriteString("conversation messages")
	case TypeFileContent:
		promptBuilder.WriteString("file content")
	case TypeValidation:
		promptBuilder.WriteString("validation results")
	case TypePlan:
		promptBuilder.WriteString("execution plan")
	default:
		promptBuilder.WriteString("content")
	}

	promptBuilder.WriteString(" while preserving the most important information.\n\n")

	// Add specific preservation instructions
	if len(req.PreserveDetails) > 0 {
		promptBuilder.WriteString("CRITICAL: You must preserve these details:\n")
		for _, detail := range req.PreserveDetails {
			promptBuilder.WriteString(fmt.Sprintf("- %s\n", detail))
		}
		promptBuilder.WriteString("\n")
	}

	// Add content-specific instructions
	switch req.Type {
	case TypeToolResult:
		promptBuilder.WriteString("Focus on:\n")
		promptBuilder.WriteString("- Final results and outputs\n")
		promptBuilder.WriteString("- Any errors or warnings\n")
		promptBuilder.WriteString("- Key changes made\n")
		promptBuilder.WriteString("Skip verbose logs and intermediate steps.\n\n")

	case TypeMessage:
		promptBuilder.WriteString("Focus on:\n")
		promptBuilder.WriteString("- User's request and intent\n")
		promptBuilder.WriteString("- Key decisions made\n")
		promptBuilder.WriteString("- Important outcomes\n")
		promptBuilder.WriteString("Skip pleasantries and redundant information.\n\n")

	case TypeFileContent:
		promptBuilder.WriteString("Focus on:\n")
		promptBuilder.WriteString("- Core functionality and purpose\n")
		promptBuilder.WriteString("- Public APIs and interfaces\n")
		promptBuilder.WriteString("- Critical implementation details\n")
		promptBuilder.WriteString("Skip boilerplate and comments.\n\n")

	case TypeValidation:
		promptBuilder.WriteString("Focus on:\n")
		promptBuilder.WriteString("- Issues found\n")
		promptBuilder.WriteString("- Resolution status\n")
		promptBuilder.WriteString("- Remaining concerns\n")
		promptBuilder.WriteString("Skip successful validation checks.\n\n")

	case TypePlan:
		promptBuilder.WriteString("Focus on:\n")
		promptBuilder.WriteString("- High-level approach\n")
		promptBuilder.WriteString("- Key steps\n")
		promptBuilder.WriteString("- Files to modify\n")
		promptBuilder.WriteString("Skip detailed explanations.\n\n")
	}

	promptBuilder.WriteString(fmt.Sprintf("Target length: approximately %d tokens (about %d characters).\n\n",
		req.TargetTokens, req.TargetTokens*4))

	promptBuilder.WriteString("Content to summarize:\n\n")
	promptBuilder.WriteString(req.Content)

	return promptBuilder.String()
}

// SummarizeBatch summarizes multiple items efficiently
func (s *Summarizer) SummarizeBatch(ctx context.Context, requests []SummarizationRequest) ([]SummarizationResult, error) {
	results := make([]SummarizationResult, 0, len(requests))

	// TODO: Implement batch processing with concurrent requests
	// For now, process sequentially
	for _, req := range requests {
		result, err := s.Summarize(ctx, req)
		if err != nil {
			s.logger.Warn("Batch summarization item failed", "error", err)
			// Add a result with original content
			results = append(results, SummarizationResult{
				Original:     req.Content,
				Summarized:   req.Content,
				OriginalSize: estimateTokens(req.Content),
				SummarySize:  estimateTokens(req.Content),
				Compression:  0,
				TokensSaved:  0,
			})
			continue
		}
		results = append(results, *result)
	}

	return results, nil
}

// Token estimation constants
const (
	CharsPerToken = 4 // Rough approximation for English text
)

// estimateTokens estimates token count from text
// Rough approximation: 1 token ≈ 4 characters for English text
func estimateTokens(text string) int {
	// Count characters
	chars := len(text)

	// Estimate tokens (4 chars per token is a common approximation)
	tokens := chars / CharsPerToken

	// Minimum of 1 token
	if tokens == 0 && chars > 0 {
		tokens = 1
	}

	return tokens
}
