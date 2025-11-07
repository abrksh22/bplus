package context

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
	"github.com/abrksh22/bplus/internal/storage"
	"github.com/abrksh22/bplus/models"
)

// CheckpointManager manages checkpoints
type CheckpointManager struct {
	db     *storage.SQLiteDB
	logger *logging.Logger
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(db *storage.SQLiteDB) *CheckpointManager {
	return &CheckpointManager{
		db:     db,
		logger: logging.NewDefaultLogger().WithComponent("checkpoint_manager"),
	}
}

// CreateCheckpoint creates a new checkpoint
func (cm *CheckpointManager) CreateCheckpoint(ctx context.Context, sessionID, name, description string, snapshot *ContextSnapshot, messages []models.Message) (*Checkpoint, error) {
	checkpointID := generateCheckpointID()
	now := time.Now()

	// Serialize snapshot
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to marshal snapshot")
	}

	// Serialize messages
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to marshal messages")
	}

	checkpoint := &Checkpoint{
		ID:          checkpointID,
		SessionID:   sessionID,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		Context:     *snapshot,
		Messages:    messages,
		FileStates:  make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}

	// Insert into database
	query := `
		INSERT INTO checkpoints (id, session_id, name, description, created_at, state_snapshot, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	metadata := map[string]interface{}{
		"messages":      string(messagesJSON),
		"context_tokens": snapshot.TotalTokens,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to marshal checkpoint metadata")
	}

	_, err = cm.db.DB().Exec(query, checkpointID, sessionID, name, description, now, string(snapshotJSON), string(metadataJSON))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to create checkpoint")
	}

	cm.logger.Info("Checkpoint created",
		"checkpoint_id", checkpointID,
		"session_id", sessionID,
		"name", name,
		"tokens", snapshot.TotalTokens)

	return checkpoint, nil
}

// GetCheckpoint retrieves a checkpoint by ID
func (cm *CheckpointManager) GetCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	query := `
		SELECT id, session_id, name, description, created_at, state_snapshot, metadata
		FROM checkpoints
		WHERE id = ?
	`

	var checkpoint Checkpoint
	var snapshotJSON, metadataJSON string

	err := cm.db.DB().QueryRow(query, checkpointID).Scan(
		&checkpoint.ID,
		&checkpoint.SessionID,
		&checkpoint.Name,
		&checkpoint.Description,
		&checkpoint.CreatedAt,
		&snapshotJSON,
		&metadataJSON,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrCodeFileNotFound, "checkpoint not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to get checkpoint")
	}

	// Deserialize snapshot
	if err := json.Unmarshal([]byte(snapshotJSON), &checkpoint.Context); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to unmarshal snapshot")
	}

	// Deserialize metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		cm.logger.Warn("Failed to unmarshal checkpoint metadata", "error", err)
	} else {
		checkpoint.Metadata = metadata

		// Extract messages
		if messagesJSON, ok := metadata["messages"].(string); ok {
			if err := json.Unmarshal([]byte(messagesJSON), &checkpoint.Messages); err != nil {
				cm.logger.Warn("Failed to unmarshal checkpoint messages", "error", err)
			}
		}
	}

	return &checkpoint, nil
}

// ListCheckpoints lists all checkpoints for a session
func (cm *CheckpointManager) ListCheckpoints(ctx context.Context, sessionID string) ([]Checkpoint, error) {
	query := `
		SELECT id, session_id, name, description, created_at
		FROM checkpoints
		WHERE session_id = ?
		ORDER BY created_at DESC
	`

	rows, err := cm.db.DB().Query(query, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to query checkpoints")
	}
	defer rows.Close()

	checkpoints := make([]Checkpoint, 0)
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.SessionID, &cp.Name, &cp.Description, &cp.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to scan checkpoint row")
		}
		checkpoints = append(checkpoints, cp)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "error iterating checkpoint rows")
	}

	return checkpoints, nil
}

// RestoreCheckpoint restores a session from a checkpoint
func (cm *CheckpointManager) RestoreCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error) {
	checkpoint, err := cm.GetCheckpoint(ctx, checkpointID)
	if err != nil {
		return nil, err
	}

	cm.logger.Info("Checkpoint restored",
		"checkpoint_id", checkpointID,
		"session_id", checkpoint.SessionID,
		"name", checkpoint.Name)

	return checkpoint, nil
}

// DeleteCheckpoint deletes a checkpoint
func (cm *CheckpointManager) DeleteCheckpoint(ctx context.Context, checkpointID string) error {
	query := `DELETE FROM checkpoints WHERE id = ?`

	result, err := cm.db.DB().Exec(query, checkpointID)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to delete checkpoint")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to get rows affected")
	}

	if rows == 0 {
		return errors.New(errors.ErrCodeFileNotFound, "checkpoint not found")
	}

	cm.logger.Info("Checkpoint deleted", "checkpoint_id", checkpointID)
	return nil
}

// CleanupOldCheckpoints removes checkpoints older than the retention period
func (cm *CheckpointManager) CleanupOldCheckpoints(ctx context.Context, sessionID string, keepLast int) error {
	if keepLast <= 0 {
		keepLast = 10 // Default to keeping last 10
	}

	// Get all checkpoints for session
	checkpoints, err := cm.ListCheckpoints(ctx, sessionID)
	if err != nil {
		return err
	}

	// If we have fewer than keepLast, nothing to clean
	if len(checkpoints) <= keepLast {
		return nil
	}

	// Delete oldest checkpoints beyond keepLast
	toDelete := checkpoints[keepLast:]
	deletedCount := 0

	for _, cp := range toDelete {
		if err := cm.DeleteCheckpoint(ctx, cp.ID); err != nil {
			cm.logger.Warn("Failed to delete old checkpoint", "error", err, "checkpoint_id", cp.ID)
			continue
		}
		deletedCount++
	}

	cm.logger.Info("Old checkpoints cleaned up",
		"session_id", sessionID,
		"deleted", deletedCount,
		"kept", keepLast)

	return nil
}

// CreateAutoCheckpoint creates an automatic checkpoint before destructive operations
func (cm *CheckpointManager) CreateAutoCheckpoint(ctx context.Context, sessionID, operation string, snapshot *ContextSnapshot, messages []models.Message) (*Checkpoint, error) {
	name := fmt.Sprintf("auto_%s_%d", operation, time.Now().Unix())
	description := fmt.Sprintf("Automatic checkpoint before %s", operation)

	checkpoint, err := cm.CreateCheckpoint(ctx, sessionID, name, description, snapshot, messages)
	if err != nil {
		return nil, err
	}

	// Auto-cleanup old auto-checkpoints (keep last 5)
	go func() {
		if err := cm.CleanupOldCheckpoints(context.Background(), sessionID, 5); err != nil {
			cm.logger.Warn("Auto-cleanup failed", "error", err)
		}
	}()

	return checkpoint, nil
}

// generateCheckpointID generates a unique checkpoint ID
func generateCheckpointID() string {
	return fmt.Sprintf("checkpoint_%d", time.Now().UnixNano())
}
