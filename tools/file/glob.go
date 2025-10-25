package file

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/abrksh22/bplus/tools"
)

// GlobTool implements the file pattern matching tool.
type GlobTool struct{}

// NewGlobTool creates a new Glob tool.
func NewGlobTool() *GlobTool {
	return &GlobTool{}
}

// Name returns the tool name.
func (t *GlobTool) Name() string {
	return "glob"
}

// Description returns the tool description.
func (t *GlobTool) Description() string {
	return "Finds files matching a glob pattern, sorted by modification time"
}

// Parameters returns the tool parameters.
func (t *GlobTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "pattern",
			Type:        tools.TypeString,
			Required:    true,
			Description: "Glob pattern to match files (e.g., '**/*.go', 'src/**/*.ts')",
		},
		{
			Name:        "path",
			Type:        tools.TypeString,
			Required:    false,
			Description: "Directory to search in (defaults to current directory)",
			Default:     ".",
		},
		{
			Name:        "respect_gitignore",
			Type:        tools.TypeBool,
			Required:    false,
			Description: "Respect .gitignore and .bplusignore files (default: true)",
			Default:     true,
		},
	}
}

// RequiresPermission returns true as file globbing requires permission.
func (t *GlobTool) RequiresPermission() bool {
	return true
}

// Execute executes the glob operation.
func (t *GlobTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	startTime := time.Now()

	// Extract parameters
	pattern := params["pattern"].(string)
	searchPath := "."
	respectGitignore := true

	if val, ok := params["path"]; ok {
		searchPath = val.(string)
	}

	if val, ok := params["respect_gitignore"]; ok {
		respectGitignore = val.(bool)
	}

	// Clean path
	searchPath = filepath.Clean(searchPath)

	// Load ignore patterns if requested
	var ignorePatterns []string
	if respectGitignore {
		ignorePatterns = loadIgnorePatterns(searchPath)
	}

	// Find matching files
	matches, err := globFiles(searchPath, pattern, ignorePatterns)
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   err,
		}, nil
	}

	// Sort by modification time (newest first)
	sortByModTime(matches)

	return &tools.Result{
		Success: true,
		Output:  matches,
		Metadata: map[string]interface{}{
			"pattern": pattern,
			"path":    searchPath,
			"count":   len(matches),
		},
		Duration: time.Since(startTime),
	}, nil
}

// Category returns the tool category.
func (t *GlobTool) Category() string {
	return "file"
}

// Version returns the tool version.
func (t *GlobTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false as this is a core tool.
func (t *GlobTool) IsExternal() bool {
	return false
}

// globFiles finds all files matching the pattern.
func globFiles(searchPath, pattern string, ignorePatterns []string) ([]string, error) {
	var matches []string

	// Handle ** patterns by walking the directory
	if strings.Contains(pattern, "**") {
		basePath := searchPath
		suffixPattern := pattern

		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			// Skip directories
			if info.IsDir() {
				// Check if directory should be ignored
				if shouldIgnore(path, ignorePatterns) {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if file should be ignored
			if shouldIgnore(path, ignorePatterns) {
				return nil
			}

			// Match against pattern
			relPath, _ := filepath.Rel(basePath, path)
			matched, _ := filepath.Match(strings.ReplaceAll(suffixPattern, "**", "*"), relPath)
			if matched || matchGlobPattern(relPath, suffixPattern) {
				matches = append(matches, path)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	} else {
		// Simple glob without **
		fullPattern := filepath.Join(searchPath, pattern)
		files, err := filepath.Glob(fullPattern)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if !shouldIgnore(file, ignorePatterns) {
				matches = append(matches, file)
			}
		}
	}

	return matches, nil
}

// matchGlobPattern matches a path against a glob pattern with ** support.
func matchGlobPattern(path, pattern string) bool {
	// Simple ** matching implementation
	parts := strings.Split(pattern, "**")
	if len(parts) == 1 {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// Check prefix and suffix
	if len(parts) == 2 {
		prefix := parts[0]
		suffix := parts[1]

		if prefix != "" && !strings.HasPrefix(path, strings.TrimSuffix(prefix, "/")) {
			return false
		}

		if suffix != "" {
			suffix = strings.TrimPrefix(suffix, "/")
			if suffix != "" {
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				return matched
			}
		}

		return true
	}

	return false
}

// loadIgnorePatterns loads .gitignore and .bplusignore patterns.
func loadIgnorePatterns(searchPath string) []string {
	var patterns []string

	// Standard patterns
	patterns = append(patterns, ".git", ".git/**", "node_modules", "node_modules/**")

	// Load .gitignore
	gitignorePath := filepath.Join(searchPath, ".gitignore")
	if content, err := os.ReadFile(gitignorePath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				patterns = append(patterns, line)
			}
		}
	}

	// Load .bplusignore
	bplusignorePath := filepath.Join(searchPath, ".bplusignore")
	if content, err := os.ReadFile(bplusignorePath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				patterns = append(patterns, line)
			}
		}
	}

	return patterns
}

// shouldIgnore checks if a path should be ignored based on patterns.
func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}

		// Check if any path component matches
		if strings.Contains(path, string(filepath.Separator)+pattern+string(filepath.Separator)) {
			return true
		}

		if strings.HasSuffix(path, string(filepath.Separator)+pattern) {
			return true
		}
	}

	return false
}

// sortByModTime sorts files by modification time (newest first).
func sortByModTime(files []string) {
	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])

		if errI != nil || errJ != nil {
			return false
		}

		return infoI.ModTime().After(infoJ.ModTime())
	})
}
