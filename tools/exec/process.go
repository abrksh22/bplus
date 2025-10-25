package exec

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/abrksh22/bplus/tools"
)

// ProcessManager manages background processes.
type ProcessManager struct {
	processes map[string]*BackgroundProcess
	mu        sync.RWMutex
}

// BackgroundProcess represents a background process.
type BackgroundProcess struct {
	ID        string
	Command   string
	Cmd       *exec.Cmd
	StartTime time.Time
	Stdout    *bytes.Buffer
	Stderr    *bytes.Buffer
	Done      bool
	ExitCode  int
	Error     error
	mu        sync.Mutex
}

var globalProcessManager = &ProcessManager{
	processes: make(map[string]*BackgroundProcess),
}

// StartBackgroundProcess starts a process in the background.
func (pm *ProcessManager) StartProcess(id, command, workingDir string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.processes[id]; exists {
		return fmt.Errorf("process %s already exists", id)
	}

	cmd := exec.Command("bash", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	proc := &BackgroundProcess{
		ID:        id,
		Command:   command,
		Cmd:       cmd,
		StartTime: time.Now(),
		Stdout:    stdout,
		Stderr:    stderr,
		Done:      false,
	}

	pm.processes[id] = proc

	// Monitor process in background
	go func() {
		err := cmd.Wait()
		proc.mu.Lock()
		defer proc.mu.Unlock()

		proc.Done = true
		proc.Error = err
		if cmd.ProcessState != nil {
			proc.ExitCode = cmd.ProcessState.ExitCode()
		}
	}()

	return nil
}

// GetProcessOutput gets the output of a background process.
func (pm *ProcessManager) GetProcessOutput(id string, filter *regexp.Regexp) (string, string, bool, error) {
	pm.mu.RLock()
	proc, exists := pm.processes[id]
	pm.mu.RUnlock()

	if !exists {
		return "", "", false, fmt.Errorf("process %s not found", id)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()

	stdout := proc.Stdout.String()
	stderr := proc.Stderr.String()

	// Apply filter if provided
	if filter != nil {
		stdout = filterOutput(stdout, filter)
		stderr = filterOutput(stderr, filter)
	}

	return stdout, stderr, proc.Done, nil
}

// KillProcess kills a background process.
func (pm *ProcessManager) KillProcess(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	proc, exists := pm.processes[id]
	if !exists {
		return fmt.Errorf("process %s not found", id)
	}

	proc.mu.Lock()
	defer proc.mu.Unlock()

	if proc.Done {
		return fmt.Errorf("process %s already finished", id)
	}

	if proc.Cmd.Process != nil {
		if err := proc.Cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	proc.Done = true
	delete(pm.processes, id)

	return nil
}

// ListProcesses lists all background processes.
func (pm *ProcessManager) ListProcesses() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	ids := make([]string, 0, len(pm.processes))
	for id := range pm.processes {
		ids = append(ids, id)
	}
	return ids
}

// filterOutput filters output lines by regex.
func filterOutput(output string, filter *regexp.Regexp) string {
	scanner := bufio.NewScanner(bytes.NewBufferString(output))
	var result bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()
		if filter.MatchString(line) {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// ProcessOutputTool implements getting process output.
type ProcessOutputTool struct{}

// NewProcessOutputTool creates a new ProcessOutput tool.
func NewProcessOutputTool() *ProcessOutputTool {
	return &ProcessOutputTool{}
}

// Name returns the tool name.
func (t *ProcessOutputTool) Name() string {
	return "process_output"
}

// Description returns the tool description.
func (t *ProcessOutputTool) Description() string {
	return "Gets output from a background process"
}

// Parameters returns the tool parameters.
func (t *ProcessOutputTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "process_id",
			Type:        tools.TypeString,
			Required:    true,
			Description: "ID of the background process",
		},
		{
			Name:        "filter",
			Type:        tools.TypeString,
			Required:    false,
			Description: "Regex filter for output lines",
			Default:     "",
		},
	}
}

// RequiresPermission returns true.
func (t *ProcessOutputTool) RequiresPermission() bool {
	return true
}

// Execute gets process output.
func (t *ProcessOutputTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	processID := params["process_id"].(string)
	filterStr := ""

	if val, ok := params["filter"]; ok {
		filterStr = val.(string)
	}

	var filter *regexp.Regexp
	var err error
	if filterStr != "" {
		filter, err = regexp.Compile(filterStr)
		if err != nil {
			return &tools.Result{
				Success: false,
				Error:   fmt.Errorf("invalid filter regex: %w", err),
			}, nil
		}
	}

	stdout, stderr, done, err := globalProcessManager.GetProcessOutput(processID, filter)
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   err,
		}, nil
	}

	return &tools.Result{
		Success: true,
		Output: map[string]interface{}{
			"stdout": stdout,
			"stderr": stderr,
			"done":   done,
		},
		Metadata: map[string]interface{}{
			"process_id": processID,
			"done":       done,
		},
	}, nil
}

// Category returns the tool category.
func (t *ProcessOutputTool) Category() string {
	return "exec"
}

// Version returns the tool version.
func (t *ProcessOutputTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false.
func (t *ProcessOutputTool) IsExternal() bool {
	return false
}

// KillProcessTool implements killing a process.
type KillProcessTool struct{}

// NewKillProcessTool creates a new KillProcess tool.
func NewKillProcessTool() *KillProcessTool {
	return &KillProcessTool{}
}

// Name returns the tool name.
func (t *KillProcessTool) Name() string {
	return "kill_process"
}

// Description returns the tool description.
func (t *KillProcessTool) Description() string {
	return "Kills a background process"
}

// Parameters returns the tool parameters.
func (t *KillProcessTool) Parameters() []tools.Parameter {
	return []tools.Parameter{
		{
			Name:        "process_id",
			Type:        tools.TypeString,
			Required:    true,
			Description: "ID of the background process to kill",
		},
	}
}

// RequiresPermission returns true.
func (t *KillProcessTool) RequiresPermission() bool {
	return true
}

// Execute kills the process.
func (t *KillProcessTool) Execute(ctx context.Context, params map[string]interface{}) (*tools.Result, error) {
	processID := params["process_id"].(string)

	err := globalProcessManager.KillProcess(processID)
	if err != nil {
		return &tools.Result{
			Success: false,
			Error:   err,
		}, nil
	}

	return &tools.Result{
		Success: true,
		Output:  fmt.Sprintf("Process %s killed successfully", processID),
		Metadata: map[string]interface{}{
			"process_id": processID,
		},
	}, nil
}

// Category returns the tool category.
func (t *KillProcessTool) Category() string {
	return "exec"
}

// Version returns the tool version.
func (t *KillProcessTool) Version() string {
	return "1.0.0"
}

// IsExternal returns false.
func (t *KillProcessTool) IsExternal() bool {
	return false
}
