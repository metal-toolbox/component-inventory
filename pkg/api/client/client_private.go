package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type httpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Option allows setting custom parameters during construction
type Option func(*componentInventoryClient) error

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer httpRequestDoer) Option {
	return func(c *componentInventoryClient) error {
		c.client = doer
		return nil
	}
}

// WithAuthToken sets the client auth token.
func WithAuthToken(authToken string) Option {
	return func(c *componentInventoryClient) error {
		c.authToken = authToken
		return nil
	}
}

func (c *componentInventoryClient) get(ctx context.Context, path string) ([]byte, error) {
	requestURL, err := url.Parse(fmt.Sprintf("%s%s", c.serverAddress, path))
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error in GET request: %v", err)
	}

	return c.do(req)
}

func (c *componentInventoryClient) post(ctx context.Context, path string, body []byte) ([]byte, error) {
	requestURL, err := url.Parse(fmt.Sprintf("%s%s", c.serverAddress, path))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL path %v: %v", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error in POST request: %v", err)
	}

	return c.do(req)
}

func (c *componentInventoryClient) do(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")

	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", c.authToken))
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v, code: %v", err, response.StatusCode)
	}

	if response == nil {
		return nil, fmt.Errorf("got empty response body. code: 0")
	}

	if response.StatusCode >= http.StatusMultiStatus {
		return nil, fmt.Errorf("got bad request. code: %v", response.StatusCode)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v, code: %v", err.Error(), response.StatusCode)
	}

	return data, nil
}
