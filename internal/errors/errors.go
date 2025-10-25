package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a categorized error type
type ErrorCode string

const (
	// Configuration errors
	ErrCodeConfig         ErrorCode = "CONFIG_ERROR"
	ErrCodeConfigNotFound ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeConfigInvalid  ErrorCode = "CONFIG_INVALID"

	// Database errors
	ErrCodeDatabase         ErrorCode = "DATABASE_ERROR"
	ErrCodeDatabaseNotFound ErrorCode = "DATABASE_NOT_FOUND"
	ErrCodeDatabaseConflict ErrorCode = "DATABASE_CONFLICT"

	// Provider errors
	ErrCodeProvider        ErrorCode = "PROVIDER_ERROR"
	ErrCodeProviderTimeout ErrorCode = "PROVIDER_TIMEOUT"
	ErrCodeProviderAuth    ErrorCode = "PROVIDER_AUTH_ERROR"
	ErrCodeProviderQuota   ErrorCode = "PROVIDER_QUOTA_EXCEEDED"

	// Tool errors
	ErrCodeTool           ErrorCode = "TOOL_ERROR"
	ErrCodeToolNotFound   ErrorCode = "TOOL_NOT_FOUND"
	ErrCodeToolPermission ErrorCode = "TOOL_PERMISSION_DENIED"
	ErrCodeToolExecution  ErrorCode = "TOOL_EXECUTION_ERROR"

	// File errors
	ErrCodeFile           ErrorCode = "FILE_ERROR"
	ErrCodeFileNotFound   ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFilePermission ErrorCode = "FILE_PERMISSION_DENIED"

	// Network errors
	ErrCodeNetwork        ErrorCode = "NETWORK_ERROR"
	ErrCodeNetworkTimeout ErrorCode = "NETWORK_TIMEOUT"

	// Validation errors
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"

	// User errors
	ErrCodeUser         ErrorCode = "USER_ERROR"
	ErrCodeUserCanceled ErrorCode = "USER_CANCELED"

	// Internal errors
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	ErrCodeUnknown  ErrorCode = "UNKNOWN_ERROR"
)

// Error represents a b+ error with additional context
type Error struct {
	Code        ErrorCode         `json:"code"`
	Message     string            `json:"message"`
	UserMsg     string            `json:"user_message,omitempty"` // User-friendly message
	Err         error             `json:"-"`                      // Wrapped error
	Context     map[string]string `json:"context,omitempty"`      // Additional context
	Retryable   bool              `json:"retryable"`
	Recoverable bool              `json:"recoverable"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *Error) WithContext(key, value string) *Error {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// WithUserMsg sets a user-friendly message
func (e *Error) WithUserMsg(msg string) *Error {
	e.UserMsg = msg
	return e
}

// New creates a new Error
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:        code,
		Message:     message,
		Retryable:   false,
		Recoverable: false,
	}
}

// Wrap wraps an existing error with a b+ error
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return nil
	}

	// If it's already a b+ Error, just add context
	var bpErr *Error
	if errors.As(err, &bpErr) {
		bpErr.Message = message + ": " + bpErr.Message
		return bpErr
	}

	return &Error{
		Code:        code,
		Message:     message,
		Err:         err,
		Retryable:   false,
		Recoverable: false,
	}
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *Error {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

// Is checks if an error matches a specific error code
func Is(err error, code ErrorCode) bool {
	var bpErr *Error
	if errors.As(err, &bpErr) {
		return bpErr.Code == code
	}
	return false
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var bpErr *Error
	if errors.As(err, &bpErr) {
		return bpErr.Retryable
	}
	return false
}

// IsRecoverable checks if an error is recoverable
func IsRecoverable(err error) bool {
	var bpErr *Error
	if errors.As(err, &bpErr) {
		return bpErr.Recoverable
	}
	return false
}

// GetUserMessage returns a user-friendly error message
func GetUserMessage(err error) string {
	var bpErr *Error
	if errors.As(err, &bpErr) {
		if bpErr.UserMsg != "" {
			return bpErr.UserMsg
		}
		return bpErr.Message
	}
	return err.Error()
}

// Common error constructors

// NewConfigError creates a configuration error
func NewConfigError(message string) *Error {
	return New(ErrCodeConfig, message).
		WithUserMsg("Configuration error: " + message)
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string) *Error {
	return New(ErrCodeDatabase, message)
}

// NewProviderError creates a provider error
func NewProviderError(message string) *Error {
	return &Error{
		Code:        ErrCodeProvider,
		Message:     message,
		Retryable:   true,
		Recoverable: true,
	}
}

// NewProviderTimeoutError creates a provider timeout error
func NewProviderTimeoutError(provider string) *Error {
	err := &Error{
		Code:        ErrCodeProviderTimeout,
		Message:     fmt.Sprintf("provider %s timed out", provider),
		UserMsg:     fmt.Sprintf("The AI provider (%s) is not responding. Please try again.", provider),
		Retryable:   true,
		Recoverable: true,
	}
	return err.WithContext("provider", provider)
}

// NewProviderAuthError creates a provider authentication error
func NewProviderAuthError(provider string) *Error {
	err := &Error{
		Code:    ErrCodeProviderAuth,
		Message: fmt.Sprintf("authentication failed for provider %s", provider),
		UserMsg: fmt.Sprintf("Authentication failed for %s. Please check your API key.", provider),
	}
	return err.WithContext("provider", provider)
}

// NewProviderQuotaError creates a provider quota error
func NewProviderQuotaError(provider string) *Error {
	err := &Error{
		Code:        ErrCodeProviderQuota,
		Message:     fmt.Sprintf("quota exceeded for provider %s", provider),
		UserMsg:     fmt.Sprintf("You've exceeded the quota for %s. Please check your usage limits.", provider),
		Retryable:   false,
		Recoverable: true,
	}
	return err.WithContext("provider", provider)
}

// NewToolError creates a tool error
func NewToolError(tool string, message string) *Error {
	err := New(ErrCodeTool, message)
	return err.WithContext("tool", tool)
}

// NewToolPermissionError creates a tool permission error
func NewToolPermissionError(tool string, action string) *Error {
	err := &Error{
		Code:    ErrCodeToolPermission,
		Message: fmt.Sprintf("permission denied for tool %s to perform %s", tool, action),
		UserMsg: fmt.Sprintf("Permission required: %s needs to %s. Please approve this action.", tool, action),
	}
	return err.WithContext("tool", tool).WithContext("action", action)
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(path string) *Error {
	err := &Error{
		Code:    ErrCodeFileNotFound,
		Message: fmt.Sprintf("file not found: %s", path),
		UserMsg: fmt.Sprintf("The file '%s' was not found.", path),
	}
	return err.WithContext("path", path)
}

// NewValidationError creates a validation error
func NewValidationError(field string, message string) *Error {
	err := &Error{
		Code:    ErrCodeValidation,
		Message: fmt.Sprintf("validation failed for %s: %s", field, message),
		UserMsg: fmt.Sprintf("Invalid %s: %s", field, message),
	}
	return err.WithContext("field", field)
}

// NewUserCanceledError creates a user canceled error
func NewUserCanceledError() *Error {
	return &Error{
		Code:        ErrCodeUserCanceled,
		Message:     "operation canceled by user",
		UserMsg:     "Operation canceled.",
		Recoverable: true,
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string) *Error {
	return &Error{
		Code:    ErrCodeInternal,
		Message: message,
		UserMsg: "An internal error occurred. Please try again or contact support.",
	}
}
