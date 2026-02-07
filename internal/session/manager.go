package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dhruvsoni1802/browser-query-ai/internal/cdp"
)

// Manager manages all active sessions and CDP connections
type Manager struct {
	sessions   map[string]*Session // sessionID → Pointer to Session Struct
	cdpClients map[int]*cdp.Client // Browser process port → Pointer to CDP Client Struct
	mu         sync.RWMutex        // Protects concurrent access
}

// NewManager creates a new session manager
func NewManager() *Manager {
	return &Manager{
		sessions:   make(map[string]*Session),
		cdpClients: make(map[int]*cdp.Client),
	}
}

// generateSessionID creates a unique session identifier
func generateSessionID() (string, error) {
	// Generate 16 random bytes
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)

	// If there is an error, return an error
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Encode the random bytes to base64
	sessionID := base64.URLEncoding.EncodeToString(randomBytes)

	// Return the session ID with prefix
	return "sess_" + sessionID, nil
}

// GetOrCreateCDPClient gets existing client or creates new one for a port
func (m *Manager) GetOrCreateCDPClient(port int) (*cdp.Client, error) {
	// Check if the client already exists for this port
	client, exists := m.cdpClients[port]
	if exists {
		return client, nil
	}

	// If the client does not exist, discover the WebSocket URL
	wsURL, err := cdp.GetWebSocketURL("localhost", strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("failed to discover WebSocket URL: %w", err)
	}

	// Create a new CDP client and connect to it
	client = cdp.NewClient(wsURL)
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to CDP client: %w", err)
	}

	// Add the client to the manager
	m.cdpClients[port] = client
	return client, nil
}

// CreateSession creates a new isolated browsing session
func (m *Manager) CreateSession(port int) (*Session, error) {
	// Acquire write lock to prevent concurrent access
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate a unique session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Get or create a CDP client for the given port
	client, err := m.GetOrCreateCDPClient(port)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create CDP client: %w", err)
	}

	// Create a browser context on the browser process
	contextID, err := client.CreateBrowserContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	// Create a new session struct
	session := &Session{
		ID:           sessionID,
		ProcessPort:  port,
		ContextID:    contextID,
		PageIDs:      []string{},
		CDPClient:    client,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Status:       SessionActive,
	}

	// Add the session to the manager
	m.sessions[sessionID] = session

	// Return the session
	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	// Acquire read lock (allows multiple concurrent reads)
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Look up session in map
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// DestroySession cleans up all resources for a session
func (m *Manager) DestroySession(sessionID string) error {
	// Acquire write lock (exclusive access)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get session from map
	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Close all pages in this session
	for _, pageID := range session.PageIDs {
		if err := session.CDPClient.CloseTarget(pageID); err != nil {
			// Log error but continue cleanup
			fmt.Printf("warning: failed to close page %s: %v\n", pageID, err)
		}
	}

	// Dispose the browser context
	if err := session.CDPClient.DisposeBrowserContext(session.ContextID); err != nil {
		return fmt.Errorf("failed to dispose browser context: %w", err)
	}

	// Mark session as closed
	session.Status = SessionClosed

	// Remove from map
	delete(m.sessions, sessionID)

	return nil
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() []*Session {
	// Acquire read lock
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create slice to hold sessions
	sessions := make([]*Session, 0, len(m.sessions))

	// Loop through sessions and append to slice
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetSessionCount returns the number of active sessions
func (m *Manager) GetSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// Close closes all CDP connections (cleanup on service shutdown)
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all CDP clients
	for port, client := range m.cdpClients {
		if err := client.Close(); err != nil {
			fmt.Printf("warning: failed to close CDP client on port %d: %v\n", port, err)
		}
	}

	// Clear maps
	m.sessions = make(map[string]*Session)
	m.cdpClients = make(map[int]*cdp.Client)

	return nil
}