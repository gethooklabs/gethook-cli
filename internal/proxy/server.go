package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ForwardResult holds the result of forwarding a single event.
type ForwardResult struct {
	StatusCode int
	DurationMs int
	Err        error
}

// Forward POSTs the event payload to the local target URL, preserving
// original headers and adding X-GetHook-* metadata.
func Forward(ctx context.Context, targetURL string, headers map[string]string, body []byte) ForwardResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return ForwardResult{Err: fmt.Errorf("build request: %w", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GetHook-Forwarded", "1")
	req.Header.Set("X-GetHook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		return ForwardResult{DurationMs: durationMs, Err: fmt.Errorf("forward: %w", err)}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return ForwardResult{StatusCode: resp.StatusCode, DurationMs: durationMs}
}
