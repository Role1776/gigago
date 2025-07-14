package gigago

import (
	"context"
	"log"
	"time"
)

// isValid checks if the token is still fresh enough for use.
// It returns true if the token's expiration time is more than 15 minutes in the future.
// This 15-minute buffer provides a safe window to prevent using an expired token
// for requests that might take time to complete.
// The expire_at timestamp is expected to be in Unix milliseconds.
func (c *Client) isValid(expire_at int64, now time.Time) bool {
	nowMs := now.UnixNano() / int64(time.Millisecond)

	remaining := expire_at - nowMs

	fifteenMinutesMs := int64(15 * 60 * 1000)

	return remaining > fifteenMinutesMs
}

// tokenRefresher runs in a background goroutine to proactively refresh the access token.
// It wakes up periodically (every minute) to check if the current token is nearing
// expiration. If it is, it triggers a refresh. Errors during the refresh are logged
// but do not stop the refresher, allowing it to retry on the next tick.
// The goroutine terminates when the client's stop channel is closed or its context is done.
func (c *Client) tokenRefresher(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			shouldRefresh := !c.isValid(c.accessToken.ExpiresAt, time.Now())
			c.mu.RUnlock()

			if shouldRefresh {

				reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				err := c.refreshToken(reqCtx)
				cancel()

				if err != nil {

					log.Printf("gigago: failed to refresh token in background: %v", err)
				}
			}

		case <-ctx.Done():

			return
		}
	}
}

func (c *Client) refreshToken(ctx context.Context) error {
	token, err := c.oauthCreate(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = token
	return nil
}
