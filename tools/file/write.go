package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/abrksh22/bplus/tools"
)

// WriteTool implements the file writing tool.
type WriteTool struct{}

// NewWriteTool creates a new Write tool.
func NewWriteTool() *WriteTool {
	return &WriteTool{}
}

// Name returns the tool name.
func (t *WriteTool) Name() string {
	return "write"
}

// Description returns the tool description.
func (t *WriteTool) Description() string {
	return "Writes content to a file with atomic writes and automatic backups"
}

// Parameters returns the tool parameters.
func (t *WriteTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "file_path",
			Type:        tools.TypeString,
			Required:    true,
			Description: "Absolute path to the file to write",
		},
		{
			Name:        "content",
			Type:        tools.TypeString,
			Required:    true,
			Description: "Content to write to the file",
		},
		{
			Name:        "create_backup",
			Type:        tools.TypeBool,
			Required:    false,
			Description: "Create backup before overwriting (default: true)",
			Default:     true,
		},
		{
			Name:        "create_dirs",
			Type:        tools.TypeBool,
			Required:    false,
			Description: "Create parent directories if they don't exist (default: true)",
			Default:     true,
		},
	}
}

// RequiresPermission returns true as file writing requires permission.
func (t *WriteTool) RequiresPermission() bool {
	return true
}

// Execute executes the write operation.
func (t *WriteTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	startTime := time.Now()

	// Extract parameters
	filePath := params["file_path"].(string)
	content := params["content"].(string)
	createBackup := true
	createDirs := true

	if val, ok := params["create_backup"]; ok {
		createBackup = val.(bool)
	}

	if val, ok := params["create_dirs"]; ok {
		createDirs = val.(bool)
	}

	// Validate path
	filePath = filepath.Clean(filePath)
	if !filepath.IsAbs(filePath) {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("file_path must be absolute"),
		}, nil
	}

	// Create parent directories if needed
	dir := filepath.Dir(filePath)
	if createDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &tools.Result{
				Success: false,
				Error:   fmt.Errorf("failed to create directories: %w", err),
			}, nil
		}
	}

	// Create backup if file exists
	var backupPath string
	if createBackup {
		if _, err := os.Stat(filePath); err == nil {
			backupPath = filePath + ".backup"
			if err := copyFile(filePath, backupPath); err != nil {
				return &tools.Result{
					Success: false,
					Error:   fmt.Errorf("failed to create backup: %w", err),
				}, nil
			}
		}
	}

	// Write file atomically
	if err := writeFileAtomic(filePath, content); err != nil {
		// Restore from backup if write failed
		if backupPath != "" {
			_ = copyFile(backupPath, filePath)
		}

		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("failed to write file: %w", err),
		}, nil
	}

	return &tools.Result{
		Success: true,
		Output:  fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), filePath),
		Metadata: map[string]interface{}{
			"path":        filePath,
			"size":        len(content),
			"backup_path": backupPath,
		},
		Duration: time.Since(startTime),
	}, nil
}

// Category returns the tool category.
func (t *WriteTool) Category() string {
	return "file"
}

// Version returns the tool version.
func (t *WriteTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false as this is a core tool.
func (t *WriteTool) IsExternal() bool {
	return false
}

// writeFileAtomic writes a file atomically by writing to a temp file first.
func writeFileAtomic(filePath, content string) error {
	// Write to temp file
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tempPath, filePath); err != nil {
		_ = os.Remove(tempPath) // Clean up temp file
		return err
	}

	return nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, content, 0644)
}
