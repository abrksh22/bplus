package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

// BoltDB wraps a bbolt key-value store
type BoltDB struct {
	db   *bbolt.DB
	path string
}

// Common bucket names
var (
	BucketCache    = []byte("cache")
	BucketContext  = []byte("context")
	BucketSessions = []byte("sessions")
	BucketState    = []byte("state")
)

// NewBoltDB creates a new BoltDB instance
func NewBoltDB(path string) (*BoltDB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := bbolt.Open(path, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt database: %w", err)
	}

	bolt := &BoltDB{
		db:   db,
		path: path,
	}

	// Initialize buckets
	if err := bolt.initBuckets(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize buckets: %w", err)
	}

	return bolt, nil
}

// initBuckets creates default buckets
func (b *BoltDB) initBuckets() error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		buckets := [][]byte{
			BucketCache,
			BucketContext,
			BucketSessions,
			BucketState,
		}

		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}

		return nil
	})
}

// Set stores a value in the specified bucket
func (b *BoltDB) Set(bucket, key []byte, value interface{}) error {
	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}

		if err := b.Put(key, data); err != nil {
			return fmt.Errorf("failed to put value: %w", err)
		}

		return nil
	})
}

// Get retrieves a value from the specified bucket
func (b *BoltDB) Get(bucket, key []byte, dest interface{}) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}

		data := b.Get(key)
		if data == nil {
			return fmt.Errorf("key not found: %s", key)
		}

		if err := json.Unmarshal(data, dest); err != nil {
			return fmt.Errorf("failed to unmarshal value: %w", err)
		}

		return nil
	})
}

// Delete removes a key from the specified bucket
func (b *BoltDB) Delete(bucket, key []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}

		if err := b.Delete(key); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}

		return nil
	})
}

// List returns all keys in the specified bucket
func (b *BoltDB) List(bucket []byte) ([][]byte, error) {
	var keys [][]byte

	err := b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}

		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			// Make a copy of the key
			key := make([]byte, len(k))
			copy(key, k)
			keys = append(keys, key)
		}

		return nil
	})

	return keys, err
}

// ForEach iterates over all key-value pairs in the bucket
func (b *BoltDB) ForEach(bucket []byte, fn func(k, v []byte) error) error {
	return b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket not found: %s", bucket)
		}

		return b.ForEach(fn)
	})
}

// CreateBucket creates a new bucket
func (b *BoltDB) CreateBucket(name []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(name)
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	})
}

// DeleteBucket deletes a bucket
func (b *BoltDB) DeleteBucket(name []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(name); err != nil {
			return fmt.Errorf("failed to delete bucket: %w", err)
		}
		return nil
	})
}

// Batch performs multiple operations in a single transaction
func (b *BoltDB) Batch(fn func(tx *bbolt.Tx) error) error {
	return b.db.Batch(fn)
}

// Update performs a writable transaction
func (b *BoltDB) Update(fn func(tx *bbolt.Tx) error) error {
	return b.db.Update(fn)
}

// View performs a read-only transaction
func (b *BoltDB) View(fn func(tx *bbolt.Tx) error) error {
	return b.db.View(fn)
}

// Close closes the database
func (b *BoltDB) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// Sync syncs the database to disk
func (b *BoltDB) Sync() error {
	return b.db.Sync()
}

// Stats returns database statistics
func (b *BoltDB) Stats() bbolt.Stats {
	return b.db.Stats()
}

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// SetCache stores a value in the cache with expiration
func (b *BoltDB) SetCache(key []byte, value interface{}, ttl time.Duration) error {
	entry := CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	return b.Set(BucketCache, key, entry)
}

// GetCache retrieves a value from the cache
func (b *BoltDB) GetCache(key []byte, dest interface{}) (bool, error) {
	var entry CacheEntry
	err := b.Get(BucketCache, key, &entry)

	if err != nil {
		if err.Error() == fmt.Sprintf("key not found: %s", key) {
			return false, nil
		}
		return false, err
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Delete expired entry
		b.Delete(BucketCache, key)
		return false, nil
	}

	// Unmarshal the cached value
	data, err := json.Marshal(entry.Value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cached value: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return true, nil
}

// ClearExpiredCache removes all expired cache entries
func (b *BoltDB) ClearExpiredCache() error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(BucketCache)
		if bucket == nil {
			return nil
		}

		c := bucket.Cursor()
		now := time.Now()
		keysToDelete := [][]byte{}

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var entry CacheEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				continue
			}

			if now.After(entry.ExpiresAt) {
				// Make a copy of the key
				key := make([]byte, len(k))
				copy(key, k)
				keysToDelete = append(keysToDelete, key)
			}
		}

		// Delete expired keys
		for _, key := range keysToDelete {
			bucket.Delete(key)
		}

		return nil
	})
}
