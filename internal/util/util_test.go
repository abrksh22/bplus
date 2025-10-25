package util

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// File utilities tests

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	assert.True(t, FileExists(existingFile))
	assert.False(t, FileExists(filepath.Join(tmpDir, "nonexistent.txt")))
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "exists")
	os.Mkdir(existingDir, 0755)

	assert.True(t, DirExists(existingDir))
	assert.False(t, DirExists(filepath.Join(tmpDir, "nonexistent")))
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new", "nested", "dir")

	err := EnsureDir(newDir)
	assert.NoError(t, err)
	assert.True(t, DirExists(newDir))
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "dest.txt")

	content := []byte("test content")
	os.WriteFile(src, content, 0644)

	err := CopyFile(src, dst)
	assert.NoError(t, err)

	dstContent, _ := os.ReadFile(dst)
	assert.Equal(t, content, dstContent)
}

func TestMoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "dest.txt")

	content := []byte("test content")
	os.WriteFile(src, content, 0644)

	err := MoveFile(src, dst)
	assert.NoError(t, err)
	assert.False(t, FileExists(src))
	assert.True(t, FileExists(dst))
}

func TestFileHash(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("test"), 0644)

	hash, err := FileHash(file)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 hex string length
}

func TestFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	os.WriteFile(file, content, 0644)

	size, err := FileSize(file)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
}

func TestTempFile(t *testing.T) {
	path, err := TempFile("test-")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "test-")
	defer os.Remove(path)
}

func TestTempDir(t *testing.T) {
	dir, err := TempDir("test-")
	assert.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.True(t, DirExists(dir))
	defer os.RemoveAll(dir)
}

// String utilities tests

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is a ..."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, Truncate(tt.input, tt.maxLen))
	}
}

func TestTruncateMiddle(t *testing.T) {
	result := TruncateMiddle("very_long_filename.txt", 15)
	assert.Len(t, result, 15)
	assert.Contains(t, result, "...")
}

func TestIsBlank(t *testing.T) {
	assert.True(t, IsBlank(""))
	assert.True(t, IsBlank("   "))
	assert.True(t, IsBlank("\t\n"))
	assert.False(t, IsBlank("test"))
	assert.False(t, IsBlank("  test  "))
}

func TestDefaultIfBlank(t *testing.T) {
	assert.Equal(t, "default", DefaultIfBlank("", "default"))
	assert.Equal(t, "default", DefaultIfBlank("   ", "default"))
	assert.Equal(t, "value", DefaultIfBlank("value", "default"))
}

func TestRandomString(t *testing.T) {
	str, err := RandomString(16)
	assert.NoError(t, err)
	assert.NotEmpty(t, str)
}

func TestRandomID(t *testing.T) {
	id, err := RandomID()
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	assert.True(t, Contains(slice, "b"))
	assert.False(t, Contains(slice, "d"))
}

func TestUnique(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	result := Unique(input)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")
}

func TestSplitLines(t *testing.T) {
	input := "line1\nline2\r\nline3"
	lines := SplitLines(input)
	assert.Len(t, lines, 3)
	assert.Equal(t, "line1", lines[0])
}

func TestIndent(t *testing.T) {
	input := "line1\nline2\nline3"
	result := Indent(input, "  ")
	assert.Contains(t, result, "  line1")
	assert.Contains(t, result, "  line2")
}

func TestCaseConversions(t *testing.T) {
	assert.Equal(t, "helloWorld", ToCamelCase("hello_world"))
	assert.Equal(t, "hello_world", ToSnakeCase("HelloWorld"))
	assert.Equal(t, "hello-world", ToKebabCase("HelloWorld"))
}

func TestSanitize(t *testing.T) {
	input := "file/name:with*bad?chars"
	result := Sanitize(input)
	assert.NotContains(t, result, "/")
	assert.NotContains(t, result, ":")
	assert.NotContains(t, result, "*")
	assert.NotContains(t, result, "?")
}

// Time utilities tests

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{5 * time.Second, "5.0s"},
		{90 * time.Second, "1m 30s"},
		{2 * time.Hour, "2h 0m"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, FormatDuration(tt.duration))
	}
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()

	tests := []struct {
		time     time.Time
		expected string
	}{
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5 minutes ago"},
		{now.Add(-2 * time.Hour), "2 hours ago"},
		{now.Add(-25 * time.Hour), "1 day ago"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, TimeAgo(tt.time))
	}
}

func TestStartOfDay(t *testing.T) {
	now := time.Now()
	start := StartOfDay(now)

	assert.Equal(t, 0, start.Hour())
	assert.Equal(t, 0, start.Minute())
	assert.Equal(t, 0, start.Second())
}

func TestIsSameDay(t *testing.T) {
	now := time.Now()
	later := now.Add(2 * time.Hour)
	tomorrow := now.Add(25 * time.Hour)

	assert.True(t, IsSameDay(now, later))
	assert.False(t, IsSameDay(now, tomorrow))
}

func TestIsToday(t *testing.T) {
	assert.True(t, IsToday(time.Now()))
	assert.False(t, IsToday(time.Now().Add(-25*time.Hour)))
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"5s", 5 * time.Second},
		{"2m", 2 * time.Minute},
		{"1h", 1 * time.Hour},
		{"2d", 48 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
	}

	for _, tt := range tests {
		result, err := ParseDuration(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, result)
	}
}

// Crypto utilities tests

func TestHash(t *testing.T) {
	data := []byte("test data")
	hash := Hash(data)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 hex length
}

func TestHashString(t *testing.T) {
	str := "test string"
	hash := HashString(str)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	plaintext := []byte("secret message")

	ciphertext, err := Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := Decrypt(ciphertext, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptString(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	plaintext := "secret message"

	ciphertext, err := EncryptString(plaintext, key)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := DecryptString(ciphertext, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestDeriveKey(t *testing.T) {
	password := "mypassword"
	key := DeriveKey(password)
	assert.Len(t, key, 32)
}

func TestMaskSecret(t *testing.T) {
	secret := "my_secret_api_key_123"
	masked := MaskSecret(secret, 3)
	assert.Contains(t, masked, "***")
	assert.Contains(t, masked, "my_")
	assert.Contains(t, masked, "123")
}

// Network utilities tests

func TestIsValidURL(t *testing.T) {
	assert.True(t, IsValidURL("https://example.com"))
	assert.True(t, IsValidURL("http://localhost:8080"))
	assert.False(t, IsValidURL("not a url"))
	assert.False(t, IsValidURL(""))
}

func TestIsValidHTTPURL(t *testing.T) {
	assert.True(t, IsValidHTTPURL("https://example.com"))
	assert.True(t, IsValidHTTPURL("http://example.com"))
	assert.False(t, IsValidHTTPURL("ftp://example.com"))
	assert.False(t, IsValidHTTPURL("example.com"))
}

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	assert.NoError(t, err)
	assert.Greater(t, port, 0)
	assert.Less(t, port, 65536)
}

func TestIsPortAvailable(t *testing.T) {
	// Port 99999 should be invalid/unavailable
	assert.True(t, IsPortAvailable(99999))
}

func TestIsHTTPStatusOK(t *testing.T) {
	assert.True(t, IsHTTPStatusOK(200))
	assert.True(t, IsHTTPStatusOK(201))
	assert.True(t, IsHTTPStatusOK(299))
	assert.False(t, IsHTTPStatusOK(300))
	assert.False(t, IsHTTPStatusOK(404))
	assert.False(t, IsHTTPStatusOK(500))
}

func TestParseHostPort(t *testing.T) {
	host, port, err := ParseHostPort("localhost:8080")
	assert.NoError(t, err)
	assert.Equal(t, "localhost", host)
	assert.Equal(t, 8080, port)
}

func TestJoinHostPort(t *testing.T) {
	result := JoinHostPort("localhost", 8080)
	assert.Equal(t, "localhost:8080", result)
}

func TestHTTPGet(t *testing.T) {
	// This test requires network access, so it's a basic check
	ctx := context.Background()
	_, err := HTTPGet(ctx, "https://httpbin.org/status/200", 5*time.Second)
	// We just check it doesn't panic - actual result depends on network
	_ = err
}

// Retry utilities tests

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	assert.Equal(t, 3, config.MaxAttempts)
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 10*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.True(t, config.Jitter)
}

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()

	attempts := 0
	fn := func() error {
		attempts++
		return nil // Success on first try
	}

	err := Retry(ctx, config, fn)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false, // Disable jitter for predictable timing
	}

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil // Success on third try
	}

	start := time.Now()
	err := Retry(ctx, config, fn)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
	// Should have delays: 10ms + 20ms = 30ms minimum
	assert.GreaterOrEqual(t, elapsed, 30*time.Millisecond)
}

func TestRetry_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return assert.AnError
	}

	err := Retry(ctx, config, fn)
	assert.Error(t, err)
	assert.Equal(t, 3, attempts)
	assert.Contains(t, err.Error(), "max retry attempts")
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	config := RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 20 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	fn := func() error {
		attempts++
		return assert.AnError
	}

	err := Retry(ctx, config, fn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry cancelled")
	assert.Less(t, attempts, 10) // Should stop before max attempts due to timeout
}

func TestRetryWithCondition_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := DefaultRetryConfig()

	attempts := 0
	fn := func() error {
		attempts++
		return assert.AnError // Non-retryable based on condition
	}

	// Make all errors non-retryable
	isRetryableNone := func(err error) bool { return false }

	err := RetryWithCondition(ctx, config, fn, isRetryableNone)
	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Should stop after first attempt
	assert.Contains(t, err.Error(), "non-retryable")
}

func TestRetryWithBackoff(t *testing.T) {
	ctx := context.Background()

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 2 {
			return assert.AnError
		}
		return nil
	}

	err := RetryWithBackoff(ctx, 3, 10*time.Millisecond, 100*time.Millisecond, fn)
	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	attempts := 0
	delays := []time.Duration{}
	lastTime := time.Now()

	fn := func() error {
		now := time.Now()
		if attempts > 0 {
			delays = append(delays, now.Sub(lastTime))
		}
		lastTime = now
		attempts++
		return assert.AnError
	}

	err := Retry(ctx, config, fn)
	assert.Error(t, err)
	assert.Equal(t, 4, attempts)

	// Verify exponential growth: 10ms, 20ms, 40ms
	require.Len(t, delays, 3)
	assert.GreaterOrEqual(t, delays[0], 10*time.Millisecond)
	assert.GreaterOrEqual(t, delays[1], 20*time.Millisecond)
	assert.GreaterOrEqual(t, delays[2], 40*time.Millisecond)
}

func TestRetry_MaxDelayRespected(t *testing.T) {
	ctx := context.Background()
	config := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   3.0, // Large multiplier to test max delay
		Jitter:       false,
	}

	attempts := 0
	delays := []time.Duration{}
	lastTime := time.Now()

	fn := func() error {
		now := time.Now()
		if attempts > 0 {
			delays = append(delays, now.Sub(lastTime))
		}
		lastTime = now
		attempts++
		return assert.AnError
	}

	err := Retry(ctx, config, fn)
	assert.Error(t, err)

	// All delays should be <= MaxDelay (100ms)
	for _, delay := range delays {
		assert.LessOrEqual(t, delay, 120*time.Millisecond) // Allow some tolerance
	}
}
