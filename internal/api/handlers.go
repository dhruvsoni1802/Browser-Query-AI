package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/dhruvsoni1802/browser-query-ai/internal/session"
	"github.com/go-chi/chi/v5"
)

// Handlers contains HTTP handlers for the API
type Handlers struct {
	sessionManager *session.Manager
	browserPort    int // For now, single browser port (later: load balancer decides)
}

// NewHandlers creates a new Handlers instance
func NewHandlers(manager *session.Manager, port int) *Handlers {
	return &Handlers{
		sessionManager: manager,
		browserPort:    port,
	}
}

// CreateSession handles POST /sessions
func (h *Handlers) CreateSession(w http.ResponseWriter, r *http.Request) {
	// Parse request body (optional port override)
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is OK, just use default port
		req.BrowserPort = h.browserPort
	}

	// Use provided port or default
	port := req.BrowserPort
	if port == 0 {
		port = h.browserPort
	}

	// Create session via session manager
	sess, err := h.sessionManager.CreateSession(port)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeSessionCreateFailed, err.Error())
		return
	}

	// Build response
	response := CreateSessionResponse{
		SessionID: sess.ID,
		ContextID: sess.ContextID,
		CreatedAt: sess.CreatedAt,
	}

	// Return 201 Created
	writeJSON(w, http.StatusCreated, response)
}

// GetSession handles GET /sessions/{id}
func (h *Handlers) GetSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from URL parameter
	sessionID := chi.URLParam(r, "id")

	// Get session from manager
	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, ErrCodeSessionNotFound, err.Error())
		return
	}

	// Build response
	response := GetSessionResponse{
		SessionID:    sess.ID,
		ContextID:    sess.ContextID,
		PageIDs:      sess.PageIDs,
		PageCount:    len(sess.PageIDs),
		CreatedAt:    sess.CreatedAt,
		LastActivity: sess.LastActivity,
		Status:       sess.Status,
	}

	writeJSON(w, http.StatusOK, response)
}

// ListSessions handles GET /sessions
func (h *Handlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Get all sessions from manager
	sessions := h.sessionManager.ListSessions()

	// Convert to API response format
	sessionInfos := make([]SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		sessionInfos = append(sessionInfos, SessionInfo{
			SessionID:    sess.ID,
			ContextID:    sess.ContextID,
			PageCount:    len(sess.PageIDs),
			CreatedAt:    sess.CreatedAt,
			LastActivity: sess.LastActivity,
			Status:       sess.Status,
		})
	}

	response := ListSessionsResponse{
		Sessions: sessionInfos,
		Count:    len(sessionInfos),
	}

	writeJSON(w, http.StatusOK, response)
}

// DestroySession handles DELETE /sessions/{id}
func (h *Handlers) DestroySession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	// Destroy session
	if err := h.sessionManager.DestroySession(sessionID); err != nil {
		// Check if it's a "not found" error
		if err.Error() == "session not found: "+sessionID {
			writeError(w, http.StatusNotFound, ErrCodeSessionNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, ErrCodeInternalError, err.Error())
		}
		return
	}

	// Return 204 No Content (success, no body)
	w.WriteHeader(http.StatusNoContent)
}

// Navigate handles POST /sessions/{id}/navigate
func (h *Handlers) Navigate(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	// Parse request body
	var req NavigateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid JSON body")
		return
	}

	// Validate URL is provided
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "URL is required")
		return
	}

	// Navigate via session manager
	pageID, err := h.sessionManager.Navigate(sessionID, req.URL)
	if err != nil {
		// Check if session not found
		if err.Error() == "failed to get session: session not found: "+sessionID {
			writeError(w, http.StatusNotFound, ErrCodeSessionNotFound, "Session not found")
		} else {
			writeError(w, http.StatusInternalServerError, ErrCodeNavigationFailed, err.Error())
		}
		return
	}

	// Build response
	response := NavigateResponse{
		SessionID: sessionID,
		PageID:    pageID,
		URL:       req.URL,
	}

	writeJSON(w, http.StatusOK, response)
}

// ExecuteJS handles POST /sessions/{id}/execute
func (h *Handlers) ExecuteJS(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	// Parse request body
	var req ExecuteJSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid JSON body")
		return
	}

	// Validate required fields
	if req.PageID == "" {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "page_id is required")
		return
	}
	if req.Script == "" {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "script is required")
		return
	}

	// Execute JavaScript via session manager
	result, err := h.sessionManager.ExecuteJavascript(sessionID, req.PageID, req.Script)
	if err != nil {
		if err.Error() == "failed to get session: session not found: "+sessionID {
			writeError(w, http.StatusNotFound, ErrCodeSessionNotFound, "Session not found")
		} else if err.Error() == "page not found in session: "+req.PageID {
			writeError(w, http.StatusNotFound, ErrCodePageNotFound, "Page not found in session")
		} else {
			writeError(w, http.StatusInternalServerError, ErrCodeExecutionFailed, err.Error())
		}
		return
	}

	// Build response
	response := ExecuteJSResponse{
		SessionID: sessionID,
		PageID:    req.PageID,
		Result:    result,
	}

	writeJSON(w, http.StatusOK, response)
}

// CaptureScreenshot handles POST /sessions/{id}/screenshot
func (h *Handlers) CaptureScreenshot(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	// Parse request body
	var req ScreenshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid JSON body")
		return
	}

	// Validate page_id
	if req.PageID == "" {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "page_id is required")
		return
	}

	// Capture screenshot via session manager
	screenshotBytes, err := h.sessionManager.CaptureScreenshot(sessionID, req.PageID)
	if err != nil {
		if err.Error() == "failed to get session: session not found: "+sessionID {
			writeError(w, http.StatusNotFound, ErrCodeSessionNotFound, "Session not found")
		} else if err.Error() == "page not found in session: "+req.PageID {
			writeError(w, http.StatusNotFound, ErrCodePageNotFound, "Page not found in session")
		} else {
			writeError(w, http.StatusInternalServerError, ErrCodeScreenshotFailed, err.Error())
		}
		return
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(screenshotBytes)

	// Build response
	format := req.Format
	if format == "" {
		format = "png"
	}

	response := ScreenshotResponse{
		SessionID:  sessionID,
		PageID:     req.PageID,
		Screenshot: encoded,
		Format:     format,
		Size:       len(screenshotBytes),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetPageContent handles GET /sessions/{id}/pages/{pageId}/content
func (h *Handlers) GetPageContent(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	pageID := chi.URLParam(r, "pageId")

	// Get page content via session manager
	content, err := h.sessionManager.GetPageContent(sessionID, pageID)
	if err != nil {
		if err.Error() == "failed to get session: session not found: "+sessionID {
			writeError(w, http.StatusNotFound, ErrCodeSessionNotFound, "Session not found")
		} else if err.Error() == "page not found in session: "+pageID {
			writeError(w, http.StatusNotFound, ErrCodePageNotFound, "Page not found in session")
		} else {
			writeError(w, http.StatusInternalServerError, ErrCodeInternalError, err.Error())
		}
		return
	}

	// Build response
	response := GetPageContentResponse{
		SessionID: sessionID,
		PageID:    pageID,
		Content:   content,
		Length:    len(content),
	}

	writeJSON(w, http.StatusOK, response)
}

// ClosePage handles DELETE /sessions/{id}/pages/{pageId}
func (h *Handlers) ClosePage(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	pageID := chi.URLParam(r, "pageId")

	// Close page via session manager
	if err := h.sessionManager.ClosePage(sessionID, pageID); err != nil {
		if err.Error() == "failed to get session: session not found: "+sessionID {
			writeError(w, http.StatusNotFound, ErrCodeSessionNotFound, "Session not found")
		} else if err.Error() == "page not found in session: "+pageID {
			writeError(w, http.StatusNotFound, ErrCodePageNotFound, "Page not found in session")
		} else {
			writeError(w, http.StatusInternalServerError, ErrCodeInternalError, err.Error())
		}
		return
	}

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
