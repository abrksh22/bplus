package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestNewBoltDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify database was created
	require.FileExists(t, dbPath)
}

func TestBoltDB_SetGet(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	t.Run("set and get string", func(t *testing.T) {
		err := db.Set(BucketCache, []byte("key1"), "value1")
		assert.NoError(t, err)

		var value string
		err = db.Get(BucketCache, []byte("key1"), &value)
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("set and get struct", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Count int
		}

		original := TestStruct{Name: "test", Count: 42}
		err := db.Set(BucketState, []byte("struct-key"), original)
		assert.NoError(t, err)

		var retrieved TestStruct
		err = db.Get(BucketState, []byte("struct-key"), &retrieved)
		assert.NoError(t, err)
		assert.Equal(t, original, retrieved)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		var value string
		err := db.Get(BucketCache, []byte("nonexistent"), &value)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key not found")
	})
}

func TestBoltDB_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Set a value
	err = db.Set(BucketCache, []byte("delete-key"), "delete-value")
	require.NoError(t, err)

	// Verify it exists
	var value string
	err = db.Get(BucketCache, []byte("delete-key"), &value)
	assert.NoError(t, err)

	// Delete it
	err = db.Delete(BucketCache, []byte("delete-key"))
	assert.NoError(t, err)

	// Verify it's gone
	err = db.Get(BucketCache, []byte("delete-key"), &value)
	assert.Error(t, err)
}

func TestBoltDB_List(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Add multiple keys
	for i := 0; i < 5; i++ {
		key := []byte("list-key-" + string(rune('0'+i)))
		err := db.Set(BucketCache, key, i)
		require.NoError(t, err)
	}

	// List all keys
	keys, err := db.List(BucketCache)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(keys), 5)
}

func TestBoltDB_ForEach(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Add test data
	testData := map[string]string{
		"foreach1": "value1",
		"foreach2": "value2",
		"foreach3": "value3",
	}

	for k, v := range testData {
		err := db.Set(BucketContext, []byte(k), v)
		require.NoError(t, err)
	}

	// Iterate over all items
	count := 0
	err = db.ForEach(BucketContext, func(k, v []byte) error {
		count++
		return nil
	})

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, len(testData))
}

func TestBoltDB_CreateDeleteBucket(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	customBucket := []byte("custom-bucket")

	t.Run("create bucket", func(t *testing.T) {
		err := db.CreateBucket(customBucket)
		assert.NoError(t, err)

		// Verify we can use it
		err = db.Set(customBucket, []byte("test"), "value")
		assert.NoError(t, err)
	})

	t.Run("delete bucket", func(t *testing.T) {
		err := db.DeleteBucket(customBucket)
		assert.NoError(t, err)

		// Verify it's gone
		err = db.Set(customBucket, []byte("test"), "value")
		assert.Error(t, err)
	})
}

func TestBoltDB_Cache(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	t.Run("set and get cache", func(t *testing.T) {
		err := db.SetCache([]byte("cache-key"), "cache-value", 5*time.Second)
		assert.NoError(t, err)

		var value string
		found, err := db.GetCache([]byte("cache-key"), &value)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, "cache-value", value)
	})

	t.Run("expired cache", func(t *testing.T) {
		// Set with very short TTL
		err := db.SetCache([]byte("expire-key"), "expire-value", 1*time.Millisecond)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		var value string
		found, err := db.GetCache([]byte("expire-key"), &value)
		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("non-existent cache key", func(t *testing.T) {
		var value string
		found, err := db.GetCache([]byte("nonexistent-cache"), &value)
		assert.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("clear expired cache", func(t *testing.T) {
		// Add expired and valid entries
		db.SetCache([]byte("valid"), "valid-value", 1*time.Hour)
		db.SetCache([]byte("expired1"), "expired-value", 1*time.Millisecond)
		db.SetCache([]byte("expired2"), "expired-value", 1*time.Millisecond)

		time.Sleep(10 * time.Millisecond)

		err := db.ClearExpiredCache()
		assert.NoError(t, err)

		// Valid entry should still exist
		var value string
		found, err := db.GetCache([]byte("valid"), &value)
		assert.NoError(t, err)
		assert.True(t, found)
	})
}

func TestBoltDB_Sync(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Add some data
	err = db.Set(BucketState, []byte("sync-key"), "sync-value")
	require.NoError(t, err)

	// Sync to disk
	err = db.Sync()
	assert.NoError(t, err)
}

func TestBoltDB_Stats(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Add some data
	for i := 0; i < 10; i++ {
		err := db.Set(BucketCache, []byte("stats-"+string(rune('0'+i))), i)
		require.NoError(t, err)
	}

	stats := db.Stats()
	// Stats should exist (just verify we can call it)
	assert.NotNil(t, stats)
}

func TestBoltDB_Batch(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// Batch operation
	err = db.Batch(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(BucketCache)
		if bucket == nil {
			return assert.AnError
		}

		// Add multiple items in a batch
		for i := 0; i < 5; i++ {
			key := []byte("batch-" + string(rune('0'+i)))
			value := []byte("value-" + string(rune('0'+i)))
			if err := bucket.Put(key, value); err != nil {
				return err
			}
		}

		return nil
	})

	assert.NoError(t, err)

	// Verify data was written
	keys, err := db.List(BucketCache)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(keys), 5)
}

func BenchmarkBoltDB_Set(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(b, err)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte("bench-key")
		db.Set(BucketCache, key, i)
	}
}

func BenchmarkBoltDB_Get(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(b, err)
	defer db.Close()

	// Setup
	err = db.Set(BucketCache, []byte("bench-key"), "bench-value")
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var value string
		db.Get(BucketCache, []byte("bench-key"), &value)
	}
}

func BenchmarkBoltDB_SetCache(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.bolt")

	db, err := NewBoltDB(dbPath)
	require.NoError(b, err)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte("cache-key")
		db.SetCache(key, i, 1*time.Hour)
	}
}
