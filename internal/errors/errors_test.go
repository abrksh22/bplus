package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	err := New(ErrCodeConfig, "test error")
	assert.NotNil(t, err)
	assert.Equal(t, ErrCodeConfig, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.False(t, err.Retryable)
	assert.False(t, err.Recoverable)
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error without wrapped error",
			err: &Error{
				Code:    ErrCodeConfig,
				Message: "config error",
			},
			expected: "[CONFIG_ERROR] config error",
		},
		{
			name: "error with wrapped error",
			err: &Error{
				Code:    ErrCodeDatabase,
				Message: "database error",
				Err:     errors.New("connection failed"),
			},
			expected: "[DATABASE_ERROR] database error: connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestError_WithContext(t *testing.T) {
	err := New(ErrCodeTool, "tool error")
	err.WithContext("tool", "bash").WithContext("action", "execute")

	assert.Equal(t, "bash", err.Context["tool"])
	assert.Equal(t, "execute", err.Context["action"])
}

func TestError_WithUserMsg(t *testing.T) {
	err := New(ErrCodeProvider, "provider error")
	err.WithUserMsg("The AI provider is not responding")

	assert.Equal(t, "The AI provider is not responding", err.UserMsg)
}

func TestWrap(t *testing.T) {
	t.Run("wrap standard error", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrapped := Wrap(baseErr, ErrCodeDatabase, "database operation failed")

		assert.NotNil(t, wrapped)
		assert.Equal(t, ErrCodeDatabase, wrapped.Code)
		assert.Equal(t, "database operation failed", wrapped.Message)
		assert.Equal(t, baseErr, wrapped.Err)
	})

	t.Run("wrap nil error", func(t *testing.T) {
		wrapped := Wrap(nil, ErrCodeConfig, "test")
		assert.Nil(t, wrapped)
	})

	t.Run("wrap bp error", func(t *testing.T) {
		baseErr := New(ErrCodeProvider, "provider error")
		wrapped := Wrap(baseErr, ErrCodeDatabase, "database operation failed")

		assert.NotNil(t, wrapped)
		assert.Contains(t, wrapped.Message, "database operation failed")
		assert.Contains(t, wrapped.Message, "provider error")
	})
}

func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrapf(baseErr, ErrCodeFile, "failed to read file %s", "/path/to/file")

	assert.NotNil(t, wrapped)
	assert.Equal(t, "failed to read file /path/to/file", wrapped.Message)
}

func TestIs(t *testing.T) {
	t.Run("matching error code", func(t *testing.T) {
		err := New(ErrCodeConfig, "config error")
		assert.True(t, Is(err, ErrCodeConfig))
		assert.False(t, Is(err, ErrCodeDatabase))
	})

	t.Run("non-bp error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, Is(err, ErrCodeConfig))
	})

	t.Run("wrapped bp error", func(t *testing.T) {
		baseErr := New(ErrCodeProvider, "provider error")
		wrapped := Wrap(baseErr, ErrCodeDatabase, "wrapped")
		assert.True(t, Is(wrapped, ErrCodeProvider))
	})
}

func TestIsRetryable(t *testing.T) {
	t.Run("retryable error", func(t *testing.T) {
		err := &Error{
			Code:      ErrCodeProviderTimeout,
			Message:   "timeout",
			Retryable: true,
		}
		assert.True(t, IsRetryable(err))
	})

	t.Run("non-retryable error", func(t *testing.T) {
		err := New(ErrCodeConfig, "config error")
		assert.False(t, IsRetryable(err))
	})

	t.Run("standard error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, IsRetryable(err))
	})
}

func TestIsRecoverable(t *testing.T) {
	t.Run("recoverable error", func(t *testing.T) {
		err := &Error{
			Code:        ErrCodeUserCanceled,
			Message:     "canceled",
			Recoverable: true,
		}
		assert.True(t, IsRecoverable(err))
	})

	t.Run("non-recoverable error", func(t *testing.T) {
		err := New(ErrCodeConfig, "config error")
		assert.False(t, IsRecoverable(err))
	})
}

func TestGetUserMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name: "bp error with user message",
			err: &Error{
				Code:    ErrCodeProvider,
				Message: "provider failed",
				UserMsg: "The AI provider is not responding",
			},
			expected: "The AI provider is not responding",
		},
		{
			name: "bp error without user message",
			err: &Error{
				Code:    ErrCodeProvider,
				Message: "provider failed",
			},
			expected: "provider failed",
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: "standard error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetUserMessage(tt.err))
		})
	}
}

func TestNewConfigError(t *testing.T) {
	err := NewConfigError("invalid config")
	assert.Equal(t, ErrCodeConfig, err.Code)
	assert.Equal(t, "invalid config", err.Message)
	assert.Contains(t, err.UserMsg, "Configuration error")
}

func TestNewDatabaseError(t *testing.T) {
	err := NewDatabaseError("connection failed")
	assert.Equal(t, ErrCodeDatabase, err.Code)
	assert.Equal(t, "connection failed", err.Message)
}

func TestNewProviderError(t *testing.T) {
	err := NewProviderError("API error")
	assert.Equal(t, ErrCodeProvider, err.Code)
	assert.True(t, err.Retryable)
	assert.True(t, err.Recoverable)
}

func TestNewProviderTimeoutError(t *testing.T) {
	err := NewProviderTimeoutError("anthropic")
	assert.Equal(t, ErrCodeProviderTimeout, err.Code)
	assert.Contains(t, err.Message, "anthropic")
	assert.Contains(t, err.UserMsg, "not responding")
	assert.True(t, err.Retryable)
	assert.Equal(t, "anthropic", err.Context["provider"])
}

func TestNewProviderAuthError(t *testing.T) {
	err := NewProviderAuthError("openai")
	assert.Equal(t, ErrCodeProviderAuth, err.Code)
	assert.Contains(t, err.Message, "openai")
	assert.Contains(t, err.UserMsg, "API key")
	assert.Equal(t, "openai", err.Context["provider"])
}

func TestNewProviderQuotaError(t *testing.T) {
	err := NewProviderQuotaError("gemini")
	assert.Equal(t, ErrCodeProviderQuota, err.Code)
	assert.Contains(t, err.Message, "gemini")
	assert.Contains(t, err.UserMsg, "quota")
	assert.False(t, err.Retryable)
	assert.True(t, err.Recoverable)
}

func TestNewToolError(t *testing.T) {
	err := NewToolError("bash", "execution failed")
	assert.Equal(t, ErrCodeTool, err.Code)
	assert.Equal(t, "execution failed", err.Message)
	assert.Equal(t, "bash", err.Context["tool"])
}

func TestNewToolPermissionError(t *testing.T) {
	err := NewToolPermissionError("write", "save file")
	assert.Equal(t, ErrCodeToolPermission, err.Code)
	assert.Contains(t, err.Message, "write")
	assert.Contains(t, err.Message, "save file")
	assert.Contains(t, err.UserMsg, "Permission required")
}

func TestNewFileNotFoundError(t *testing.T) {
	err := NewFileNotFoundError("/path/to/file")
	assert.Equal(t, ErrCodeFileNotFound, err.Code)
	assert.Contains(t, err.Message, "/path/to/file")
	assert.Contains(t, err.UserMsg, "not found")
	assert.Equal(t, "/path/to/file", err.Context["path"])
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email", "invalid format")
	assert.Equal(t, ErrCodeValidation, err.Code)
	assert.Contains(t, err.Message, "email")
	assert.Contains(t, err.Message, "invalid format")
	assert.Equal(t, "email", err.Context["field"])
}

func TestNewUserCanceledError(t *testing.T) {
	err := NewUserCanceledError()
	assert.Equal(t, ErrCodeUserCanceled, err.Code)
	assert.True(t, err.Recoverable)
	assert.Contains(t, err.UserMsg, "canceled")
}

func TestNewInternalError(t *testing.T) {
	err := NewInternalError("unexpected panic")
	assert.Equal(t, ErrCodeInternal, err.Code)
	assert.Equal(t, "unexpected panic", err.Message)
	assert.Contains(t, err.UserMsg, "internal error")
}

func TestError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrap(baseErr, ErrCodeDatabase, "database error")

	unwrapped := wrapped.Unwrap()
	assert.Equal(t, baseErr, unwrapped)
}

func TestErrorChaining(t *testing.T) {
	// Test error unwrapping chain
	err1 := errors.New("root cause")
	err2 := Wrap(err1, ErrCodeDatabase, "database error")
	err3 := Wrap(err2, ErrCodeProvider, "provider error")

	// Use errors.Is to check the chain
	assert.True(t, errors.Is(err3, err1))
	assert.True(t, Is(err3, ErrCodeDatabase))
}

func TestErrorContext(t *testing.T) {
	err := New(ErrCodeTool, "tool error")
	err.WithContext("tool", "bash").
		WithContext("command", "ls").
		WithContext("exit_code", "1")

	assert.Len(t, err.Context, 3)
	assert.Equal(t, "bash", err.Context["tool"])
	assert.Equal(t, "ls", err.Context["command"])
	assert.Equal(t, "1", err.Context["exit_code"])
}

func TestMultipleErrorConstructors(t *testing.T) {
	errors := []struct {
		name string
		err  *Error
		code ErrorCode
	}{
		{"config", NewConfigError("test"), ErrCodeConfig},
		{"database", NewDatabaseError("test"), ErrCodeDatabase},
		{"provider", NewProviderError("test"), ErrCodeProvider},
		{"tool", NewToolError("bash", "test"), ErrCodeTool},
		{"validation", NewValidationError("field", "test"), ErrCodeValidation},
		{"internal", NewInternalError("test"), ErrCodeInternal},
	}

	for _, tc := range errors {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.err)
			assert.Equal(t, tc.code, tc.err.Code)
		})
	}
}
