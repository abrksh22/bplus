// Package security provides security and permission management.
package security

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Permission represents a permission category.
type Permission string

const (
	PermissionRead    Permission = "read"    // File reading
	PermissionWrite   Permission = "write"   // File writing/modification
	PermissionExecute Permission = "execute" // Command execution
	PermissionNetwork Permission = "network" // Network access
	PermissionMCP     Permission = "mcp"     // MCP tool execution
	PermissionAll     Permission = "*"       // All permissions
)

// PermissionManager handles permission checking and granting.
type PermissionManager struct {
	grants        map[Permission]bool // Granted permissions
	mode          PermissionMode      // Permission mode
	promptHandler PromptHandler       // Handler for permission prompts
	auditLog      []AuditEntry        // Audit log
	mu            sync.RWMutex
}

// PermissionMode defines how permissions are handled.
type PermissionMode string

const (
	ModeInteractive PermissionMode = "interactive" // Prompt for each permission
	ModeYOLO        PermissionMode = "yolo"        // Auto-approve all
	ModeAutoApprove PermissionMode = "auto"        // Auto-approve safe operations
	ModeDeny        PermissionMode = "deny"        // Deny all
)

// PromptHandler is called to request permission from the user.
type PromptHandler func(ctx context.Context, req *PermissionRequest) (bool, error)

// PermissionRequest represents a permission request.
type PermissionRequest struct {
	Permission  Permission // Permission being requested
	Resource    string     // Resource being accessed (file path, command, etc.)
	Operation   string     // Operation being performed
	Reason      string     // Why this permission is needed
	Risk        RiskLevel  // Risk assessment
	ToolName    string     // Tool requesting permission
	RequestedAt time.Time  // When permission was requested
}

// RiskLevel represents the risk level of an operation.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"    // Safe operation
	RiskMedium RiskLevel = "medium" // Potentially risky
	RiskHigh   RiskLevel = "high"   // Dangerous operation
)

// AuditEntry records a permission check for audit purposes.
type AuditEntry struct {
	Timestamp  time.Time
	Permission Permission
	Resource   string
	Operation  string
	Granted    bool
	Mode       PermissionMode
	ToolName   string
}

// NewPermissionManager creates a new permission manager.
func NewPermissionManager(mode PermissionMode, handler PromptHandler) *PermissionManager {
	return &PermissionManager{
		grants:        make(map[Permission]bool),
		mode:          mode,
		promptHandler: handler,
		auditLog:      make([]AuditEntry, 0),
	}
}

// Check checks if a permission is granted.
func (pm *PermissionManager) Check(ctx context.Context, req *PermissionRequest) (bool, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check mode-specific behavior
	switch pm.mode {
	case ModeYOLO:
		pm.logAudit(req, true)
		return true, nil

	case ModeDeny:
		pm.logAudit(req, false)
		return false, nil

	case ModeAutoApprove:
		// Auto-approve low-risk operations
		if req.Risk == RiskLow {
			pm.logAudit(req, true)
			return true, nil
		}
		// Fall through to interactive for higher risk
		fallthrough

	case ModeInteractive:
		// Check if already granted
		if pm.grants[req.Permission] || pm.grants[PermissionAll] {
			pm.logAudit(req, true)
			return true, nil
		}

		// Prompt user
		if pm.promptHandler != nil {
			granted, err := pm.promptHandler(ctx, req)
			if err != nil {
				return false, err
			}

			if granted {
				pm.grants[req.Permission] = true
			}

			pm.logAudit(req, granted)
			return granted, nil
		}

		// No prompt handler and not granted - return false without error
		pm.logAudit(req, false)
		return false, nil
	}

	pm.logAudit(req, false)
	return false, nil
}

// Grant explicitly grants a permission.
func (pm *PermissionManager) Grant(permission Permission) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.grants[permission] = true
}

// Revoke explicitly revokes a permission.
func (pm *PermissionManager) Revoke(permission Permission) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.grants, permission)
}

// GrantAll grants all permissions.
func (pm *PermissionManager) GrantAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.grants[PermissionAll] = true
}

// RevokeAll revokes all permissions.
func (pm *PermissionManager) RevokeAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.grants = make(map[Permission]bool)
}

// GetAuditLog returns the audit log.
func (pm *PermissionManager) GetAuditLog() []AuditEntry {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy
	log := make([]AuditEntry, len(pm.auditLog))
	copy(log, pm.auditLog)
	return log
}

// logAudit adds an entry to the audit log.
func (pm *PermissionManager) logAudit(req *PermissionRequest, granted bool) {
	entry := AuditEntry{
		Timestamp:  time.Now(),
		Permission: req.Permission,
		Resource:   req.Resource,
		Operation:  req.Operation,
		Granted:    granted,
		Mode:       pm.mode,
		ToolName:   req.ToolName,
	}

	pm.auditLog = append(pm.auditLog, entry)
}

// AssessRisk assesses the risk level of an operation.
func AssessRisk(operation string, resource string) RiskLevel {
	operation = strings.ToLower(operation)
	resource = strings.ToLower(resource)

	// High-risk patterns
	highRiskPatterns := []string{
		"rm -rf",
		"delete",
		"drop",
		"truncate",
		"format",
		"sudo",
		"chmod 777",
		"system",
		"eval",
	}

	for _, pattern := range highRiskPatterns {
		if strings.Contains(operation, pattern) || strings.Contains(resource, pattern) {
			return RiskHigh
		}
	}

	// Medium-risk patterns
	mediumRiskPatterns := []string{
		"write",
		"modify",
		"execute",
		"chmod",
		"chown",
		"git push",
	}

	for _, pattern := range mediumRiskPatterns {
		if strings.Contains(operation, pattern) {
			return RiskMedium
		}
	}

	// Default to low risk
	return RiskLow
}

// ValidateResource validates that a resource path is safe.
func ValidateResource(resource string) error {
	// Check for path traversal attempts
	if strings.Contains(resource, "..") {
		return fmt.Errorf("path traversal detected: %s", resource)
	}

	// Check for absolute paths outside allowed directories
	// This would be configurable in production
	if strings.HasPrefix(resource, "/etc/") || strings.HasPrefix(resource, "/sys/") {
		return fmt.Errorf("access to system directories not allowed: %s", resource)
	}

	return nil
}

// SandboxValidator provides sandboxing validation.
type SandboxValidator struct {
	allowedPaths []string // Allowed path prefixes
	deniedPaths  []string // Denied path prefixes
}

// NewSandboxValidator creates a new sandbox validator.
func NewSandboxValidator() *SandboxValidator {
	return &SandboxValidator{
		allowedPaths: make([]string, 0),
		deniedPaths:  make([]string, 0),
	}
}

// AddAllowedPath adds an allowed path prefix.
func (sv *SandboxValidator) AddAllowedPath(path string) {
	sv.allowedPaths = append(sv.allowedPaths, path)
}

// AddDeniedPath adds a denied path prefix.
func (sv *SandboxValidator) AddDeniedPath(path string) {
	sv.deniedPaths = append(sv.deniedPaths, path)
}

// ValidatePath validates if a path is allowed in the sandbox.
func (sv *SandboxValidator) ValidatePath(path string) error {
	// Check denied paths first
	for _, denied := range sv.deniedPaths {
		if strings.HasPrefix(path, denied) {
			return fmt.Errorf("path denied by sandbox: %s", path)
		}
	}

	// If there are allowed paths, check them
	if len(sv.allowedPaths) > 0 {
		for _, allowed := range sv.allowedPaths {
			if strings.HasPrefix(path, allowed) {
				return nil
			}
		}
		return fmt.Errorf("path not in allowed sandbox paths: %s", path)
	}

	// No allowed paths configured means all paths allowed (except denied)
	return nil
}
