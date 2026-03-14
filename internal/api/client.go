package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client is a typed HTTP client for the GetHook management API.
type Client struct {
	base    string
	apiKey  string
	http    *http.Client
}

func New(base, apiKey string) *Client {
	return &Client{
		base:   base,
		apiKey: apiKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (c *Client) do(ctx context.Context, method, path string, body, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var e errorEnvelope
		if json.Unmarshal(raw, &e) == nil && e.Error != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, e.Error)
		}
		return fmt.Errorf("API error %d", resp.StatusCode)
	}

	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) post(ctx context.Context, path string, body, out interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

func (c *Client) patch(ctx context.Context, path string, body, out interface{}) error {
	return c.do(ctx, http.MethodPatch, path, body, out)
}

func (c *Client) delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// ── Sources ───────────────────────────────────────────────────────────────────

func (c *Client) CreateSource(ctx context.Context, name string) (*Source, error) {
	var env dataEnvelope[*Source]
	if err := c.post(ctx, "/v1/sources", map[string]string{"name": name}, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) ListSources(ctx context.Context) ([]Source, error) {
	var env dataEnvelope[[]Source]
	if err := c.get(ctx, "/v1/sources", &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) GetSource(ctx context.Context, id string) (*Source, error) {
	var env dataEnvelope[*Source]
	if err := c.get(ctx, "/v1/sources/"+id, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) DeleteSource(ctx context.Context, id string) error {
	return c.delete(ctx, "/v1/sources/"+id)
}

// ── Destinations ──────────────────────────────────────────────────────────────

func (c *Client) CreateDestination(ctx context.Context, name, destURL string, timeoutSecs int) (*Destination, error) {
	body := map[string]interface{}{
		"name":            name,
		"url":             destURL,
		"timeout_seconds": timeoutSecs,
	}
	var env dataEnvelope[*Destination]
	if err := c.post(ctx, "/v1/destinations", body, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) ListDestinations(ctx context.Context) ([]Destination, error) {
	var env dataEnvelope[[]Destination]
	if err := c.get(ctx, "/v1/destinations", &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) GetDestination(ctx context.Context, id string) (*Destination, error) {
	var env dataEnvelope[*Destination]
	if err := c.get(ctx, "/v1/destinations/"+id, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) DeleteDestination(ctx context.Context, id string) error {
	return c.delete(ctx, "/v1/destinations/"+id)
}

// ── Routes ────────────────────────────────────────────────────────────────────

func (c *Client) CreateRoute(ctx context.Context, sourceID, destID, pattern string) (*Route, error) {
	body := map[string]interface{}{
		"destination_id":     destID,
		"event_type_pattern": pattern,
	}
	if sourceID != "" {
		body["source_id"] = sourceID
	}
	var env dataEnvelope[*Route]
	if err := c.post(ctx, "/v1/routes", body, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) ListRoutes(ctx context.Context) ([]Route, error) {
	var env dataEnvelope[[]Route]
	if err := c.get(ctx, "/v1/routes", &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) DeleteRoute(ctx context.Context, id string) error {
	return c.delete(ctx, "/v1/routes/"+id)
}

// ── Events ────────────────────────────────────────────────────────────────────

func (c *Client) ListEvents(ctx context.Context, p ListEventsParams) ([]Event, error) {
	q := url.Values{}
	if p.SourceID != "" {
		q.Set("source_id", p.SourceID)
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.Direction != "" {
		q.Set("direction", p.Direction)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}

	path := "/v1/events"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var env dataEnvelope[[]Event]
	if err := c.get(ctx, path, &env); err != nil {
		return nil, err
	}
	if env.Data == nil {
		return []Event{}, nil
	}
	return env.Data, nil
}

func (c *Client) GetEvent(ctx context.Context, id string) (*EventDetail, error) {
	var env dataEnvelope[*EventDetail]
	if err := c.get(ctx, "/v1/events/"+id, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) ReplayEvent(ctx context.Context, id string) error {
	return c.post(ctx, "/v1/events/"+id+"/replay", nil, nil)
}

// ── API Keys ──────────────────────────────────────────────────────────────────

func (c *Client) ListAPIKeys(ctx context.Context) ([]APIKey, error) {
	var env dataEnvelope[[]APIKey]
	if err := c.get(ctx, "/v1/api-keys", &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (c *Client) CreateAPIKey(ctx context.Context, name string) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := c.post(ctx, "/v1/api-keys", map[string]string{"name": name}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DeleteAPIKey(ctx context.Context, id string) error {
	return c.delete(ctx, "/v1/api-keys/"+id)
}
