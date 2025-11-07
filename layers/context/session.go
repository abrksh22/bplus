package context

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
	"github.com/abrksh22/bplus/models"
)

// SessionExporter handles session export and import
type SessionExporter struct {
	logger *logging.Logger
}

// NewSessionExporter creates a new session exporter
func NewSessionExporter() *SessionExporter {
	return &SessionExporter{
		logger: logging.NewDefaultLogger().WithComponent("session_exporter"),
	}
}

// SessionExport represents an exported session
type SessionExport struct {
	Version     string                 `json:"version"`
	ExportedAt  time.Time              `json:"exported_at"`
	Session     SessionExportData      `json:"session"`
	Messages    []models.Message       `json:"messages"`
	Context     ContextSnapshot        `json:"context"`
	Checkpoints []Checkpoint           `json:"checkpoints,omitempty"`
	Files       map[string]string      `json:"files,omitempty"` // path -> content
	Config      map[string]interface{} `json:"config,omitempty"`
}

// SessionExportData represents session metadata
type SessionExportData struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	TotalTokens int                    `json:"total_tokens"`
	TotalCost   float64                `json:"total_cost"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ExportOptions configures export behavior
type ExportOptions struct {
	IncludeMessages    bool
	IncludeContext     bool
	IncludeCheckpoints bool
	IncludeFiles       bool
	IncludeConfig      bool
	FilePaths          []string // Specific files to include
}

// ExportSession exports a session to JSON
func (se *SessionExporter) ExportSession(ctx context.Context, sessionID string, snapshot *ContextSnapshot, messages []models.Message, opts ExportOptions) (*SessionExport, error) {
	export := &SessionExport{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Session: SessionExportData{
			ID:       sessionID,
			Metadata: make(map[string]interface{}),
		},
	}

	// Add messages if requested
	if opts.IncludeMessages {
		export.Messages = messages

		// Calculate totals
		for _, msg := range messages {
			// This is simplified - in reality, we'd parse message metadata
			export.Session.TotalTokens += 100 // Placeholder
		}
	}

	// Add context if requested
	if opts.IncludeContext && snapshot != nil {
		export.Context = *snapshot
		export.Session.TotalTokens = snapshot.TotalTokens
	}

	// Add files if requested
	if opts.IncludeFiles && len(opts.FilePaths) > 0 {
		export.Files = make(map[string]string)
		for _, path := range opts.FilePaths {
			content, err := os.ReadFile(path)
			if err != nil {
				se.logger.Warn("Failed to read file for export", "path", path, "error", err)
				continue
			}
			export.Files[path] = string(content)
		}
	}

	se.logger.Info("Session exported",
		"session_id", sessionID,
		"messages", len(export.Messages),
		"files", len(export.Files),
		"tokens", export.Session.TotalTokens)

	return export, nil
}

// ExportToFile exports session to a JSON file
func (se *SessionExporter) ExportToFile(ctx context.Context, export *SessionExport, filePath string) error {
	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to marshal export")
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.Wrap(err, errors.ErrCodeFileWrite, "failed to write export file")
	}

	se.logger.Info("Session exported to file",
		"session_id", export.Session.ID,
		"file", filePath,
		"size", len(data))

	return nil
}

// ImportFromFile imports a session from a JSON file
func (se *SessionExporter) ImportFromFile(ctx context.Context, filePath string) (*SessionExport, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeFileRead, "failed to read import file")
	}

	// Unmarshal JSON
	var export SessionExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to unmarshal import")
	}

	// Validate version
	if export.Version != "1.0" {
		return nil, errors.New(errors.ErrCodeInvalidInput, fmt.Sprintf("unsupported export version: %s", export.Version))
	}

	se.logger.Info("Session imported from file",
		"session_id", export.Session.ID,
		"file", filePath,
		"messages", len(export.Messages))

	return &export, nil
}

// ImportSession imports a session export into the system
func (se *SessionExporter) ImportSession(ctx context.Context, export *SessionExport, newSessionID string) error {
	// Validate export
	if export.Session.ID == "" {
		return errors.New(errors.ErrCodeInvalidInput, "export missing session ID")
	}

	// If newSessionID is provided, use it (allows importing to new session)
	if newSessionID != "" {
		export.Session.ID = newSessionID
	}

	// TODO: Actually import into database
	// For now, just validate

	se.logger.Info("Session import validated",
		"original_id", export.Session.ID,
		"new_id", newSessionID,
		"messages", len(export.Messages))

	return nil
}

// CreateShareableSnapshot creates a read-only shareable snapshot
func (se *SessionExporter) CreateShareableSnapshot(ctx context.Context, sessionID string, snapshot *ContextSnapshot, messages []models.Message) (*SessionExport, error) {
	opts := ExportOptions{
		IncludeMessages: true,
		IncludeContext:  true,
		IncludeFiles:    false, // Don't include file contents for sharing
		IncludeConfig:   false, // Don't include sensitive config
	}

	export, err := se.ExportSession(ctx, sessionID, snapshot, messages, opts)
	if err != nil {
		return nil, err
	}

	// Anonymize sensitive data
	export = se.anonymizeExport(export)

	return export, nil
}

// anonymizeExport removes sensitive information from export
func (se *SessionExporter) anonymizeExport(export *SessionExport) *SessionExport {
	// Create a copy
	anonymized := *export

	// Remove sensitive metadata
	if anonymized.Session.Metadata != nil {
		delete(anonymized.Session.Metadata, "api_key")
		delete(anonymized.Session.Metadata, "token")
		delete(anonymized.Session.Metadata, "password")
		delete(anonymized.Session.Metadata, "secret")
	}

	// Clear file contents if any
	anonymized.Files = nil

	// Clear config
	anonymized.Config = nil

	return &anonymized
}
