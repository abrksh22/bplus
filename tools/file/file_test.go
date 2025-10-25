package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadTool tests the Read tool.
func TestReadTool(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	tool := NewReadTool()

	t.Run("Read entire file", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path": testFile,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Output)
	})

	t.Run("Read with offset and limit", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path": testFile,
			"offset":    2,
			"limit":     2,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path": filepath.Join(tmpDir, "nonexistent.txt"),
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Error(t, result.Error)
	})

	t.Run("Directory path", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path": tmpDir,
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
	})
}

// TestWriteTool tests the Write tool.
func TestWriteTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewWriteTool()

	t.Run("Write new file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "new.txt")
		content := "Hello, World!"

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path": testFile,
			"content":   content,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Verify file was written
		written, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, content, string(written))
	})

	t.Run("Overwrite existing file with backup", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "existing.txt")
		original := "original content"
		updated := "updated content"

		// Write original
		err := os.WriteFile(testFile, []byte(original), 0644)
		require.NoError(t, err)

		// Overwrite
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path":     testFile,
			"content":       updated,
			"create_backup": true,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Verify updated content
		written, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, updated, string(written))

		// Verify backup exists
		backupPath := testFile + ".backup"
		backup, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, original, string(backup))
	})

	t.Run("Create directories", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "deep", "nested", "path", "file.txt")
		content := "test"

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path":   testFile,
			"content":     content,
			"create_dirs": true,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Verify file exists
		_, err = os.Stat(testFile)
		assert.NoError(t, err)
	})
}

// TestEditTool tests the Edit tool.
func TestEditTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewEditTool()

	t.Run("Replace unique string", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "edit.txt")
		content := "Hello, World!"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path":  testFile,
			"old_string": "World",
			"new_string": "Go",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Verify edit
		edited, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "Hello, Go!", string(edited))
	})

	t.Run("Replace all occurrences", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "edit_all.txt")
		content := "foo bar foo baz foo"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path":   testFile,
			"old_string":  "foo",
			"new_string":  "qux",
			"replace_all": true,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Verify all replaced
		edited, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "qux bar qux baz qux", string(edited))
	})

	t.Run("Non-unique string without replace_all", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "edit_nonunique.txt")
		content := "foo bar foo"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path":  testFile,
			"old_string": "foo",
			"new_string": "qux",
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Error(t, result.Error)
	})

	t.Run("String not found", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "edit_notfound.txt")
		content := "Hello, World!"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"file_path":  testFile,
			"old_string": "Missing",
			"new_string": "Found",
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
	})
}

// TestGlobTool tests the Glob tool.
func TestGlobTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewGlobTool()

	// Create test files
	files := []string{
		"test1.go",
		"test2.go",
		"test.txt",
		"src/main.go",
		"src/util.go",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		dir := filepath.Dir(path)
		os.MkdirAll(dir, 0755)
		os.WriteFile(path, []byte("test"), 0644)
	}

	t.Run("Simple pattern", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"pattern": "*.go",
			"path":    tmpDir,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		matches := result.Output.([]string)
		assert.Greater(t, len(matches), 0)
	})

	t.Run("Recursive pattern", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"pattern": "**/*.go",
			"path":    tmpDir,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		matches := result.Output.([]string)
		assert.GreaterOrEqual(t, len(matches), 4) // Should find all .go files
	})
}

// TestGrepTool tests the Grep tool.
func TestGrepTool(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewGrepTool()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	os.WriteFile(file1, []byte("Hello World\nGoodbye World\n"), 0644)
	os.WriteFile(file2, []byte("Testing grep\nHello there\n"), 0644)

	t.Run("Files with matches", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"pattern":     "Hello",
			"path":        tmpDir,
			"output_mode": "files_with_matches",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		matches := result.Output.([]string)
		assert.Equal(t, 2, len(matches))
	})

	t.Run("Match count", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"pattern":     "World",
			"path":        tmpDir,
			"output_mode": "count",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		counts := result.Output.(map[string]int)
		assert.Equal(t, 2, counts[file1]) // "World" appears twice in file1
	})

	t.Run("Content with context", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"pattern":        "Hello",
			"path":           tmpDir,
			"output_mode":    "content",
			"context_before": 1,
			"context_after":  1,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		matches := result.Output.([]map[string]interface{})
		assert.Greater(t, len(matches), 0)
	})

	t.Run("Case insensitive", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"pattern":          "hello",
			"path":             tmpDir,
			"output_mode":      "files_with_matches",
			"case_insensitive": true,
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		matches := result.Output.([]string)
		assert.Equal(t, 2, len(matches))
	})
}

// TestToolMetadata tests tool metadata methods.
func TestToolMetadata(t *testing.T) {
	tools := []struct {
		tool interface {
			Name() string
			Description() string
			Category() string
			Version() string
			IsExternal() bool
			RequiresPermission() bool
		}
		expectedName     string
		expectedCategory string
	}{
		{NewReadTool(), "read", "file"},
		{NewWriteTool(), "write", "file"},
		{NewEditTool(), "edit", "file"},
		{NewGlobTool(), "glob", "file"},
		{NewGrepTool(), "grep", "file"},
	}

	for _, tt := range tools {
		t.Run(tt.expectedName, func(t *testing.T) {
			assert.Equal(t, tt.expectedName, tt.tool.Name())
			assert.Equal(t, tt.expectedCategory, tt.tool.Category())
			assert.Equal(t, "1.0.0", tt.tool.Version())
			assert.False(t, tt.tool.IsExternal())
			assert.True(t, tt.tool.RequiresPermission())
			assert.NotEmpty(t, tt.tool.Description())
		})
	}
}
