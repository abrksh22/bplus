// Package tools provides the core tool execution framework.
package tools

import (
	"context"
	"time"
)

// Tool defines the interface that all tools must implement.
// This interface is designed to support future plugin extensions.
type Tool interface {
	// Core metadata
	Name() string
	Description() string
	Parameters() []Parameter

	// Execution
	Execute(ctx context.Context, params map[string]interface{}) (*Result, error)

	// Permission and safety
	RequiresPermission() bool

	// Plugin support hooks (for future use)
	Category() string // "file", "exec", "git", "web", "mcp", "custom"
	Version() string  // Semantic versioning
	IsExternal() bool // true if loaded from plugin
}

// Parameter defines a tool parameter specification.
type Parameter struct {
	Name        string        // Parameter name
	Type        ParameterType // Data type
	Required    bool          // Whether this parameter is required
	Description string        // Human-readable description
	Default     interface{}   // Default value if not provided
	Validation  *Validation   // Optional validation rules
}

// ParameterType represents the data type of a parameter.
type ParameterType string

const (
	TypeString ParameterType = "string"
	TypeInt    ParameterType = "int"
	TypeFloat  ParameterType = "float"
	TypeBool   ParameterType = "bool"
	TypeArray  ParameterType = "array"
	TypeObject ParameterType = "object"
	TypeAny    ParameterType = "any"
)

// Validation defines validation rules for parameters.
type Validation struct {
	MinLength *int                    // Minimum string length or array size
	MaxLength *int                    // Maximum string length or array size
	Min       *float64                // Minimum numeric value
	Max       *float64                // Maximum numeric value
	Pattern   *string                 // Regex pattern for string validation
	Enum      []string                // Valid enum values
	Custom    func(interface{}) error // Custom validation function
}

// Result represents the result of a tool execution.
type Result struct {
	Success  bool                   // Whether execution succeeded
	Output   interface{}            // Tool output (can be any type)
	Error    error                  // Error if execution failed
	Metadata map[string]interface{} // Additional metadata
	Duration time.Duration          // Execution duration
}

// ExecutionContext provides context for tool execution.
type ExecutionContext struct {
	WorkingDir        string            // Working directory
	Environment       map[string]string // Environment variables
	Timeout           time.Duration     // Execution timeout
	PermissionGrants  []string          // Granted permissions
	AuditTrail        []AuditEntry      // Execution history
	CancellationToken context.Context   // For cancellation support
}

// AuditEntry records a single tool execution for audit purposes.
type AuditEntry struct {
	Timestamp  time.Time              // When the tool was executed
	ToolName   string                 // Name of the tool
	Parameters map[string]interface{} // Parameters provided
	Success    bool                   // Whether execution succeeded
	Duration   time.Duration          // How long it took
	Error      error                  // Error if any
}

// ToolError represents an error from tool execution.
type ToolError struct {
	Tool    string // Tool name
	Message string // Error message
	Cause   error  // Underlying error
	Code    string // Error code
}

func (e *ToolError) Error() string {
	if e.Cause != nil {
		return e.Tool + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.Tool + ": " + e.Message
}

func (e *ToolError) Unwrap() error {
	return e.Cause
}

// NewToolError creates a new ToolError.
func NewToolError(tool, message string, cause error) *ToolError {
	return &ToolError{
		Tool:    tool,
		Message: message,
		Cause:   cause,
	}
}

// ValidateParameters validates parameters against their specifications.
func ValidateParameters(params map[string]interface{}, specs []Parameter) error {
	// Check required parameters
	for _, spec := range specs {
		val, exists := params[spec.Name]

		if !exists {
			if spec.Required {
				return NewToolError("validation", "missing required parameter: "+spec.Name, nil)
			}
			// Use default if available
			if spec.Default != nil {
				params[spec.Name] = spec.Default
			}
			continue
		}

		// Type validation
		if err := validateType(val, spec.Type); err != nil {
			return NewToolError("validation", "parameter "+spec.Name+": "+err.Error(), err)
		}

		// Additional validation
		if spec.Validation != nil {
			if err := validateValue(val, spec.Validation); err != nil {
				return NewToolError("validation", "parameter "+spec.Name+": "+err.Error(), err)
			}
		}
	}

	return nil
}

func validateType(val interface{}, expectedType ParameterType) error {
	if expectedType == TypeAny {
		return nil
	}

	switch expectedType {
	case TypeString:
		if _, ok := val.(string); !ok {
			return NewToolError("type", "expected string", nil)
		}
	case TypeInt:
		switch val.(type) {
		case int, int32, int64, float64: // Allow numeric types
		default:
			return NewToolError("type", "expected int", nil)
		}
	case TypeFloat:
		if _, ok := val.(float64); !ok {
			return NewToolError("type", "expected float", nil)
		}
	case TypeBool:
		if _, ok := val.(bool); !ok {
			return NewToolError("type", "expected bool", nil)
		}
	case TypeArray:
		if _, ok := val.([]interface{}); !ok {
			return NewToolError("type", "expected array", nil)
		}
	case TypeObject:
		if _, ok := val.(map[string]interface{}); !ok {
			return NewToolError("type", "expected object", nil)
		}
	}

	return nil
}

func validateValue(val interface{}, validation *Validation) error {
	// String/Array length validation
	if validation.MinLength != nil || validation.MaxLength != nil {
		var length int
		switch v := val.(type) {
		case string:
			length = len(v)
		case []interface{}:
			length = len(v)
		}

		if validation.MinLength != nil && length < *validation.MinLength {
			return NewToolError("validation", "length below minimum", nil)
		}
		if validation.MaxLength != nil && length > *validation.MaxLength {
			return NewToolError("validation", "length exceeds maximum", nil)
		}
	}

	// Numeric range validation
	if validation.Min != nil || validation.Max != nil {
		var numVal float64
		switch v := val.(type) {
		case int:
			numVal = float64(v)
		case int32:
			numVal = float64(v)
		case int64:
			numVal = float64(v)
		case float64:
			numVal = v
		}

		if validation.Min != nil && numVal < *validation.Min {
			return NewToolError("validation", "value below minimum", nil)
		}
		if validation.Max != nil && numVal > *validation.Max {
			return NewToolError("validation", "value exceeds maximum", nil)
		}
	}

	// Pattern validation
	if validation.Pattern != nil {
		if str, ok := val.(string); ok {
			// Pattern matching would go here (using regexp)
			_ = str
		}
	}

	// Enum validation
	if len(validation.Enum) > 0 {
		if str, ok := val.(string); ok {
			valid := false
			for _, enum := range validation.Enum {
				if str == enum {
					valid = true
					break
				}
			}
			if !valid {
				return NewToolError("validation", "value not in allowed enum", nil)
			}
		}
	}

	// Custom validation
	if validation.Custom != nil {
		if err := validation.Custom(val); err != nil {
			return NewToolError("validation", "custom validation failed", err)
		}
	}

	return nil
}
