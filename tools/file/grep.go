package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/abrksh22/bplus/tools"
)

// GrepTool implements the content search tool.
type GrepTool struct{}

// NewGrepTool creates a new Grep tool.
func NewGrepTool() *GrepTool {
	return &GrepTool{}
}

// Name returns the tool name.
func (t *GrepTool) Name() string {
	return "grep"
}

// Description returns the tool description.
func (t *GrepTool) Description() string {
	return "Searches for pattern in files with ripgrep-style functionality"
}

// Parameters returns the tool parameters.
func (t *GrepTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "pattern",
			Type:        tools.TypeString,
			Required:    true,
			Description: "Regular expression pattern to search for",
		},
		{
			Name:        "path",
			Type:        tools.TypeString,
			Required:    false,
			Description: "File or directory to search in (defaults to current directory)",
			Default:     ".",
		},
		{
			Name:        "output_mode",
			Type:        tools.TypeString,
			Required:    false,
			Description: "Output mode: content, files_with_matches, count",
			Default:     "files_with_matches",
		},
		{
			Name:        "case_insensitive",
			Type:        tools.TypeBool,
			Required:    false,
			Description: "Case insensitive search",
			Default:     false,
		},
		{
			Name:        "context_before",
			Type:        tools.TypeInt,
			Required:    false,
			Description: "Number of lines to show before match (-B)",
			Default:     0,
		},
		{
			Name:        "context_after",
			Type:        tools.TypeInt,
			Required:    false,
			Description: "Number of lines to show after match (-A)",
			Default:     0,
		},
		{
			Name:        "show_line_numbers",
			Type:        tools.TypeBool,
			Required:    false,
			Description: "Show line numbers in output (-n)",
			Default:     false,
		},
		{
			Name:        "file_glob",
			Type:        tools.TypeString,
			Required:    false,
			Description: "Glob pattern to filter files (e.g., '*.go')",
			Default:     "",
		},
	}
}

// RequiresPermission returns true as grep requires permission.
func (t *GrepTool) RequiresPermission() bool {
	return true
}

// Execute executes the grep operation.
func (t *GrepTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	startTime := time.Now()

	// Extract parameters
	pattern := params["pattern"].(string)
	searchPath := "."
	outputMode := "files_with_matches"
	caseInsensitive := false
	contextBefore := 0
	contextAfter := 0
	showLineNumbers := false
	fileGlob := ""

	if val, ok := params["path"]; ok {
		searchPath = val.(string)
	}
	if val, ok := params["output_mode"]; ok {
		outputMode = val.(string)
	}
	if val, ok := params["case_insensitive"]; ok {
		caseInsensitive = val.(bool)
	}
	if val, ok := params["context_before"]; ok {
		switch v := val.(type) {
		case int:
			contextBefore = v
		case float64:
			contextBefore = int(v)
		}
	}
	if val, ok := params["context_after"]; ok {
		switch v := val.(type) {
		case int:
			contextAfter = v
		case float64:
			contextAfter = int(v)
		}
	}
	if val, ok := params["show_line_numbers"]; ok {
		showLineNumbers = val.(bool)
	}
	if val, ok := params["file_glob"]; ok {
		fileGlob = val.(string)
	}

	// Compile regex
	var re *regexp.Regexp
	var err error
	if caseInsensitive {
		re, err = regexp.Compile("(?i)" + pattern)
	} else {
		re, err = regexp.Compile(pattern)
	}
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("invalid regex pattern: %w", err),
		}, nil
	}

	// Perform search
	results, err := grepSearch(searchPath, re, outputMode, fileGlob, contextBefore, contextAfter, showLineNumbers)
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   err,
		}, nil
	}

	return &tools.Result{
		Success: true,
		Output:  results,
		Metadata: map[string]interface{}{
			"pattern":     pattern,
			"path":        searchPath,
			"output_mode": outputMode,
			"match_count": countMatches(results),
		},
		Duration: time.Since(startTime),
	}, nil
}

// Category returns the tool category.
func (t *GrepTool) Category() string {
	return "file"
}

// Version returns the tool version.
func (t *GrepTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false as this is a core tool.
func (t *GrepTool) IsExternal() bool {
	return false
}

// grepSearch performs the search operation.
func grepSearch(searchPath string, re *regexp.Regexp, outputMode, fileGlob string, contextBefore, contextAfter int, showLineNumbers bool) (interface{}, error) {
	info, err := os.Stat(searchPath)
	if err != nil {
		return nil, err
	}

	var files []string
	if info.IsDir() {
		// Walk directory
		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if info.IsDir() {
				return nil
			}

			// Check file glob
			if fileGlob != "" {
				matched, _ := filepath.Match(fileGlob, filepath.Base(path))
				if !matched {
					return nil
				}
			}

			// Skip binary files (simple check)
			if !isTextFile(path) {
				return nil
			}

			files = append(files, path)
			return nil
		})

		if err != nil {
			return nil, err
		}
	} else {
		files = []string{searchPath}
	}

	// Search in files based on output mode
	switch outputMode {
	case "files_with_matches":
		return filesWithMatches(files, re)

	case "count":
		return matchCounts(files, re)

	case "content":
		return contentMatches(files, re, contextBefore, contextAfter, showLineNumbers)

	default:
		return nil, fmt.Errorf("invalid output_mode: %s", outputMode)
	}
}

// filesWithMatches returns list of files containing matches.
func filesWithMatches(files []string, re *regexp.Regexp) ([]string, error) {
	var matches []string

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue // Skip unreadable files
		}

		if re.Match(content) {
			matches = append(matches, file)
		}
	}

	return matches, nil
}

// matchCounts returns match counts per file.
func matchCounts(files []string, re *regexp.Regexp) (map[string]int, error) {
	counts := make(map[string]int)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		matches := re.FindAll(content, -1)
		if len(matches) > 0 {
			counts[file] = len(matches)
		}
	}

	return counts, nil
}

// contentMatches returns matching lines with context.
func contentMatches(files []string, re *regexp.Regexp, contextBefore, contextAfter int, showLineNumbers bool) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		lineNum := 0
		var lines []string

		// Read all lines for context support
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		f.Close()

		// Find matches
		for i, line := range lines {
			lineNum = i + 1

			if re.MatchString(line) {
				match := map[string]interface{}{
					"file": file,
					"line": lineNum,
					"text": line,
				}

				// Add context lines
				if contextBefore > 0 || contextAfter > 0 {
					context := make([]string, 0)

					// Before context
					for j := max(0, i-contextBefore); j < i; j++ {
						context = append(context, lines[j])
					}

					// Match line
					context = append(context, line)

					// After context
					for j := i + 1; j < min(len(lines), i+contextAfter+1); j++ {
						context = append(context, lines[j])
					}

					match["context"] = context
				}

				results = append(results, match)
			}
		}
	}

	return results, nil
}

// isTextFile performs a simple check if file is text.
func isTextFile(path string) bool {
	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	textExts := map[string]bool{
		".go": true, ".txt": true, ".md": true, ".js": true, ".ts": true,
		".jsx": true, ".tsx": true, ".json": true, ".yaml": true, ".yml": true,
		".toml": true, ".xml": true, ".html": true, ".css": true, ".scss": true,
		".py": true, ".rb": true, ".java": true, ".c": true, ".cpp": true,
		".h": true, ".hpp": true, ".rs": true, ".sh": true, ".bash": true,
	}

	return textExts[ext]
}

// countMatches counts the total number of matches in results.
func countMatches(results interface{}) int {
	switch v := results.(type) {
	case []string:
		return len(v)
	case map[string]int:
		total := 0
		for _, count := range v {
			total += count
		}
		return total
	case []map[string]interface{}:
		return len(v)
	}
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
