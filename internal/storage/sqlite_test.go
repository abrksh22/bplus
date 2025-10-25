package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLiteDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify database was created
	require.FileExists(t, dbPath)
}

func TestSQLiteDB_SessionOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	t.Run("create session", func(t *testing.T) {
		err := db.CreateSession("test-session-1", "Test Session 1")
		assert.NoError(t, err)
	})

	t.Run("get session", func(t *testing.T) {
		session, err := db.GetSession("test-session-1")
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "test-session-1", session.ID)
		assert.Equal(t, "Test Session 1", session.Name)
	})

	t.Run("update session", func(t *testing.T) {
		session, err := db.GetSession("test-session-1")
		require.NoError(t, err)

		contextSnap := `{"key": "value"}`
		session.ContextSnapshot = &contextSnap
		session.Name = "Updated Session"

		err = db.UpdateSession(session)
		assert.NoError(t, err)

		// Verify update
		updated, err := db.GetSession("test-session-1")
		assert.NoError(t, err)
		assert.Equal(t, "Updated Session", updated.Name)
		assert.NotNil(t, updated.ContextSnapshot)
	})

	t.Run("list sessions", func(t *testing.T) {
		// Create another session
		err := db.CreateSession("test-session-2", "Test Session 2")
		require.NoError(t, err)

		sessions, err := db.ListSessions()
		assert.NoError(t, err)
		assert.Len(t, sessions, 2)
	})

	t.Run("delete session", func(t *testing.T) {
		err := db.DeleteSession("test-session-1")
		assert.NoError(t, err)

		// Verify deletion
		_, err = db.GetSession("test-session-1")
		assert.Error(t, err)
	})

	t.Run("get non-existent session", func(t *testing.T) {
		_, err := db.GetSession("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}

func TestSQLiteDB_MessageOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create a session first
	err = db.CreateSession("msg-session", "Message Session")
	require.NoError(t, err)

	t.Run("add message", func(t *testing.T) {
		msg := &Message{
			SessionID:    "msg-session",
			Role:         "user",
			Content:      "Hello, world!",
			TokensInput:  10,
			TokensOutput: 0,
			Cost:         0.001,
		}

		err := db.AddMessage(msg)
		assert.NoError(t, err)
		assert.NotZero(t, msg.ID)
	})

	t.Run("get messages", func(t *testing.T) {
		// Add more messages
		for i := 0; i < 5; i++ {
			msg := &Message{
				SessionID: "msg-session",
				Role:      "assistant",
				Content:   "Response message",
			}
			err := db.AddMessage(msg)
			require.NoError(t, err)
		}

		messages, err := db.GetMessages("msg-session", 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(messages), 5)
	})

	t.Run("get messages with limit", func(t *testing.T) {
		messages, err := db.GetMessages("msg-session", 3)
		assert.NoError(t, err)
		assert.Len(t, messages, 3)
	})

	t.Run("search messages", func(t *testing.T) {
		// Add a searchable message
		msg := &Message{
			SessionID: "msg-session",
			Role:      "user",
			Content:   "This is a unique searchable message about database testing",
		}
		err := db.AddMessage(msg)
		require.NoError(t, err)

		// Search for it
		results, err := db.SearchMessages("database")
		assert.NoError(t, err)
		assert.NotEmpty(t, results)

		// Verify result contains our message
		found := false
		for _, r := range results {
			if r.Content == msg.Content {
				found = true
				break
			}
		}
		assert.True(t, found, "Search should find the message")
	})
}

func TestSQLiteDB_CheckpointOperations(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create a session
	err = db.CreateSession("cp-session", "Checkpoint Session")
	require.NoError(t, err)

	t.Run("create checkpoint", func(t *testing.T) {
		name := "Checkpoint 1"
		cp := &Checkpoint{
			SessionID:     "cp-session",
			Name:          &name,
			StateSnapshot: `{"state": "data"}`,
		}

		err := db.CreateCheckpoint(cp)
		assert.NoError(t, err)
		assert.NotZero(t, cp.ID)
	})

	t.Run("get checkpoints", func(t *testing.T) {
		// Create more checkpoints
		for i := 0; i < 3; i++ {
			cp := &Checkpoint{
				SessionID:     "cp-session",
				StateSnapshot: `{"state": "data"}`,
			}
			err := db.CreateCheckpoint(cp)
			require.NoError(t, err)
		}

		checkpoints, err := db.GetCheckpoints("cp-session")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(checkpoints), 3)
	})
}

func TestSQLiteDB_Backup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	backupPath := filepath.Join(tmpDir, "backup.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create some data
	err = db.CreateSession("backup-session", "Backup Session")
	require.NoError(t, err)

	// Backup
	err = db.Backup(backupPath)
	assert.NoError(t, err)
	assert.FileExists(t, backupPath)

	// Verify backup by opening it
	backupDB, err := NewSQLiteDB(backupPath)
	require.NoError(t, err)
	defer backupDB.Close()

	// Verify data exists in backup
	session, err := backupDB.GetSession("backup-session")
	assert.NoError(t, err)
	assert.Equal(t, "Backup Session", session.Name)
}

func TestSQLiteDB_ForeignKeyConstraints(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Create session
	err = db.CreateSession("fk-session", "FK Session")
	require.NoError(t, err)

	// Add message
	msg := &Message{
		SessionID: "fk-session",
		Role:      "user",
		Content:   "Test message",
	}
	err = db.AddMessage(msg)
	require.NoError(t, err)

	// Delete session (should cascade delete messages)
	err = db.DeleteSession("fk-session")
	assert.NoError(t, err)

	// Verify messages were deleted
	messages, err := db.GetMessages("fk-session", 0)
	assert.NoError(t, err)
	assert.Empty(t, messages)
}

func BenchmarkSQLiteDB_AddMessage(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(b, err)
	defer db.Close()

	err = db.CreateSession("bench-session", "Benchmark Session")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := &Message{
			SessionID: "bench-session",
			Role:      "user",
			Content:   "Benchmark message",
		}
		db.AddMessage(msg)
	}
}

func BenchmarkSQLiteDB_GetMessages(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")

	db, err := NewSQLiteDB(dbPath)
	require.NoError(b, err)
	defer db.Close()

	err = db.CreateSession("bench-session", "Benchmark Session")
	require.NoError(b, err)

	// Add 100 messages
	for i := 0; i < 100; i++ {
		msg := &Message{
			SessionID: "bench-session",
			Role:      "user",
			Content:   "Benchmark message",
		}
		db.AddMessage(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.GetMessages("bench-session", 50)
	}
}
