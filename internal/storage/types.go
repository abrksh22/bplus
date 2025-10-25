package storage

import "time"

// Session represents a b+ session
type Session struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ContextSnapshot *string   `json:"context_snapshot,omitempty"`
	Metadata        *string   `json:"metadata,omitempty"`
}

// Message represents a conversation message
type Message struct {
	ID           int64     `json:"id"`
	SessionID    string    `json:"session_id"`
	Role         string    `json:"role"` // user, assistant, system, tool
	Content      string    `json:"content"`
	Timestamp    time.Time `json:"timestamp"`
	TokensInput  int       `json:"tokens_input"`
	TokensOutput int       `json:"tokens_output"`
	Cost         float64   `json:"cost"`
	Metadata     *string   `json:"metadata,omitempty"`
}

// File represents a file tracked in a session
type File struct {
	ID          int64     `json:"id"`
	SessionID   string    `json:"session_id"`
	Path        string    `json:"path"`
	ContentHash *string   `json:"content_hash,omitempty"`
	ModifiedAt  time.Time `json:"modified_at"`
	SizeBytes   int64     `json:"size_bytes"`
}

// Checkpoint represents a session checkpoint
type Checkpoint struct {
	ID            int64     `json:"id"`
	SessionID     string    `json:"session_id"`
	Name          *string   `json:"name,omitempty"`
	StateSnapshot string    `json:"state_snapshot"`
	CreatedAt     time.Time `json:"created_at"`
}

// Operation represents an operation for undo/redo
type Operation struct {
	ID         int64     `json:"id"`
	SessionID  string    `json:"session_id"`
	Type       string    `json:"type"`
	Details    *string   `json:"details,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Reversible bool      `json:"reversible"`
}

// Metric represents a performance or cost metric
type Metric struct {
	ID         int64     `json:"id"`
	SessionID  *string   `json:"session_id,omitempty"`
	MetricType string    `json:"metric_type"`
	MetricName *string   `json:"metric_name,omitempty"`
	Value      float64   `json:"value"`
	Timestamp  time.Time `json:"timestamp"`
	Metadata   *string   `json:"metadata,omitempty"`
}
