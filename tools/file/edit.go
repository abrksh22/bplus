package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abrksh22/bplus/tools"
)

// EditTool implements the file editing tool.
type EditTool struct{}

// NewEditTool creates a new Edit tool.
func NewEditTool() *EditTool {
	return &EditTool{}
}

// Name returns the tool name.
func (t *EditTool) Name() string {
	return "edit"
}

// Description returns the tool description.
func (t *EditTool) Description() string {
	return "Edits a file by performing exact string replacement"
}

// Parameters returns the tool parameters.
func (t *EditTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "file_path",
			Type:        tools.TypeString,
			Required:    true,
			Description: "Absolute path to the file to edit",
		},
		{
			Name:        "old_string",
			Type:        tools.TypeString,
			Required:    true,
			Description: "The exact string to replace",
		},
		{
			Name:        "new_string",
			Type:        tools.TypeString,
			Required:    true,
			Description: "The replacement string",
		},
		{
			Name:        "replace_all",
			Type:        tools.TypeBool,
			Required:    false,
			Description: "Replace all occurrences (default: false, replaces first only)",
			Default:     false,
		},
	}
}

// RequiresPermission returns true as file editing requires permission.
func (t *EditTool) RequiresPermission() bool {
	return true
}

// Execute executes the edit operation.
func (t *EditTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	startTime := time.Now()

	// Extract parameters
	filePath := params["file_path"].(string)
	oldString := params["old_string"].(string)
	newString := params["new_string"].(string)
	replaceAll := false

	if val, ok := params["replace_all"]; ok {
		replaceAll = val.(bool)
	}

	// Validate path
	filePath = filepath.Clean(filePath)
	if !filepath.IsAbs(filePath) {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("file_path must be absolute"),
		}, nil
	}

	// Check if strings are different
	if oldString == newString {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("old_string and new_string must be different"),
		}, nil
	}

	// Read existing content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("failed to read file: %w", err),
		}, nil
	}

	contentStr := string(content)

	// Check if old_string exists
	if !strings.Contains(contentStr, oldString) {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("old_string not found in file"),
		}, nil
	}

	// Perform replacement
	var newContent string
	var replacements int

	if replaceAll {
		replacements = strings.Count(contentStr, oldString)
		newContent = strings.ReplaceAll(contentStr, oldString, newString)
	} else {
		// Count occurrences for validation
		count := strings.Count(contentStr, oldString)
		if count > 1 {
			return &tools.Result{
				Success: false,
				Error:   fmt.Errorf("old_string occurs %d times, must be unique or use replace_all=true", count),
			}, nil
		}
		replacements = 1
		newContent = strings.Replace(contentStr, oldString, newString, 1)
	}

	// Create backup
	backupPath := filePath + ".backup"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("failed to create backup: %w", err),
		}, nil
	}

	// Write new content atomically
	if err := writeFileAtomic(filePath, newContent); err != nil {
		// Restore from backup
		_ = os.WriteFile(filePath, content, 0644)

		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("failed to write edited file: %w", err),
		}, nil
	}

	return &tools.Result{
		Success: true,
		Output:  fmt.Sprintf("Successfully replaced %d occurrence(s) in %s", replacements, filePath),
		Metadata: map[string]interface{}{
			"path":         filePath,
			"replacements": replacements,
			"backup_path":  backupPath,
			"old_size":     len(content),
			"new_size":     len(newContent),
		},
		Duration: time.Since(startTime),
	}, nil
}

// Category returns the tool category.
func (t *EditTool) Category() string {
	return "file"
}

// Version returns the tool version.
func (t *EditTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false as this is a core tool.
func (t *EditTool) IsExternal() bool {
	return false
}
