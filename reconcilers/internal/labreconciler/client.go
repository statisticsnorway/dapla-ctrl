package labreconciler

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type client struct {
	endpoint   string
	secret     string
	httpClient *http.Client
}

type optFunc func(*client)

func WithHttpClient(httpClient *http.Client) optFunc {
	return func(c *client) {
		c.httpClient = httpClient
	}
}

func New(endpoint, secret string, opts ...optFunc) (*client, error) {
	c := &client{
		endpoint: endpoint,
		secret:   secret,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return c, nil
}

func (c *client) do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.secret))

	return c.httpClient.Do(req)
}

func (c *client) EnableFfunkEditing(ctx context.Context, team string) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/ffunk-editering/%s", c.endpoint, team), nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotModified {
		return nil
	}

	return fmt.Errorf("unexpected response code %d: %s", resp.StatusCode, resp.Status)
}

func (c *client) HasFfunkEditing(ctx context.Context, team string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/ffunk-editering/%s", c.endpoint, team), nil)
	if err != nil {
		return false, err
	}

	resp, err := c.do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected response code %d: %s", resp.StatusCode, resp.Status)
}
func (c *client) DisableFfunkEditing(ctx context.Context, team string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/ffunk-editering/%s", c.endpoint, team), nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotModified {
		return nil
	}

	return fmt.Errorf("unexpected response code %d: %s", resp.StatusCode, resp.Status)
}
