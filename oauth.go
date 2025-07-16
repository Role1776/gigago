package gigago

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

func (c *Client) oauthCreate(ctx context.Context) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("scope", c.scope)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURLOauth, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Set a unique request ID for tracing, as required by the Sberbank API.
	req.Header.Set("RqUID", uuid.NewString())
	req.Header.Set("Authorization", "Basic "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oauth request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var token tokenResponse
	if err = json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &token, nil
}
