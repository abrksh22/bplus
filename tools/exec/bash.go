// Package exec provides command execution tools.
package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/abrksh22/bplus/tools"
)

// BashTool implements the command execution tool.
type BashTool struct{}

// NewBashTool creates a new Bash tool.
func NewBashTool() *BashTool {
	return &BashTool{}
}

// Name returns the tool name.
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns the tool description.
func (t *BashTool) Description() string {
	return "Executes a bash command with timeout and safety checks"
}

// Parameters returns the tool parameters.
func (t *BashTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "command",
			Type:        tools.TypeString,
			Required:    true,
			Description: "The command to execute",
		},
		{
			Name:        "working_dir",
			Type:        tools.TypeString,
			Required:    false,
			Description: "Working directory for command execution",
			Default:     "",
		},
		{
			Name:        "timeout",
			Type:        tools.TypeInt,
			Required:    false,
			Description: "Timeout in milliseconds (default: 120000, max: 600000)",
			Default:     120000,
		},
		{
			Name:        "shell",
			Type:        tools.TypeString,
			Required:    false,
			Description: "Shell to use (bash, zsh, sh, pwsh)",
			Default:     "bash",
		},
	}
}

// RequiresPermission returns true as command execution requires permission.
func (t *BashTool) RequiresPermission() bool {
	return true
}

// Execute executes the bash command.
func (t *BashTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	startTime := time.Now()

	// Extract parameters
	command := params["command"].(string)
	workingDir := ""
	timeoutMs := 120000
	shell := "bash"

	if val, ok := params["working_dir"]; ok {
		workingDir = val.(string)
	}
	if val, ok := params["timeout"]; ok {
		switch v := val.(type) {
		case int:
			timeoutMs = v
		case float64:
			timeoutMs = int(v)
		}
	}
	if val, ok := params["shell"]; ok {
		shell = val.(string)
	}

	// Validate timeout
	if timeoutMs > 600000 {
		timeoutMs = 600000
	}
	if timeoutMs < 1000 {
		timeoutMs = 1000
	}

	// Safety checks for dangerous commands
	if isDangerousCommand(command) {
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("command blocked: potentially dangerous operation detected"),
			Metadata: map[string]interface{}{
				"command": command,
				"reason":  "safety_check_failed",
			},
		}, nil
	}

	// Create timeout context
	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond
	cmdCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	// Determine shell command
	var cmd *exec.Cmd
	switch shell {
	case "bash":
		cmd = exec.CommandContext(cmdCtx, "bash", "-c", command)
	case "zsh":
		cmd = exec.CommandContext(cmdCtx, "zsh", "-c", command)
	case "sh":
		cmd = exec.CommandContext(cmdCtx, "sh", "-c", command)
	case "pwsh", "powershell":
		cmd = exec.CommandContext(cmdCtx, "pwsh", "-Command", command)
	default:
		return &tools.Result{
			Success: false,
			Error:   fmt.Errorf("unsupported shell: %s", shell),
		}, nil
	}

	// Set working directory
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Truncate output if too long
	const maxOutputLen = 30000
	if len(stdoutStr) > maxOutputLen {
		stdoutStr = stdoutStr[:maxOutputLen] + "\n... (output truncated)"
	}
	if len(stderrStr) > maxOutputLen {
		stderrStr = stderrStr[:maxOutputLen] + "\n... (output truncated)"
	}

	// Prepare result
	result := &tools.Result{
		Success: err == nil,
		Metadata: map[string]interface{}{
			"command":     command,
			"shell":       shell,
			"working_dir": workingDir,
			"stdout":      stdoutStr,
			"stderr":      stderrStr,
			"exit_code":   cmd.ProcessState.ExitCode(),
		},
		Duration: time.Since(startTime),
	}

	if err != nil {
		result.Error = fmt.Errorf("command failed: %w", err)
		result.Output = map[string]interface{}{
			"stdout":    stdoutStr,
			"stderr":    stderrStr,
			"exit_code": cmd.ProcessState.ExitCode(),
		}
	} else {
		result.Output = stdoutStr
	}

	return result, nil
}

// Category returns the tool category.
func (t *BashTool) Category() string {
	return "exec"
}

// Version returns the tool version.
func (t *BashTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false as this is a core tool.
func (t *BashTool) IsExternal() bool {
	return false
}

// isDangerousCommand checks if a command is potentially dangerous.
func isDangerousCommand(command string) bool {
	cmd := strings.ToLower(strings.TrimSpace(command))

	// List of dangerous patterns (all lowercase since we lowercase the command)
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		"rm -rf ~",
		"dd if=/dev/zero",
		"dd if=/dev/random",
		"> /dev/sda",
		"mkfs",
		"format c:",
		":(){:|:&};:", // Fork bomb
		"chmod -r 777 /",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmd, pattern) {
			return true
		}
	}

	return false
}
