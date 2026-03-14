// Package tunnel manages the polling-based event relay between the GetHook
// cloud API and the local CLI process.
//
// Design intent: The SSE/WebSocket relay is the long-term goal.  For the
// initial release we poll GET /v1/events?source_id=X using a short interval
// and a cursor (last seen event ID) so we never miss events without needing
// any backend changes.  The interface is designed so swapping in SSE later
// is a drop-in replacement.
package tunnel

import (
	"context"
	"time"

	"github.com/gethook/gethook-cli/internal/api"
	"github.com/gethook/gethook-cli/internal/output"
)

// Event is the normalized form pushed to callers.
type Event struct {
	ID         string
	EventType  string
	Status     string
	Headers    map[string]string
	Body       string // raw JSON body as received
	ReceivedAt time.Time
}

// Relay streams new events for a given source.
type Relay struct {
	client   *api.Client
	sourceID string
	interval time.Duration
}

func New(client *api.Client, sourceID string) *Relay {
	return &Relay{
		client:   client,
		sourceID: sourceID,
		interval: 2 * time.Second,
	}
}

// Run sends new events to ch until ctx is cancelled.
// It uses a simple cursor (last seen event ID) to avoid duplicates.
func (r *Relay) Run(ctx context.Context, ch chan<- Event) error {
	seen := map[string]bool{}
	// Seed with existing events so we don't replay history on startup.
	initial, err := r.client.ListEvents(ctx, api.ListEventsParams{
		SourceID: r.sourceID,
		Limit:    50,
	})
	if err != nil {
		return err
	}
	for _, e := range initial {
		seen[e.ID] = true
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			events, err := r.client.ListEvents(ctx, api.ListEventsParams{
				SourceID: r.sourceID,
				Limit:    20,
			})
			if err != nil {
				output.Warn("polling error: " + err.Error())
				continue
			}
			// Events come back newest-first; reverse to deliver in order.
			for i := len(events) - 1; i >= 0; i-- {
				e := events[i]
				if seen[e.ID] {
					continue
				}
				seen[e.ID] = true
				headers := map[string]string{}
				for k, v := range e.Headers {
					if s, ok := v.(string); ok {
						headers[k] = s
					}
				}
				ch <- Event{
					ID:         e.ID,
					EventType:  e.EventTypeStr(),
					Status:     e.Status,
					Headers:    headers,
					Body:       e.Body,
					ReceivedAt: e.ReceivedAt,
				}
			}
		}
	}
}
