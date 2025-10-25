package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/abrksh22/bplus/internal/config"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	loggerKey contextKey = "logger"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	logger zerolog.Logger
}

// NewLogger creates a new logger based on configuration
func NewLogger(cfg config.LoggingConfig) (*Logger, error) {
	// Parse log level
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// Set global log level
	zerolog.SetGlobalLevel(level)

	// Create output writers
	var writers []io.Writer

	// Console writer (stderr)
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	if cfg.Format == "json" {
		// Use JSON format instead of console format
		writers = append(writers, os.Stderr)
	} else {
		// Use pretty console format
		writers = append(writers, consoleWriter)
	}

	// Add file writer if configured
	if cfg.File != "" {
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize, // megabytes
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge, // days
			Compress:   true,
		}
		writers = append(writers, fileWriter)
	}

	// Create multi-writer
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = io.MultiWriter(writers...)
	}

	// Create zerolog logger
	zlog := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{
		logger: zlog,
	}, nil
}

// NewDefaultLogger creates a logger with default settings
func NewDefaultLogger() *Logger {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "text",
	}

	logger, _ := NewLogger(cfg)
	return logger
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) (zerolog.Level, error) {
	switch level {
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.logger.Info().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields ...interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	event.Fields(fieldsToMap(fields...)).Msg(msg)
}

// With creates a new logger with additional fields
func (l *Logger) With(fields ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Fields(fieldsToMap(fields...)).Logger(),
	}
}

// WithComponent creates a new logger with a component field
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("component", component).Logger(),
	}
}

// WithContext creates a new logger with context fields
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract any context values and add to logger
	// This can be extended to extract specific context values
	return &Logger{
		logger: l.logger.With().Logger(),
	}
}

// fieldsToMap converts variadic fields to a map
// Expected format: key1, value1, key2, value2, ...
func fieldsToMap(fields ...interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			m[key] = fields[i+1]
		}
	}
	return m
}

// ToContext adds the logger to a context
func (l *Logger) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext extracts a logger from a context
// Returns default logger if not found
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	return NewDefaultLogger()
}

// Middleware provides logging middleware for function tracing
type Middleware struct {
	logger *Logger
}

// NewMiddleware creates a new logging middleware
func NewMiddleware(logger *Logger) *Middleware {
	return &Middleware{
		logger: logger,
	}
}

// Trace logs function entry and exit
func (m *Middleware) Trace(functionName string, fields ...interface{}) func() {
	start := time.Now()
	m.logger.Debug("function.enter",
		append(fields, "function", functionName)...)

	return func() {
		duration := time.Since(start)
		m.logger.Debug("function.exit",
			append(fields,
				"function", functionName,
				"duration_ms", duration.Milliseconds())...)
	}
}

// TraceWithError logs function entry and exit with error checking
func (m *Middleware) TraceWithError(functionName string, fields ...interface{}) func(error) {
	start := time.Now()
	m.logger.Debug("function.enter",
		append(fields, "function", functionName)...)

	return func(err error) {
		duration := time.Since(start)
		if err != nil {
			m.logger.Error("function.error",
				err,
				append(fields,
					"function", functionName,
					"duration_ms", duration.Milliseconds())...)
		} else {
			m.logger.Debug("function.exit",
				append(fields,
					"function", functionName,
					"duration_ms", duration.Milliseconds())...)
		}
	}
}

// Global logger instance (can be replaced with application logger)
var globalLogger = NewDefaultLogger()

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// Global convenience functions

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...interface{}) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...interface{}) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...interface{}) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, err error, fields ...interface{}) {
	globalLogger.Error(msg, err, fields...)
}
