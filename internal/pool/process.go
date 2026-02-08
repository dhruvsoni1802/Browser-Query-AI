package pool

import (
	"sync/atomic"
	"time"

	"github.com/dhruvsoni1802/browser-query-ai/internal/browser"
)

// ManagedProcess wraps the actual browser process with session count and other metrics
type ManagedProcess struct {
	Process      *browser.Process // The actual browser process
	sessionCount int64            // Active session count
	startedAt    time.Time        // When process was started
	lastHealthy  time.Time        // Last successful health check
}

// ProcessMetrics contains metrics about a managed process
type ProcessMetrics struct {
	Port             int           `json:"port"`
	SessionCount     int64         `json:"session_count"`
	Uptime           time.Duration `json:"uptime"`
	LastHealthyCheck time.Time     `json:"last_healthy_check"`
}

// NewManagedProcess creates a new managed process
func NewManagedProcess(chromiumPath string) (*ManagedProcess, error) {
	// Create a new browser process
	process, err := browser.NewProcess(chromiumPath)
	if err != nil {
		return nil, err
	}

	// Start the browser process
	if err := process.Start(); err != nil {
		return nil, err
	}

	// Wait for the browser process to be ready
	time.Sleep(2 * time.Second)

	return &ManagedProcess{
		Process:      process,
		sessionCount: 0,
		startedAt:    time.Now(),
		lastHealthy:  time.Now(),
	}, nil
}

// GetSessionCount returns the current session count using atomic operations
func (mp *ManagedProcess) GetSessionCount() int64 {
	return atomic.LoadInt64(&mp.sessionCount)
}

// IncrementSessionCount increments the session count using atomic operations
func (mp *ManagedProcess) IncrementSessionCount() {
	atomic.AddInt64(&mp.sessionCount, 1)
}

// DecrementSessionCount decrements the session count using atomic operations
func (mp *ManagedProcess) DecrementSessionCount() {
	atomic.AddInt64(&mp.sessionCount, -1)
}

// GetPort returns the browser process port
func (mp *ManagedProcess) GetPort() int {
	return mp.Process.DebugPort
}

// IsHealthy checks if the browser process is still alive
func (mp *ManagedProcess) IsHealthy() bool {
	if mp.Process.IsAlive() {
		mp.lastHealthy = time.Now()
		return true
	}
	return false
}

// Stop stops the browser process
func (mp *ManagedProcess) Stop() error {
	return mp.Process.Stop()
}

// GetMetrics returns the process metrics
func (mp *ManagedProcess) GetMetrics() ProcessMetrics {
	return ProcessMetrics{
		Port:             mp.GetPort(),
		SessionCount:     atomic.LoadInt64(&mp.sessionCount),
		Uptime:           time.Since(mp.startedAt),
		LastHealthyCheck: mp.lastHealthy,
	}
}