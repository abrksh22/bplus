package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

// SQLiteDB wraps a SQLite database connection
type SQLiteDB struct {
	db   *sql.DB
	path string
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(path string) (*SQLiteDB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	sqlite := &SQLiteDB{
		db:   db,
		path: path,
	}

	// Initialize schema
	if err := sqlite.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return sqlite, nil
}

// initSchema creates all required tables
func (s *SQLiteDB) initSchema() error {
	schema := `
	-- Schema version tracking
	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Sessions table
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		context_snapshot TEXT,
		metadata TEXT -- JSON
	);

	-- Messages table
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL, -- 'user', 'assistant', 'system', 'tool'
		content TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		tokens_input INTEGER DEFAULT 0,
		tokens_output INTEGER DEFAULT 0,
		cost REAL DEFAULT 0.0,
		metadata TEXT, -- JSON
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	-- Files table (tracks files in session context)
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		path TEXT NOT NULL,
		content_hash TEXT,
		modified_at TIMESTAMP,
		size_bytes INTEGER,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
		UNIQUE(session_id, path)
	);

	-- Checkpoints table
	CREATE TABLE IF NOT EXISTS checkpoints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		name TEXT,
		state_snapshot TEXT NOT NULL, -- JSON
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	-- Operations table (for undo/redo)
	CREATE TABLE IF NOT EXISTS operations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		type TEXT NOT NULL, -- 'file_write', 'file_delete', 'command', etc.
		details TEXT, -- JSON
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		reversible BOOLEAN DEFAULT 1,
		FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
	);

	-- Metrics table
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT,
		metric_type TEXT NOT NULL, -- 'cost', 'tokens', 'duration', etc.
		metric_name TEXT,
		value REAL NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT -- JSON
	);

	-- Create indexes for common queries
	CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_files_session ON files(session_id);
	CREATE INDEX IF NOT EXISTS idx_operations_session ON operations(session_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_session ON metrics(session_id, timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(metric_type, timestamp);

	-- Create FTS5 virtual table for full-text search
	CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
		session_id UNINDEXED,
		role UNINDEXED,
		content,
		content=messages,
		content_rowid=id
	);

	-- Triggers to keep FTS table in sync
	CREATE TRIGGER IF NOT EXISTS messages_fts_insert AFTER INSERT ON messages BEGIN
		INSERT INTO messages_fts(rowid, session_id, role, content)
		VALUES (new.id, new.session_id, new.role, new.content);
	END;

	CREATE TRIGGER IF NOT EXISTS messages_fts_delete AFTER DELETE ON messages BEGIN
		DELETE FROM messages_fts WHERE rowid = old.id;
	END;

	CREATE TRIGGER IF NOT EXISTS messages_fts_update AFTER UPDATE ON messages BEGIN
		DELETE FROM messages_fts WHERE rowid = old.id;
		INSERT INTO messages_fts(rowid, session_id, role, content)
		VALUES (new.id, new.session_id, new.role, new.content);
	END;
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Check and update schema version
	return s.updateSchemaVersion(1)
}

// updateSchemaVersion records the schema version
func (s *SQLiteDB) updateSchemaVersion(version int) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO schema_version (version) VALUES (?)",
		version,
	)
	return err
}

// Session operations

// CreateSession creates a new session
func (s *SQLiteDB) CreateSession(id, name string) error {
	_, err := s.db.Exec(
		"INSERT INTO sessions (id, name) VALUES (?, ?)",
		id, name,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSession retrieves a session by ID
func (s *SQLiteDB) GetSession(id string) (*Session, error) {
	var session Session
	err := s.db.QueryRow(
		"SELECT id, name, created_at, updated_at, context_snapshot, metadata FROM sessions WHERE id = ?",
		id,
	).Scan(&session.ID, &session.Name, &session.CreatedAt, &session.UpdatedAt,
		&session.ContextSnapshot, &session.Metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// UpdateSession updates a session
func (s *SQLiteDB) UpdateSession(session *Session) error {
	_, err := s.db.Exec(
		"UPDATE sessions SET name = ?, updated_at = ?, context_snapshot = ?, metadata = ? WHERE id = ?",
		session.Name, time.Now(), session.ContextSnapshot, session.Metadata, session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// DeleteSession deletes a session and all related data
func (s *SQLiteDB) DeleteSession(id string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// ListSessions returns all sessions
func (s *SQLiteDB) ListSessions() ([]*Session, error) {
	rows, err := s.db.Query(
		"SELECT id, name, created_at, updated_at FROM sessions ORDER BY updated_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		if err := rows.Scan(&session.ID, &session.Name, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}

// Message operations

// AddMessage adds a message to a session
func (s *SQLiteDB) AddMessage(msg *Message) error {
	result, err := s.db.Exec(
		"INSERT INTO messages (session_id, role, content, tokens_input, tokens_output, cost, metadata) VALUES (?, ?, ?, ?, ?, ?, ?)",
		msg.SessionID, msg.Role, msg.Content, msg.TokensInput, msg.TokensOutput, msg.Cost, msg.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to add message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get message ID: %w", err)
	}
	msg.ID = id

	return nil
}

// GetMessages retrieves all messages for a session
func (s *SQLiteDB) GetMessages(sessionID string, limit int) ([]*Message, error) {
	query := "SELECT id, session_id, role, content, timestamp, tokens_input, tokens_output, cost, metadata FROM messages WHERE session_id = ? ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content,
			&msg.Timestamp, &msg.TokensInput, &msg.TokensOutput, &msg.Cost, &msg.Metadata); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	return messages, rows.Err()
}

// SearchMessages performs full-text search on messages
func (s *SQLiteDB) SearchMessages(query string) ([]*Message, error) {
	rows, err := s.db.Query(`
		SELECT m.id, m.session_id, m.role, m.content, m.timestamp, m.tokens_input, m.tokens_output, m.cost, m.metadata
		FROM messages m
		JOIN messages_fts fts ON m.id = fts.rowid
		WHERE messages_fts MATCH ?
		ORDER BY m.timestamp DESC
		LIMIT 100
	`, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content,
			&msg.Timestamp, &msg.TokensInput, &msg.TokensOutput, &msg.Cost, &msg.Metadata); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	return messages, rows.Err()
}

// Checkpoint operations

// CreateCheckpoint creates a checkpoint for a session
func (s *SQLiteDB) CreateCheckpoint(checkpoint *Checkpoint) error {
	result, err := s.db.Exec(
		"INSERT INTO checkpoints (session_id, name, state_snapshot) VALUES (?, ?, ?)",
		checkpoint.SessionID, checkpoint.Name, checkpoint.StateSnapshot,
	)
	if err != nil {
		return fmt.Errorf("failed to create checkpoint: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get checkpoint ID: %w", err)
	}
	checkpoint.ID = id

	return nil
}

// GetCheckpoints retrieves all checkpoints for a session
func (s *SQLiteDB) GetCheckpoints(sessionID string) ([]*Checkpoint, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, name, state_snapshot, created_at FROM checkpoints WHERE session_id = ? ORDER BY created_at DESC",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoints: %w", err)
	}
	defer rows.Close()

	var checkpoints []*Checkpoint
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.SessionID, &cp.Name, &cp.StateSnapshot, &cp.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan checkpoint: %w", err)
		}
		checkpoints = append(checkpoints, &cp)
	}

	return checkpoints, rows.Err()
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB returns the underlying sql.DB for advanced queries
func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

// Backup creates a backup of the database
func (s *SQLiteDB) Backup(destPath string) error {
	// Ensure destination directory exists
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Use VACUUM INTO for backup (SQLite 3.27.0+)
	_, err := s.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", destPath))
	if err != nil {
		return fmt.Errorf("failed to backup database: %w", err)
	}

	return nil
}

// Restore restores the database from a backup
func (s *SQLiteDB) Restore(sourcePath string) error {
	// Close current connection
	if err := s.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	// Copy backup to current database path
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if err := os.WriteFile(s.path, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	// Reopen database
	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		return fmt.Errorf("failed to reopen database: %w", err)
	}

	s.db = db
	return nil
}
