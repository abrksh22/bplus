package logging

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/abrksh22/bplus/internal/config"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  config.LoggingConfig
		wantErr bool
	}{
		{
			name: "valid text logger",
			config: config.LoggingConfig{
				Level:  "info",
				Format: "text",
			},
			wantErr: false,
		},
		{
			name: "valid json logger",
			config: config.LoggingConfig{
				Level:  "debug",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "all log levels",
			config: config.LoggingConfig{
				Level:  "warn",
				Format: "text",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestNewLogger_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		File:       logFile,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
	}

	logger, err := NewLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Write a log message
	logger.Info("test message", "key", "value")

	// Verify log file was created
	_, err = os.Stat(logFile)
	assert.NoError(t, err)
}

func TestLogger_Debug(t *testing.T) {
	logger := NewDefaultLogger()
	assert.NotNil(t, logger)

	// Should not panic
	logger.Debug("debug message")
	logger.Debug("debug with fields", "key1", "value1", "key2", 123)
}

func TestLogger_Info(t *testing.T) {
	logger := NewDefaultLogger()
	assert.NotNil(t, logger)

	logger.Info("info message")
	logger.Info("info with fields", "user", "test", "count", 5)
}

func TestLogger_Warn(t *testing.T) {
	logger := NewDefaultLogger()
	assert.NotNil(t, logger)

	logger.Warn("warning message")
	logger.Warn("warning with fields", "resource", "database", "status", "slow")
}

func TestLogger_Error(t *testing.T) {
	logger := NewDefaultLogger()
	assert.NotNil(t, logger)

	testErr := errors.New("test error")

	logger.Error("error message", testErr)
	logger.Error("error with fields", testErr, "operation", "save", "id", 123)
}

func TestLogger_With(t *testing.T) {
	logger := NewDefaultLogger()

	// Create logger with fields
	loggerWithFields := logger.With("service", "test", "version", "1.0")
	assert.NotNil(t, loggerWithFields)

	// Should not panic
	loggerWithFields.Info("test message")
}

func TestLogger_WithComponent(t *testing.T) {
	logger := NewDefaultLogger()

	componentLogger := logger.WithComponent("auth")
	assert.NotNil(t, componentLogger)

	// Should not panic
	componentLogger.Info("authentication successful")
}

func TestLogger_WithContext(t *testing.T) {
	logger := NewDefaultLogger()
	ctx := context.Background()

	contextLogger := logger.WithContext(ctx)
	assert.NotNil(t, contextLogger)

	// Should not panic
	contextLogger.Info("context message")
}

func TestLogger_ToContext_FromContext(t *testing.T) {
	logger := NewDefaultLogger()
	ctx := context.Background()

	// Add logger to context
	ctxWithLogger := logger.ToContext(ctx)
	assert.NotNil(t, ctxWithLogger)

	// Extract logger from context
	extractedLogger := FromContext(ctxWithLogger)
	assert.NotNil(t, extractedLogger)

	// Verify it's the same logger
	// (We can't directly compare, but we can verify it works)
	extractedLogger.Info("message from context logger")
}

func TestFromContext_WithoutLogger(t *testing.T) {
	ctx := context.Background()

	// Should return default logger if no logger in context
	logger := FromContext(ctx)
	assert.NotNil(t, logger)

	// Should not panic
	logger.Info("message from default logger")
}

func TestMiddleware_Trace(t *testing.T) {
	logger := NewDefaultLogger()
	middleware := NewMiddleware(logger)

	// Should not panic
	defer middleware.Trace("test_function", "param", "value")()

	// Simulate function work
	// No assertions needed, just verify it doesn't panic
}

func TestMiddleware_TraceWithError(t *testing.T) {
	logger := NewDefaultLogger()
	middleware := NewMiddleware(logger)

	t.Run("with error", func(t *testing.T) {
		defer middleware.TraceWithError("test_function_error", "param", "value")(errors.New("test error"))
		// Function that returns error
	})

	t.Run("without error", func(t *testing.T) {
		defer middleware.TraceWithError("test_function_success", "param", "value")(nil)
		// Function that succeeds
	})
}

func TestFieldsToMap(t *testing.T) {
	tests := []struct {
		name     string
		fields   []interface{}
		expected map[string]interface{}
	}{
		{
			name:     "empty fields",
			fields:   []interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:   "single pair",
			fields: []interface{}{"key", "value"},
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name:   "multiple pairs",
			fields: []interface{}{"key1", "value1", "key2", 123, "key3", true},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
		},
		{
			name:   "odd number of fields",
			fields: []interface{}{"key1", "value1", "key2"},
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name:     "non-string key",
			fields:   []interface{}{123, "value"},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fieldsToMap(tt.fields...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test global convenience functions
	// These should not panic
	Debug("debug message", "key", "value")
	Info("info message", "key", "value")
	Warn("warn message", "key", "value")
	Error("error message", errors.New("test"), "key", "value")
}

func TestSetGlobalLogger(t *testing.T) {
	// Create a custom logger
	customLogger := NewDefaultLogger()

	// Set it as global
	SetGlobalLogger(customLogger)

	// Use global functions (should not panic)
	Info("test with custom global logger")
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "debug level",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "info level",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "warn level",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "error level",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseLogLevel(tt.level)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkLogger_Info(b *testing.B) {
	logger := NewDefaultLogger()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "key", "value", "number", i)
	}
}

func BenchmarkLogger_WithFields(b *testing.B) {
	logger := NewDefaultLogger()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.With("service", "test", "version", "1.0").Info("benchmark message")
	}
}

// Helper to capture log output for testing
type testWriter struct {
	*bytes.Buffer
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	return tw.Buffer.Write(p)
}
