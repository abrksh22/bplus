package util

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"time"
)

// IsValidURL checks if a string is a valid URL
func IsValidURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// IsValidHTTPURL checks if a string is a valid HTTP/HTTPS URL
func IsValidHTTPURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// GetFreePort finds an available TCP port
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("failed to resolve TCP address: %w", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("failed to listen: %w", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// IsPortAvailable checks if a port is available
func IsPortAvailable(port int) bool {
	addr := net.JoinHostPort("localhost", fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

// GetOutboundIP gets the preferred outbound IP of this machine
func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("failed to get outbound IP: %w", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// IsReachable checks if a host is reachable
func IsReachable(host string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// WaitForPort waits for a port to become available
func WaitForPort(host string, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for port %d", port)
}

// HTTPGet performs a simple HTTP GET request with timeout
func HTTPGet(ctx context.Context, url string, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil
}

// IsHTTPStatusOK checks if an HTTP status code is in the 2xx range
func IsHTTPStatusOK(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// ParseHostPort parses host:port string
func ParseHostPort(hostPort string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", 0, fmt.Errorf("failed to split host:port: %w", err)
	}

	var port int
	_, err = fmt.Sscanf(portStr, "%d", &port)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %w", err)
	}

	return host, port, nil
}

// JoinHostPort joins host and port into host:port string
func JoinHostPort(host string, port int) string {
	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}

// RetryConfig configures retry behavior with exponential backoff
type RetryConfig struct {
	MaxAttempts  int           // Maximum number of retry attempts (default: 3)
	InitialDelay time.Duration // Initial delay between retries (default: 100ms)
	MaxDelay     time.Duration // Maximum delay between retries (default: 10s)
	Multiplier   float64       // Backoff multiplier (default: 2.0)
	Jitter       bool          // Add jitter to delays (default: true)
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryFunc is a function that may fail and should be retried
type RetryFunc func() error

// IsRetryable determines if an error should trigger a retry
type IsRetryable func(error) bool

// Retry executes a function with exponential backoff
func Retry(ctx context.Context, config RetryConfig, fn RetryFunc) error {
	return RetryWithCondition(ctx, config, fn, nil)
}

// RetryWithCondition executes a function with exponential backoff and custom retry condition
func RetryWithCondition(ctx context.Context, config RetryConfig, fn RetryFunc, isRetryable IsRetryable) error {
	// Validate and apply defaults
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3
	}
	if config.InitialDelay <= 0 {
		config.InitialDelay = 100 * time.Millisecond
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 10 * time.Second
	}
	if config.Multiplier <= 0 {
		config.Multiplier = 2.0
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check context cancellation before attempt
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if we should retry this error
		if isRetryable != nil && !isRetryable(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep after the last attempt
		if attempt >= config.MaxAttempts {
			break
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * config.Multiplier)
		if nextDelay > config.MaxDelay {
			nextDelay = config.MaxDelay
		}

		// Add jitter if enabled (±25% randomness)
		if config.Jitter {
			jitterRange := float64(nextDelay) * 0.25
			jitterOffset := (math.Float64frombits(uint64(time.Now().UnixNano())) - 0.5) * jitterRange
			nextDelay = time.Duration(float64(nextDelay) + jitterOffset)
		}

		// Wait before next attempt
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled after %d attempts: %w", attempt, ctx.Err())
		case <-time.After(delay):
		}

		delay = nextDelay
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// RetryWithBackoff is a convenience wrapper for Retry with custom backoff parameters
func RetryWithBackoff(ctx context.Context, maxAttempts int, initialDelay, maxDelay time.Duration, fn RetryFunc) error {
	config := RetryConfig{
		MaxAttempts:  maxAttempts,
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   2.0,
		Jitter:       true,
	}
	return Retry(ctx, config, fn)
}
