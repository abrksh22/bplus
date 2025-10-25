package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Registry manages all available tools with support for namespacing and plugins.
type Registry struct {
	tools map[string]Tool // Namespaced tool name -> tool
	mu    sync.RWMutex
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool in the registry with automatic namespacing.
// Core tools are registered as "core.name", plugins as "plugin.name".
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()

	// Add namespace prefix if not already present
	if !strings.Contains(name, ".") {
		// Determine namespace based on whether it's external
		namespace := "core"
		if tool.IsExternal() {
			namespace = "plugin"
		}
		name = namespace + "." + name
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// RegisterWithNamespace registers a tool with an explicit namespace.
func (r *Registry) RegisterWithNamespace(namespace string, tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := namespace + "." + tool.Name()

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name. Supports both namespaced and non-namespaced names.
// For non-namespaced names, tries "core." prefix first, then "plugin.".
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try exact match first
	if tool, exists := r.tools[name]; exists {
		return tool, nil
	}

	// If name doesn't contain namespace, try common namespaces
	if !strings.Contains(name, ".") {
		// Try core namespace
		if tool, exists := r.tools["core."+name]; exists {
			return tool, nil
		}
		// Try plugin namespace
		if tool, exists := r.tools["plugin."+name]; exists {
			return tool, nil
		}
	}

	return nil, fmt.Errorf("tool %s not found", name)
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ListByCategory returns all tools in a specific category.
func (r *Registry) ListByCategory(category string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tools []Tool
	for _, tool := range r.tools {
		if tool.Category() == category {
			tools = append(tools, tool)
		}
	}
	return tools
}

// ListByNamespace returns all tools in a specific namespace.
func (r *Registry) ListByNamespace(namespace string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix := namespace + "."
	var tools []Tool
	for name, tool := range r.tools {
		if strings.HasPrefix(name, prefix) {
			tools = append(tools, tool)
		}
	}
	return tools
}

// Unregister removes a tool from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Try exact match
	if _, exists := r.tools[name]; exists {
		delete(r.tools, name)
		return nil
	}

	// Try with core namespace
	if !strings.Contains(name, ".") {
		coreName := "core." + name
		if _, exists := r.tools[coreName]; exists {
			delete(r.tools, coreName)
			return nil
		}
	}

	return fmt.Errorf("tool %s not found", name)
}

// Execute executes a tool with the given parameters and context.
// Includes permission checking and audit logging.
func (r *Registry) Execute(ctx context.Context, toolName string, params map[string]interface{}, execCtx *ExecutionContext) (*Result, error) {
	tool, err := r.Get(toolName)
	if err != nil {
		return nil, err
	}

	// Validate parameters
	if err := ValidateParameters(params, tool.Parameters()); err != nil {
		return nil, err
	}

	// Check permissions
	if tool.RequiresPermission() {
		if !r.checkPermission(tool, execCtx) {
			return nil, fmt.Errorf("permission denied for tool %s", toolName)
		}
	}

	// Create audit entry
	startTime := time.Now()
	auditEntry := AuditEntry{
		Timestamp:  startTime,
		ToolName:   toolName,
		Parameters: params,
	}

	// Execute tool
	result, err := tool.Execute(ctx, params)

	// Complete audit entry
	auditEntry.Duration = time.Since(startTime)
	auditEntry.Success = (err == nil && result != nil && result.Success)
	auditEntry.Error = err

	// Add to execution context audit trail
	if execCtx != nil {
		execCtx.AuditTrail = append(execCtx.AuditTrail, auditEntry)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = auditEntry.Duration
	return result, nil
}

// checkPermission checks if a tool execution is permitted.
func (r *Registry) checkPermission(tool Tool, execCtx *ExecutionContext) bool {
	if execCtx == nil {
		return false
	}

	// Check if the tool's category is in granted permissions
	category := tool.Category()
	for _, grant := range execCtx.PermissionGrants {
		if grant == category || grant == "*" {
			return true
		}
	}

	return false
}

// GetToolInfo returns detailed information about a tool.
func (r *Registry) GetToolInfo(name string) (*ToolInfo, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	return &ToolInfo{
		Name:               tool.Name(),
		Description:        tool.Description(),
		Parameters:         tool.Parameters(),
		Category:           tool.Category(),
		Version:            tool.Version(),
		IsExternal:         tool.IsExternal(),
		RequiresPermission: tool.RequiresPermission(),
	}, nil
}

// ToolInfo provides detailed information about a tool.
type ToolInfo struct {
	Name               string
	Description        string
	Parameters         []Parameter
	Category           string
	Version            string
	IsExternal         bool
	RequiresPermission bool
}

// ValidateVersion checks if a tool version is compatible with required version.
// Uses semantic versioning comparison.
func ValidateVersion(toolVersion, requiredVersion string) bool {
	// Simplified version check - in production, use a proper semver library
	return toolVersion >= requiredVersion
}
