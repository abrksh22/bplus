package models

import (
	"fmt"
	"strings"
)

// ParseModelName parses a model name in the format "provider/model-id".
// Returns the provider name and model ID.
func ParseModelName(fullName string) (provider, modelID string, err error) {
	parts := strings.SplitN(fullName, "/", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid model name format: %s (expected 'provider/model-id')", fullName)
	}

	provider = strings.TrimSpace(parts[0])
	modelID = strings.TrimSpace(parts[1])

	if provider == "" {
		return "", "", fmt.Errorf("provider name cannot be empty")
	}

	if modelID == "" {
		return "", "", fmt.Errorf("model ID cannot be empty")
	}

	return provider, modelID, nil
}

// FormatModelName formats a provider and model ID into the standard format.
func FormatModelName(provider, modelID string) string {
	return provider + "/" + modelID
}

// ValidateModelName validates a model name format.
func ValidateModelName(fullName string) error {
	_, _, err := ParseModelName(fullName)
	return err
}

// GetProviderFromModel extracts the provider name from a full model name.
func GetProviderFromModel(fullName string) (string, error) {
	provider, _, err := ParseModelName(fullName)
	return provider, err
}

// GetModelIDFromModel extracts the model ID from a full model name.
func GetModelIDFromModel(fullName string) (string, error) {
	_, modelID, err := ParseModelName(fullName)
	return modelID, err
}
