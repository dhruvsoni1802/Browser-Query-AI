package cdp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Target represents a CDP target (page/tab)
// We do a simple mapping from JSON response to a Go Struct
type Target struct {
    ID                    string `json:"id"`
    Type                  string `json:"type"`
    Title                 string `json:"title"`
    URL                   string `json:"url"`
    WebSocketDebuggerURL  string `json:"webSocketDebuggerUrl"`
}

// Client represents a CDP WebSocket client connection to a browser
type Client struct {
	wsURL      string                  // WebSocket URL
	conn       *websocket.Conn         // WebSocket connection
	requestID  int                     // Counter for generating unique request IDs
	pending    map[int]chan *Response  // Pending requests waiting for responses
	mu         sync.Mutex              // Protects requestID and pending map
	ctx        context.Context         // Context for cancellation
	cancel     context.CancelFunc      // Cancel function
	closeOnce  sync.Once               // Ensures Close() only runs once
}

// NewClient creates a new CDP client (doesn't connect yet)
func NewClient(wsURL string) *Client {

	//Context is used so that when we close the client, the context is done and we can cancel the background reader loop
	ctx, cancel := context.WithCancel(context.Background())

	// Return the client
	return &Client{
		wsURL: wsURL,
		conn: nil,
		requestID: 0,
		pending: make(map[int]chan *Response),
		ctx: ctx,
		cancel: cancel,
		closeOnce: sync.Once{},
	}
}

// Connect establishes the WebSocket connection and starts the message reader
func (c *Client) Connect() error {
	slog.Info("connecting to CDP WebSocket", "url", c.wsURL)

	// Create a new WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(c.wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	//Set the connection inside the client struct
	c.conn = conn

	//Start the background reader loop which is a goroutine that reads from the Websocket either responses or events
	go c.readLoop()

	slog.Info("CDP WebSocket connected successfully")
	return nil
}

// Function to read from the Websocket either responses or events
func (c *Client) readLoop() {
	// Defer ensures message reader logs when stopped
	defer func() {
		slog.Info("message reader stopped")
	}()

	// Loop forever until the context is done
	for {
		select {
		case <-c.ctx.Done():
			// Client is closing - exit silently
			return
		default:
			// Read message from WebSocket
			_, message, err := c.conn.ReadMessage()

			// If there is an error reading the message
			if err != nil {
				// Check if context was cancelled (normal shutdown)
				select {
				case <-c.ctx.Done():
					// Context cancelled - this is expected during shutdown
					return
				default:
					// Unexpected error - log it
					slog.Error("error reading WebSocket message", "error", err)
					return
				}
			}

			// Handle the message that was received from the browser
			c.handleMessage(message)
		}
	}
}

// Function to send a command to the browser and wait for the response
func (c *Client) SendCommand(method string, params map[string]interface{}) (json.RawMessage, error) {
	// Generate unique request ID
	c.mu.Lock()
	c.requestID++
	id := c.requestID
	
	// Create channel for response
	responseChan := make(chan *Response, 1)
	c.pending[id] = responseChan
	c.mu.Unlock()
	
	// Build command
	command := Command{
		ID:     id,
		Method: method,
		Params: params,
	}
	
	// Marshal to JSON
	data, err := json.Marshal(command)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	
	// Send over WebSocket
	slog.Debug("sending CDP command", "method", method, "id", id)
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		// Remove from pending since we failed to send
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to send command: %w", err)
	}
	
	// Wait for response with timeout
	select {
	case response := <-responseChan:
		// Check if response has error
		if response.Error != nil {
			return nil, fmt.Errorf("CDP error: %s (code %d)", response.Error.Message, response.Error.Code)
		}
		return response.Result, nil
		
	case <-time.After(10 * time.Second):
		// Timeout
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("command timeout after 10 seconds")
		
	case <-c.ctx.Done():
		// Client is closing
		return nil, fmt.Errorf("client closed")
	}
}

// Function to close the websocket connection
func (c *Client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		slog.Info("closing CDP client")
		
		// Cancel context (stops message reader)
		c.cancel()
		
		// Close WebSocket connection
		if c.conn != nil {
			err = c.conn.Close()
		}
		
		// Clean up pending requests
		c.mu.Lock()
		for id, ch := range c.pending {
			close(ch)
			delete(c.pending, id)
		}
		c.mu.Unlock()
	})
	
	return err
}