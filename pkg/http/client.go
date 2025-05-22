package pkg

import (
	"context"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// CreateHTTPClientWithRetry creates a new HTTP client with retry mechanism
func CreateHTTPClientWithRetry() *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	return client
}

// RetryHTTPRequest executes an HTTP request with exponential backoff retry
func RetryHTTPRequest(ctx context.Context, f func() error) error {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 1 * time.Minute

	return backoff.Retry(f, b)
}
