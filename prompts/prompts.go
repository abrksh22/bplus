package prompts

import (
	"fmt"
	"strings"
)

// GetLayer4Prompt returns the system prompt for Layer 4 (Main Agent).
func GetLayer4Prompt() string {
	return Layer4MainAgent
}

// GetLayer4PromptWithContext returns the Layer 4 prompt with additional context.
func GetLayer4PromptWithContext(context string) string {
	if context == "" {
		return Layer4MainAgent
	}

	return fmt.Sprintf("%s\n\n## Additional Context\n\n%s", Layer4MainAgent, context)
}

// GetLayer4PromptWithTools returns the Layer 4 prompt with tool descriptions.
func GetLayer4PromptWithTools(toolDescriptions []string) string {
	if len(toolDescriptions) == 0 {
		return Layer4MainAgent
	}

	toolList := strings.Join(toolDescriptions, "\n")
	return fmt.Sprintf("%s\n\n## Available Tools\n\n%s", Layer4MainAgent, toolList)
}

// CustomizePrompt allows customization of any prompt with additional instructions.
func CustomizePrompt(basePrompt string, customInstructions string) string {
	if customInstructions == "" {
		return basePrompt
	}

	return fmt.Sprintf("%s\n\n## Custom Instructions\n\n%s", basePrompt, customInstructions)
}
