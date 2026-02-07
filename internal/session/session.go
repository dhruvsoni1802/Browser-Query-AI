package session

import (
	"time"

	"github.com/dhruvsoni1802/browser-query-ai/internal/cdp"
)

// SessionStatus represents the current state of a session
type SessionStatus string

const (
	SessionActive  SessionStatus = "active"   // Session is running
	SessionClosed  SessionStatus = "closed"   // Session was explicitly closed
	SessionExpired SessionStatus = "expired"  // Session timed out
)

// Session represents an AI agent's isolated browsing session
type Session struct {
	ID           string          // Unique session identifier
	ProcessPort  int             // Which browser process (9222, 9223, etc.)
	ContextID    string          // CDP browser context ID
	PageIDs      []string        // List of page IDs in this context
	CDPClient    *cdp.Client     // WebSocket connection to browser
	CreatedAt    time.Time       // When session was created
	LastActivity time.Time       // Last time session was used
	Status       SessionStatus   // Current session status
}

// IsExpired checks if the session has been inactive too long
func (s *Session) IsExpired(timeout time.Duration) bool {
	return time.Since(s.LastActivity) > timeout
}

// UpdateActivity updates the last activity timestamp
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
}

// AddPage tracks a new page in this session
func (s *Session) AddPage(pageID string) {
	s.PageIDs = append(s.PageIDs, pageID)
	s.UpdateActivity()
}

// RemovePage removes a page from tracking
func (s *Session) RemovePage(pageID string) {
	for i, id := range s.PageIDs {
		if id == pageID {
			// Remove by swapping with last element and truncating
			s.PageIDs[i] = s.PageIDs[len(s.PageIDs)-1]
			s.PageIDs = s.PageIDs[:len(s.PageIDs)-1]
			break
		}
	}
	s.UpdateActivity()
}
