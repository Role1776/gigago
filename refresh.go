package gigago

import (
	"context"
	"log"
	"time"
)

const (
	// tokenRefreshBuffer is the time buffer before token expiration to trigger refresh
	tokenRefreshBuffer = 15 * time.Minute
	// tokenRefreshInterval is how often to check if token needs refresh
	tokenRefreshInterval = 1 * time.Minute
	// refreshTimeout is the timeout for token refresh requests
	refreshTimeout = 30 * time.Second
)

// isValid checks if the token is still fresh enough for use.
// It returns true if the token's expiration time is more than 15 minutes in the future.
// This 15-minute buffer provides a safe window to prevent using an expired token
// for requests that might take time to complete.
// The expire_at timestamp is expected to be in Unix milliseconds.
func (c *Client) isValid(expire_at int64, now time.Time) bool {
	nowMs := now.UnixNano() / int64(time.Millisecond)

	remaining := expire_at - nowMs

	fifteenMinutesMs := int64(tokenRefreshBuffer / time.Millisecond)

	return remaining > fifteenMinutesMs
}

// tokenRefresher runs in a background goroutine to proactively refresh the access token.
// It wakes up periodically (every minute) to check if the current token is nearing
// expiration. If it is, it triggers a refresh. Errors during the refresh are logged
// but do not stop the refresher, allowing it to retry on the next tick.
// The goroutine terminates when the client's stop channel is closed or its context is done.
func (c *Client) tokenRefresher(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(tokenRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if context is cancelled before proceeding
			if ctx.Err() != nil {
				return
			}

			c.mu.RLock()
			shouldRefresh := !c.isValid(c.accessToken.ExpiresAt, time.Now())
			c.mu.RUnlock()

			if shouldRefresh {
				reqCtx, cancel := context.WithTimeout(ctx, refreshTimeout)
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

// TODO: Рассмотреть возможность добавления отдельного мьютекса для защиты от проблемы "Thundering Herd"
// В текущей реализации это может привести к лишним запросам на аутентификацию.
// Пока нагрузка и лимиты это позволяют, оставляем как есть для простоты.
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
