package gigago

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	defaultBaseURLForAI    = "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"
	defaultBaseURLForOauth = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	defaultTimeout         = 30 * time.Second
	defaultScope           = "GIGACHAT_API_PERS"
)

// Client is the main entry point for interacting with the GigaChat API.
// It manages authentication, token refreshing, and request sending.
//
// A Client should be created using the NewClient function.
type Client struct {
	// httpClient is the underlying HTTP client used for requests.
	httpClient *http.Client
	// baseURLAI is the base URL for the main chat completions API.
	baseURLAI string
	// baseURLOauth is the base URL for the OAuth 2.0 token endpoint.
	baseURLOauth string
	// scope defines the permission scope for the access token.
	scope       string
	apiKey      string
	mu          sync.RWMutex
	wg          *sync.WaitGroup
	accessToken *tokenResponse
	ctxCancel   context.CancelFunc
}

// Option is a function type used to configure a Client.
// It's used in the NewClient constructor to customize client behavior.
type Option func(*Client)

// WithCustomURLAI provides an Option to set a custom base URL for the main AI API.
// This is primarily used for testing or connecting to a proxy.
func WithCustomURLAI(url string) Option {
	return func(c *Client) {
		c.baseURLAI = url
	}
}

// WithCustomURLOauth provides an Option to set a custom base URL for the OAuth 2.0 endpoint.
// This is primarily used for testing or connecting to a proxy.
func WithCustomURLOauth(url string) Option {
	return func(c *Client) {
		c.baseURLOauth = url
	}
}

// WithCustomClient provides an Option to use a custom http.Client.
// This is the recommended way for advanced configuration, such as setting custom
// transport for proxies or mTLS. If this option is used, it should typically
// be the FIRST option passed to NewClient, as other options like WithCustomTimeout
// or WithCustomInsecureSkipVerify will modify the provided client.
func WithCustomClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithCustomTimeout provides an Option to set a custom timeout for the http.Client.
// If WithCustomClient is also used, this option will be applied to the custom client,
// potentially overwriting its original timeout.
func WithCustomTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithCustomScope provides an Option to set a custom scope for OAuth 2.0 authorization.
// Defaults to "GIGACHAT_API_PERS" if not specified.
func WithCustomScope(scope string) Option {
	return func(c *Client) {
		c.scope = scope
	}
}

// WithCustomInsecureSkipVerify provides an Option to control SSL/TLS certificate verification.
// WARNING: Setting this to true disables certificate validation and makes the connection
// vulnerable to man-in-the-middle attacks. This should only be used for
// specific testing or development scenarios with trusted networks.
// By default, verification is enabled (false).
func WithCustomInsecureSkipVerify(insecureSkipVerify bool) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}

		transport, ok := c.httpClient.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
			c.httpClient.Transport = transport
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}

		transport.TLSClientConfig.InsecureSkipVerify = insecureSkipVerify
	}
}

// NewClient creates, configures, and returns a new Client instance.
// It requires an API key for authentication and accepts a variadic number of
// Option functions to customize its behavior (e.g., setting custom URLs or HTTP client).
//
// On initialization, it performs an initial request to obtain an access token.
// It also launches a background goroutine to automatically refresh the token before it expires.
// An error is returned if the initial token fetch fails.
func NewClient(ctx context.Context, apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey cannot be empty")
	}

	client := &Client{
		apiKey:       apiKey,
		baseURLAI:    defaultBaseURLForAI,
		baseURLOauth: defaultBaseURLForOauth,
		scope:        defaultScope,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false,
				},
			},
			Timeout: defaultTimeout,
		},
		wg: &sync.WaitGroup{},
	}

	ctxWithCancel, cancel := context.WithCancel(context.Background())
	client.ctxCancel = cancel

	for _, opt := range opts {
		opt(client)
	}

	access, err := client.oauthCreate(ctx)
	if err != nil {
		return nil, fmt.Errorf("token fetch failed: %w", err)
	}

	client.accessToken = access

	client.wg.Add(1)
	go client.tokenRefresher(ctxWithCancel)

	return client, nil
}

// Close gracefully shuts down the client. It closes idle HTTP connections
// and stops the background token refresher goroutine. It's recommended to
// call Close when the client is no longer needed to prevent resource leaks.
func (c *Client) Close() {
	c.ctxCancel()
	c.wg.Wait()
	c.httpClient.CloseIdleConnections()
}
