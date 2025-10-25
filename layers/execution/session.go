package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/abrksh22/bplus/internal/errors"
	"github.com/abrksh22/bplus/internal/logging"
	"github.com/abrksh22/bplus/internal/storage"
	"github.com/abrksh22/bplus/models"
)

// SessionManager manages agent sessions and persists conversation history.
type SessionManager struct {
	db     *storage.SQLiteDB
	logger *logging.Logger
}

// NewSessionManager creates a new session manager.
func NewSessionManager(db *storage.SQLiteDB) *SessionManager {
	return &SessionManager{
		db:     db,
		logger: logging.NewDefaultLogger().WithComponent("session_manager"),
	}
}

// Session represents an agent session.
type Session struct {
	ID              string
	Name            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Messages        []models.Message
	TotalTokens     int
	TotalCost       float64
	ContextSnapshot string
	Metadata        map[string]interface{}
}

// CreateSession creates a new session.
func (sm *SessionManager) CreateSession(ctx context.Context, name string) (*Session, error) {
	sessionID := generateSessionID()
	now := time.Now()

	session := &Session{
		ID:        sessionID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  make([]models.Message, 0),
		Metadata:  make(map[string]interface{}),
	}

	// Insert into database
	query := `
		INSERT INTO sessions (id, name, created_at, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?)
	`

	metadataJSON, err := json.Marshal(session.Metadata)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to marshal session metadata")
	}

	_, err = sm.db.DB().Exec(query, sessionID, name, now, now, string(metadataJSON))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to create session")
	}

	sm.logger.Info("Session created", "session_id", sessionID, "name", name)
	return session, nil
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT id, name, created_at, updated_at, context_snapshot, metadata
		FROM sessions
		WHERE id = ?
	`

	var session Session
	var contextSnapshot, metadataJSON *string

	err := sm.db.DB().QueryRow(query, sessionID).Scan(&session.ID, &session.Name, &session.CreatedAt, &session.UpdatedAt, &contextSnapshot, &metadataJSON)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeFileNotFound, "session not found")
	}

	if contextSnapshot != nil {
		session.ContextSnapshot = *contextSnapshot
	}

	if metadataJSON != nil {
		if err := json.Unmarshal([]byte(*metadataJSON), &session.Metadata); err != nil {
			sm.logger.Warn("Failed to unmarshal session metadata", "error", err)
			session.Metadata = make(map[string]interface{})
		}
	}

	// Load messages
	messages, err := sm.GetMessages(ctx, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to load session messages")
	}
	session.Messages = messages

	// Calculate totals
	for _, msg := range messages {
		// Parse metadata to get tokens and cost
		if msg.Name != "" {
			// This is a metadata field we'll use for JSON storage
			var msgMeta struct {
				TokensInput  int     `json:"tokens_input"`
				TokensOutput int     `json:"tokens_output"`
				Cost         float64 `json:"cost"`
			}
			if err := json.Unmarshal([]byte(msg.Name), &msgMeta); err == nil {
				session.TotalTokens += msgMeta.TokensInput + msgMeta.TokensOutput
				session.TotalCost += msgMeta.Cost
			}
		}
	}

	return &session, nil
}

// SaveMessage saves a message to a session.
func (sm *SessionManager) SaveMessage(ctx context.Context, sessionID string, message models.Message, tokensInput, tokensOutput int, cost float64) error {
	query := `
		INSERT INTO messages (session_id, role, content, tokens_input, tokens_output, cost, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	// Store additional metadata
	metadata := make(map[string]interface{})
	if message.Name != "" {
		metadata["name"] = message.Name
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to marshal message metadata")
	}

	_, err = sm.db.DB().Exec(query, sessionID, message.Role, message.Content, tokensInput, tokensOutput, cost, string(metadataJSON))
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to save message")
	}

	// Update session updated_at
	updateQuery := `UPDATE sessions SET updated_at = ? WHERE id = ?`
	_, err = sm.db.DB().Exec(updateQuery, time.Now(), sessionID)
	if err != nil {
		sm.logger.Warn("Failed to update session timestamp", "error", err)
	}

	return nil
}

// GetMessages retrieves all messages for a session.
func (sm *SessionManager) GetMessages(ctx context.Context, sessionID string) ([]models.Message, error) {
	query := `
		SELECT role, content, metadata
		FROM messages
		WHERE session_id = ?
		ORDER BY id ASC
	`

	rows, err := sm.db.DB().Query(query, sessionID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to query messages")
	}
	defer rows.Close()

	messages := make([]models.Message, 0)
	for rows.Next() {
		var msg models.Message
		var metadataJSON *string

		if err := rows.Scan(&msg.Role, &msg.Content, &metadataJSON); err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to scan message row")
		}

		// Parse metadata
		if metadataJSON != nil {
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(*metadataJSON), &metadata); err == nil {
				if name, ok := metadata["name"].(string); ok {
					msg.Name = name
				}
			}
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "error iterating message rows")
	}

	return messages, nil
}

// ListSessions lists all sessions.
func (sm *SessionManager) ListSessions(ctx context.Context) ([]Session, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM sessions
		ORDER BY updated_at DESC
	`

	rows, err := sm.db.DB().Query(query)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to query sessions")
	}
	defer rows.Close()

	sessions := make([]Session, 0)
	for rows.Next() {
		var session Session
		if err := rows.Scan(&session.ID, &session.Name, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to scan session row")
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "error iterating session rows")
	}

	return sessions, nil
}

// DeleteSession deletes a session and all its messages.
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`

	_, err := sm.db.DB().Exec(query, sessionID)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to delete session")
	}

	sm.logger.Info("Session deleted", "session_id", sessionID)
	return nil
}

// UpdateSessionContext updates the context snapshot for a session.
func (sm *SessionManager) UpdateSessionContext(ctx context.Context, sessionID string, contextSnapshot string) error {
	query := `UPDATE sessions SET context_snapshot = ?, updated_at = ? WHERE id = ?`

	_, err := sm.db.DB().Exec(query, contextSnapshot, time.Now(), sessionID)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to update session context")
	}

	return nil
}

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
