package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type Config struct {
	HTTPClient  *http.Client
	Timeout     time.Duration
	MaxAttempts int
	RetryWait   time.Duration
}

func NewDefault(baseURL string) (*ClientWithResponses, error) {
	return New(baseURL, Config{})
}

func New(baseURL string, config Config) (*ClientWithResponses, error) {
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = DefaultHTTPClient(config)
	}
	return NewClientWithResponses(baseURL, WithHTTPClient(httpClient))
}

func DefaultHTTPClient(config Config) *http.Client {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	maxAttempts := config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 2
	}

	retryWait := config.RetryWait
	if retryWait <= 0 {
		retryWait = 100 * time.Millisecond
	}

	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	return &http.Client{
		Timeout: timeout,
		Transport: &retryTransport{
			base:        baseTransport,
			maxAttempts: maxAttempts,
			retryWait:   retryWait,
		},
	}
}

type retryTransport struct {
	base        http.RoundTripper
	maxAttempts int
	retryWait   time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.base
	if transport == nil {
		transport = http.DefaultTransport
	}

	attempts := t.maxAttempts
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		cloned := req.Clone(req.Context())
		response, err := transport.RoundTrip(cloned)
		if !shouldRetry(req.Method, response, err) || attempt == attempts {
			return response, err
		}

		lastErr = err
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}

		if sleepErr := sleepWithContext(req.Context(), t.retryWait); sleepErr != nil {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, sleepErr
		}
	}

	return nil, lastErr
}

func shouldRetry(method string, response *http.Response, err error) bool {
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		return false
	}

	if err != nil {
		var netErr net.Error
		return reqIsTemporary(err, &netErr)
	}
	if response == nil {
		return false
	}

	switch response.StatusCode {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func reqIsTemporary(err error, target *net.Error) bool {
	if target != nil && errors.As(err, target) {
		return (*target).Timeout()
	}
	return false
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
