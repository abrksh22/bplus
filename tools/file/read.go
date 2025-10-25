// Package file provides file operation tools.
package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abrksh22/bplus/tools"
)

// ReadTool implements the file reading tool.
type ReadTool struct{}

// NewReadTool creates a new Read tool.
func NewReadTool() *ReadTool {
	return &ReadTool{}
}

// Name returns the tool name.
func (t *ReadTool) Name() string {
	return "read"
}

// Description returns the tool description.
func (t *ReadTool) Description() string {
	return "Reads a file from the filesystem with optional line offset and limit"
}

// Parameters returns the tool parameters.
func (t *ReadTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "file_path",
			Type:        tools.TypeString,
			Required:    true,
			Description: "Absolute path to the file to read",
		},
		{
			Name:        "offset",
			Type:        tools.TypeInt,
			Required:    false,
			Description: "Line number to start reading from (1-indexed)",
			Default:     1,
		},
		{
			Name:        "limit",
			Type:        tools.TypeInt,
			Required:    false,
			Description: "Number of lines to read (0 = all)",
			Default:     0,
		},
	}
}

// RequiresPermission returns true as file reading requires permission.
func (t *ReadTool) RequiresPermission() bool {
	return true
}

// Execute executes the read operation.
func (t *ReadTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	startTime := time.Now()

	// Extract parameters
	filePath := params["file_path"].(string)
	offset := 1
	limit := 0

	if offsetVal, ok := params["offset"]; ok {
		switch v := offsetVal.(type) {
		case int:
			offset = v
		case float64:
			offset = int(v)
		}
	}

	if limitVal, ok := params["limit"]; ok {
		switch v := limitVal.(type) {
		case int:
			limit = v
		case float64:
			limit = int(v)
		}
	}

	// Validate path
	filePath = filepath.Clean(filePath)
	if !filepath.IsAbs(filePath) {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("file_path must be absolute"),
		}, nil
	}

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("failed to stat file: %w", err),
		}, nil
	}

	if info.IsDir() {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("path is a directory, not a file"),
		}, nil
	}

	// Read file content
	content, err := readFileContent(filePath, offset, limit)
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   err,
		}, nil
	}

	return &tools.Result{
		Success: true,
		Output:  content,
		Metadata: map[string]interface{}{
			"path":     filePath,
			"size":     info.Size(),
			"modified": info.ModTime(),
			"offset":   offset,
			"limit":    limit,
		},
		Duration: time.Since(startTime),
	}, nil
}

// Category returns the tool category.
func (t *ReadTool) Category() string {
	return "file"
}

// Version returns the tool version.
func (t *ReadTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false as this is a core tool.
func (t *ReadTool) IsExternal() bool {
	return false
}

// readFileContent reads file content with offset and limit.
func readFileContent(filePath string, offset, limit int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(file)

	// Set larger buffer for long lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 1
	linesRead := 0

	for scanner.Scan() {
		// Skip lines before offset
		if lineNum < offset {
			lineNum++
			continue
		}

		// Check limit
		if limit > 0 && linesRead >= limit {
			break
		}

		// Write line with line number
		line := scanner.Text()
		if len(line) > 2000 {
			line = line[:2000] + "... (truncated)"
		}

		fmt.Fprintf(&builder, "%6d→%s\n", lineNum, line)

		lineNum++
		linesRead++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return builder.String(), nil
}
