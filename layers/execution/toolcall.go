package execution

import (
	"encoding/json"
	"fmt"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/models"
	"github.com/abrksh22/bplus/tools"
)

// ValidateToolCall validates a tool call before execution.
func ValidateToolCall(toolCall models.ToolCall, tool tools.Tool) error {
	if toolCall.Name == "" {
		return errors.New(errors.ErrCodeValidation, "tool call must have a name")
	}

	if tool == nil {
		return errors.Newf(errors.ErrCodeToolNotFound, "tool %s not found", toolCall.Name)
	}

	// Validate parameters
	params := tool.Parameters()
	provided := toolCall.Arguments

	// Check for required parameters
	for _, param := range params {
		if param.Required {
			if _, ok := provided[param.Name]; !ok {
				return errors.Newf(
					errors.ErrCodeValidation,
					"missing required parameter %s for tool %s",
					param.Name,
					toolCall.Name,
				)
			}
		}
	}

	// Validate parameter types
	for key, value := range provided {
		// Find parameter definition
		var paramDef *tools.Parameter
		for i := range params {
			if params[i].Name == key {
				paramDef = &params[i]
				break
			}
		}

		if paramDef == nil {
			// Unknown parameter - warn but don't fail
			continue
		}

		// Type checking
		if err := validateParameterType(value, paramDef.Type); err != nil {
			return errors.Wrapf(
				err,
				errors.ErrCodeValidation,
				"invalid type for parameter %s in tool %s",
				key,
				toolCall.Name,
			)
		}
	}

	return nil
}

// validateParameterType checks if a value matches the expected parameter type.
func validateParameterType(value interface{}, expectedType tools.ParameterType) error {
	if value == nil {
		return nil // Null is generally acceptable
	}

	switch expectedType {
	case tools.TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case tools.TypeInt:
		switch v := value.(type) {
		case int, int8, int16, int32, int64:
			// OK
		case float64:
			// JSON numbers are float64, accept if it's an integer value
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer, got float %f", v)
			}
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case tools.TypeFloat:
		switch value.(type) {
		case float32, float64, int, int8, int16, int32, int64:
			// OK - integers can be converted to floats
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case tools.TypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case tools.TypeArray:
		switch value.(type) {
		case []interface{}, []string, []int, []float64:
			// OK
		default:
			return fmt.Errorf("expected array, got %T", value)
		}
	case tools.TypeObject:
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	case tools.TypeAny:
		// Any type is acceptable
	default:
		return fmt.Errorf("unknown parameter type: %s", expectedType)
	}

	return nil
}

// FormatToolResult formats a tool result for return to the LLM.
func FormatToolResult(result *tools.Result, toolName string) string {
	if result == nil {
		return fmt.Sprintf("Tool %s returned no result", toolName)
	}

	if !result.Success {
		errMsg := "unknown error"
		if result.Error != nil {
			errMsg = result.Error.Error()
		}
		return fmt.Sprintf("Tool %s failed: %s", toolName, errMsg)
	}

	// Format output based on type
	switch output := result.Output.(type) {
	case string:
		return output
	case []byte:
		return string(output)
	case nil:
		return fmt.Sprintf("Tool %s completed successfully", toolName)
	default:
		// Try to format as JSON for structured data
		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", output)
		}
		return string(jsonBytes)
	}
}

// FormatToolResultWithMetadata includes metadata in the formatted result.
func FormatToolResultWithMetadata(result *tools.Result, toolName string) string {
	baseResult := FormatToolResult(result, toolName)

	if result == nil || len(result.Metadata) == 0 {
		return baseResult
	}

	// Add metadata if present
	metadataJSON, err := json.MarshalIndent(result.Metadata, "", "  ")
	if err != nil {
		return baseResult
	}

	return fmt.Sprintf("%s\n\nMetadata:\n%s", baseResult, string(metadataJSON))
}

// ParseToolArguments parses tool arguments from various formats.
func ParseToolArguments(raw interface{}) (map[string]interface{}, error) {
	switch args := raw.(type) {
	case map[string]interface{}:
		return args, nil
	case string:
		// Try to parse as JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(args), &parsed); err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeValidation, "failed to parse tool arguments from JSON string")
		}
		return parsed, nil
	case []byte:
		// Try to parse as JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(args, &parsed); err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeValidation, "failed to parse tool arguments from JSON bytes")
		}
		return parsed, nil
	case nil:
		return make(map[string]interface{}), nil
	default:
		return nil, errors.Newf(errors.ErrCodeValidation, "unsupported tool arguments type: %T", raw)
	}
}

// RecoverFromMalformedToolCall attempts to recover from a malformed tool call.
func RecoverFromMalformedToolCall(toolCall models.ToolCall) (*models.ToolCall, error) {
	// Create a copy to modify
	recovered := toolCall

	// Ensure name is present
	if recovered.Name == "" {
		return nil, errors.New(errors.ErrCodeValidation, "tool call has no name")
	}

	// Ensure arguments is not nil
	if recovered.Arguments == nil {
		recovered.Arguments = make(map[string]interface{})
	}

	// Try to fix common issues with arguments
	if len(recovered.Arguments) == 0 {
		// Check if there's a single argument that might be the whole object
		for key, value := range recovered.Arguments {
			if mapValue, ok := value.(map[string]interface{}); ok {
				// Replace arguments with the nested map
				recovered.Arguments = mapValue
				break
			} else if strValue, ok := value.(string); ok && key == "arguments" {
				// Try to parse as JSON
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(strValue), &parsed); err == nil {
					recovered.Arguments = parsed
					break
				}
			}
		}
	}

	return &recovered, nil
}

// ToolCallError represents an error during tool calling.
type ToolCallError struct {
	ToolName string
	CallID   string
	Phase    string // "validation", "execution", "formatting"
	Err      error
}

func (e *ToolCallError) Error() string {
	return fmt.Sprintf("tool call error [%s/%s]: %v", e.ToolName, e.Phase, e.Err)
}

func (e *ToolCallError) Unwrap() error {
	return e.Err
}
