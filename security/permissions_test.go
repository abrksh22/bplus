package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPermissionManager tests the PermissionManager.
func TestPermissionManager(t *testing.T) {
	t.Run("YOLO mode", func(t *testing.T) {
		pm := NewPermissionManager(ModeYOLO, nil)

		req := &PermissionRequest{
			Permission: PermissionWrite,
			Resource:   "/tmp/test.txt",
			Operation:  "write file",
			Risk:       RiskHigh,
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, granted)

		// Check audit log
		log := pm.GetAuditLog()
		assert.Len(t, log, 1)
		assert.Equal(t, PermissionWrite, log[0].Permission)
		assert.True(t, log[0].Granted)
	})

	t.Run("Deny mode", func(t *testing.T) {
		pm := NewPermissionManager(ModeDeny, nil)

		req := &PermissionRequest{
			Permission: PermissionRead,
			Resource:   "/tmp/test.txt",
			Operation:  "read file",
			Risk:       RiskLow,
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, granted)

		// Check audit log
		log := pm.GetAuditLog()
		assert.Len(t, log, 1)
		assert.False(t, log[0].Granted)
	})

	t.Run("Auto-approve mode for low risk", func(t *testing.T) {
		pm := NewPermissionManager(ModeAutoApprove, nil)

		req := &PermissionRequest{
			Permission: PermissionRead,
			Resource:   "/tmp/test.txt",
			Operation:  "read file",
			Risk:       RiskLow,
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, granted)
	})

	t.Run("Interactive mode with handler", func(t *testing.T) {
		handler := func(ctx context.Context, req *PermissionRequest) (bool, error) {
			// Simulate user approving
			return true, nil
		}

		pm := NewPermissionManager(ModeInteractive, handler)

		req := &PermissionRequest{
			Permission: PermissionWrite,
			Resource:   "/tmp/test.txt",
			Operation:  "write file",
			Risk:       RiskMedium,
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, granted)

		// Second request should be auto-granted (already granted)
		granted2, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, granted2)
	})

	t.Run("Interactive mode denying", func(t *testing.T) {
		handler := func(ctx context.Context, req *PermissionRequest) (bool, error) {
			// Simulate user denying
			return false, nil
		}

		pm := NewPermissionManager(ModeInteractive, handler)

		req := &PermissionRequest{
			Permission: PermissionExecute,
			Resource:   "rm -rf /",
			Operation:  "execute command",
			Risk:       RiskHigh,
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, granted)

		log := pm.GetAuditLog()
		assert.Len(t, log, 1)
		assert.False(t, log[0].Granted)
	})

	t.Run("Grant and revoke", func(t *testing.T) {
		pm := NewPermissionManager(ModeInteractive, nil)

		// Grant explicitly
		pm.Grant(PermissionRead)

		req := &PermissionRequest{
			Permission: PermissionRead,
			Resource:   "/tmp/test.txt",
			Operation:  "read file",
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, granted)

		// Revoke
		pm.Revoke(PermissionRead)

		granted2, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, granted2)
	})

	t.Run("Grant all", func(t *testing.T) {
		pm := NewPermissionManager(ModeInteractive, nil)
		pm.GrantAll()

		req := &PermissionRequest{
			Permission: PermissionExecute,
			Resource:   "ls",
			Operation:  "execute",
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, granted)
	})

	t.Run("Revoke all", func(t *testing.T) {
		pm := NewPermissionManager(ModeInteractive, nil)
		pm.GrantAll()
		pm.RevokeAll()

		req := &PermissionRequest{
			Permission: PermissionRead,
			Resource:   "/tmp/test.txt",
			Operation:  "read",
		}

		granted, err := pm.Check(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, granted)
	})
}

// TestRiskAssessment tests risk assessment.
func TestRiskAssessment(t *testing.T) {
	tests := []struct {
		operation    string
		resource     string
		expectedRisk RiskLevel
	}{
		{"read file", "/tmp/test.txt", RiskLow},
		{"rm -rf /", "/", RiskHigh},
		{"delete file", "/tmp/test.txt", RiskHigh},
		{"sudo command", "/bin/bash", RiskHigh},
		{"write file", "/tmp/test.txt", RiskMedium},
		{"execute command", "ls", RiskMedium},
		{"chmod 777", "/tmp/file", RiskHigh},
		{"format drive", "/dev/sda", RiskHigh},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			risk := AssessRisk(tt.operation, tt.resource)
			assert.Equal(t, tt.expectedRisk, risk)
		})
	}
}

// TestValidateResource tests resource validation.
func TestValidateResource(t *testing.T) {
	tests := []struct {
		resource string
		valid    bool
	}{
		{"/tmp/test.txt", true},
		{"/home/user/file.txt", true},
		{"../../../etc/passwd", false},
		{"/etc/passwd", false},
		{"/sys/kernel/config", false},
		{"relative/path/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			err := ValidateResource(tt.resource)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestSandboxValidator tests sandbox validation.
func TestSandboxValidator(t *testing.T) {
	sv := NewSandboxValidator()

	t.Run("No restrictions", func(t *testing.T) {
		err := sv.ValidatePath("/tmp/test.txt")
		assert.NoError(t, err)
	})

	t.Run("Allowed paths", func(t *testing.T) {
		sv.AddAllowedPath("/home/user")
		sv.AddAllowedPath("/tmp")

		err := sv.ValidatePath("/home/user/file.txt")
		assert.NoError(t, err)

		err = sv.ValidatePath("/tmp/test.txt")
		assert.NoError(t, err)

		err = sv.ValidatePath("/etc/passwd")
		assert.Error(t, err)
	})

	t.Run("Denied paths", func(t *testing.T) {
		sv2 := NewSandboxValidator()
		sv2.AddDeniedPath("/etc")
		sv2.AddDeniedPath("/sys")

		err := sv2.ValidatePath("/tmp/test.txt")
		assert.NoError(t, err)

		err = sv2.ValidatePath("/etc/passwd")
		assert.Error(t, err)

		err = sv2.ValidatePath("/sys/kernel/config")
		assert.Error(t, err)
	})

	t.Run("Both allowed and denied", func(t *testing.T) {
		sv3 := NewSandboxValidator()
		sv3.AddAllowedPath("/home")
		sv3.AddDeniedPath("/home/user/.ssh")

		err := sv3.ValidatePath("/home/user/file.txt")
		assert.NoError(t, err)

		err = sv3.ValidatePath("/home/user/.ssh/id_rsa")
		assert.Error(t, err)
	})
}

// TestPermissionRequest tests permission request structure.
func TestPermissionRequest(t *testing.T) {
	req := &PermissionRequest{
		Permission:  PermissionWrite,
		Resource:    "/tmp/test.txt",
		Operation:   "write file",
		Reason:      "save user data",
		Risk:        RiskMedium,
		ToolName:    "write",
		RequestedAt: time.Now(),
	}

	assert.Equal(t, PermissionWrite, req.Permission)
	assert.Equal(t, "/tmp/test.txt", req.Resource)
	assert.Equal(t, RiskMedium, req.Risk)
	assert.Equal(t, "write", req.ToolName)
}

// TestAuditLog tests audit logging.
func TestAuditLog(t *testing.T) {
	pm := NewPermissionManager(ModeYOLO, nil)

	requests := []PermissionRequest{
		{Permission: PermissionRead, Resource: "file1.txt", Operation: "read"},
		{Permission: PermissionWrite, Resource: "file2.txt", Operation: "write"},
		{Permission: PermissionExecute, Resource: "ls", Operation: "execute"},
	}

	for _, req := range requests {
		pm.Check(context.Background(), &req)
	}

	log := pm.GetAuditLog()
	assert.Len(t, log, 3)

	// Verify log entries
	for i, entry := range log {
		assert.Equal(t, requests[i].Permission, entry.Permission)
		assert.Equal(t, requests[i].Resource, entry.Resource)
		assert.True(t, entry.Granted) // YOLO mode grants all
		assert.Equal(t, ModeYOLO, entry.Mode)
	}
}
