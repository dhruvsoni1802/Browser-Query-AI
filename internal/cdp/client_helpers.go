package cdp

import (
	"encoding/json"
	"log/slog"
)

// Function to handle the message that was received from the browser
func (c *Client) handleMessage(message []byte) {
	// Try to parse as Response first (has "id" field)
	var response Response
	if err := json.Unmarshal(message, &response); err != nil {
		slog.Error("failed to unmarshal message", "error", err)
		return
	}
	
	// If it has an ID, it's a response to our command
	if response.ID != 0 {
		c.handleResponse(&response)
		return
	}
	
	// Otherwise, it's an event
	var event Event
	if err := json.Unmarshal(message, &event); err != nil {
		slog.Error("failed to unmarshal event", "error", err)
		return
	}
	
	c.handleEvent(&event)
}

// Function to handle the response that was received from the browser
// handleResponse matches response to pending request
func (c *Client) handleResponse(response *Response) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Find the channel waiting for this response
	ch, exists := c.pending[response.ID]
	if !exists {
		slog.Warn("received response for unknown request ID", "id", response.ID)
		return
	}
	
	// Send response to the waiting channel
	ch <- response
	
	// Remove from pending map
	delete(c.pending, response.ID)
}

// Function to handle the event that was received from the browser
func (c *Client) handleEvent(event *Event) {
	// For now, just log events
	// TODO: Later, we can add event handlers
	slog.Debug("received CDP event", "method", event.Method)
}