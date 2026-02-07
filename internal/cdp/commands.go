package cdp

import (
	"encoding/json"
	"fmt"
)

// CreateBrowserContext creates a new isolated browser context
func (c *Client) CreateBrowserContext() (string, error) {
	result, err := c.SendCommand("Target.createBrowserContext", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create browser context: %w", err)
	}

	// Parse the result to extract browserContextId
	var response struct {
		BrowserContextID string `json:"browserContextId"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return "", fmt.Errorf("failed to parse browser context response: %w", err)
	}

	return response.BrowserContextID, nil
}

// DisposeBrowserContext closes and removes a browser context
func (c *Client) DisposeBrowserContext(contextID string) error {
	params := map[string]interface{}{
		"browserContextId": contextID,
	}

	_, err := c.SendCommand("Target.disposeBrowserContext", params)
	if err != nil {
		return fmt.Errorf("failed to dispose browser context: %w", err)
	}

	return nil
}

// CreateTarget creates a new page in the specified browser context
func (c *Client) CreateTarget(url string, contextID string) (string, error) {
	params := map[string]interface{}{
		"url": url,
	}

	// Add browser context if specified
	if contextID != "" {
		params["browserContextId"] = contextID
	}

	result, err := c.SendCommand("Target.createTarget", params)
	if err != nil {
		return "", fmt.Errorf("failed to create target: %w", err)
	}

	// Parse the result to extract targetId
	var response struct {
		TargetID string `json:"targetId"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return "", fmt.Errorf("failed to parse create target response: %w", err)
	}

	return response.TargetID, nil
}

// CloseTarget closes a page/target
func (c *Client) CloseTarget(targetID string) error {
	params := map[string]interface{}{
		"targetId": targetID,
	}

	_, err := c.SendCommand("Target.closeTarget", params)
	if err != nil {
		return fmt.Errorf("failed to close target: %w", err)
	}

	return nil
}

// GetBrowserVersion returns browser version information
func (c *Client) GetBrowserVersion() (map[string]string, error) {
	result, err := c.SendCommand("Browser.getVersion", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get browser version: %w", err)
	}

	var version map[string]string
	if err := json.Unmarshal(result, &version); err != nil {
		return nil, fmt.Errorf("failed to parse version response: %w", err)
	}

	return version, nil
}