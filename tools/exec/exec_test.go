package exec

import (
	"context"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBashTool tests the Bash tool.
func TestBashTool(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping bash tests on Windows")
	}

	tool := NewBashTool()

	t.Run("Simple command", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"command": "echo 'Hello, World!'",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output.(string), "Hello, World!")
	})

	t.Run("Command with exit code", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"command": "exit 1",
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Error(t, result.Error)

		metadata := result.Metadata
		exitCode := metadata["exit_code"].(int)
		assert.Equal(t, 1, exitCode)
	})

	t.Run("Working directory", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"command":     "pwd",
			"working_dir": "/tmp",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Output.(string), "tmp")
	})

	t.Run("Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		result, err := tool.Execute(ctx, map[string]interface{}{
			"command": "sleep 10",
			"timeout": 1000, // 1 second
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
	})

	t.Run("Dangerous command blocked", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"command": "rm -rf /",
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "dangerous")
	})

	t.Run("Different shell", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"command": "echo 'test'",
			"shell":   "sh",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})
}

// TestProcessManager tests the process manager.
func TestProcessManager(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping process tests on Windows")
	}

	pm := &ProcessManager{
		processes: make(map[string]*BackgroundProcess),
	}

	t.Run("Start and get output", func(t *testing.T) {
		err := pm.StartProcess("test1", "echo 'Hello' && sleep 1 && echo 'World'", "")
		require.NoError(t, err)

		// Wait a bit for output
		time.Sleep(500 * time.Millisecond)

		stdout, stderr, done, err := pm.GetProcessOutput("test1", nil)
		require.NoError(t, err)
		assert.Contains(t, stdout, "Hello")
		assert.Empty(t, stderr)

		// Wait for completion
		time.Sleep(1 * time.Second)

		stdout, stderr, done, err = pm.GetProcessOutput("test1", nil)
		require.NoError(t, err)
		assert.Contains(t, stdout, "World")
		assert.True(t, done)
	})

	t.Run("Filter output", func(t *testing.T) {
		err := pm.StartProcess("test2", "echo 'Line 1' && echo 'Match this' && echo 'Line 3'", "")
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)

		filter := regexp.MustCompile("Match")
		stdout, _, _, err := pm.GetProcessOutput("test2", filter)
		require.NoError(t, err)
		assert.Contains(t, stdout, "Match this")
		assert.NotContains(t, stdout, "Line 1")
	})

	t.Run("Kill process", func(t *testing.T) {
		err := pm.StartProcess("test3", "sleep 100", "")
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		err = pm.KillProcess("test3")
		require.NoError(t, err)

		// Verify process is gone
		_, _, _, err = pm.GetProcessOutput("test3", nil)
		assert.Error(t, err)
	})

	t.Run("List processes", func(t *testing.T) {
		pm.processes = make(map[string]*BackgroundProcess)

		err := pm.StartProcess("proc1", "sleep 1", "")
		require.NoError(t, err)

		err = pm.StartProcess("proc2", "sleep 1", "")
		require.NoError(t, err)

		list := pm.ListProcesses()
		assert.Len(t, list, 2)
		assert.Contains(t, list, "proc1")
		assert.Contains(t, list, "proc2")
	})

	t.Run("Process already exists", func(t *testing.T) {
		pm.processes = make(map[string]*BackgroundProcess)

		err := pm.StartProcess("dup", "echo 'test'", "")
		require.NoError(t, err)

		err = pm.StartProcess("dup", "echo 'test2'", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

// TestProcessOutputTool tests the ProcessOutput tool.
func TestProcessOutputTool(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping process tests on Windows")
	}

	tool := NewProcessOutputTool()

	// Start a test process
	err := globalProcessManager.StartProcess("output_test", "echo 'Test output'", "")
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	t.Run("Get output", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"process_id": "output_test",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)

		output := result.Output.(map[string]interface{})
		stdout := output["stdout"].(string)
		assert.Contains(t, stdout, "Test output")
	})

	t.Run("Process not found", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"process_id": "nonexistent",
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Error(t, result.Error)
	})
}

// TestKillProcessTool tests the KillProcess tool.
func TestKillProcessTool(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping process tests on Windows")
	}

	tool := NewKillProcessTool()

	// Start a test process
	err := globalProcessManager.StartProcess("kill_test", "sleep 100", "")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	t.Run("Kill process", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"process_id": "kill_test",
		})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Process not found", func(t *testing.T) {
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"process_id": "nonexistent",
		})
		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Error(t, result.Error)
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
		{NewBashTool(), "bash", "exec"},
		{NewProcessOutputTool(), "process_output", "exec"},
		{NewKillProcessTool(), "kill_process", "exec"},
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

// TestDangerousCommandDetection tests dangerous command detection.
func TestDangerousCommandDetection(t *testing.T) {
	tests := []struct {
		command   string
		dangerous bool
	}{
		{"echo 'hello'", false},
		{"ls -la", false},
		{"rm -rf /", true},
		{"dd if=/dev/zero of=/dev/sda", true},
		{"mkfs.ext4 /dev/sda", true},
		{":(){:|:&};:", true},
		{"chmod -r 777 /", true},     // Lowercase to match pattern
		{"rm -rf ~/documents", true}, // Contains "rm -rf ~" pattern
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			result := isDangerousCommand(tt.command)
			assert.Equal(t, tt.dangerous, result)
		})
	}
}
