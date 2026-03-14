package api

import "time"

// ── Accounts ──────────────────────────────────────────────────────────────────

type Account struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}

// ── API Keys ──────────────────────────────────────────────────────────────────

type APIKey struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"key_prefix"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// ── Sources ───────────────────────────────────────────────────────────────────

type Source struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PathToken string    `json:"path_token"`
	Status    string    `json:"status"`
	AuthMode  string    `json:"auth_mode"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Destinations ──────────────────────────────────────────────────────────────

type Destination struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	URL            string    `json:"url"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	CreatedAt      time.Time `json:"created_at"`
}

// ── Routes ────────────────────────────────────────────────────────────────────

type Route struct {
	ID               string    `json:"id"`
	SourceID         *string   `json:"source_id,omitempty"`
	DestinationID    string    `json:"destination_id"`
	EventTypePattern string    `json:"event_type_pattern"`
	CreatedAt        time.Time `json:"created_at"`
}

// ── Events ────────────────────────────────────────────────────────────────────

type Event struct {
	ID             string            `json:"id"`
	Direction      string            `json:"direction"`
	SourceID       *string           `json:"source_id,omitempty"`
	EventType      string            `json:"event_type"`
	Status         string            `json:"status"`
	AttemptsCount  int               `json:"attempts_count"`
	Headers        map[string]string `json:"headers,omitempty"`
	Payload        interface{}       `json:"payload,omitempty"`
	ReceivedAt     time.Time         `json:"received_at"`
	NextAttemptAt  *time.Time        `json:"next_attempt_at,omitempty"`
}

type DeliveryAttempt struct {
	ID            string    `json:"id"`
	EventID       string    `json:"event_id"`
	DestinationID string    `json:"destination_id"`
	AttemptNumber int       `json:"attempt_number"`
	Outcome       string    `json:"outcome"`
	ResponseStatus *int     `json:"response_status,omitempty"`
	ResponseBody  string    `json:"response_body,omitempty"`
	AttemptedAt   time.Time `json:"attempted_at"`
	DurationMs    int       `json:"duration_ms,omitempty"`
}

type EventDetail struct {
	Event
	Attempts []DeliveryAttempt `json:"attempts,omitempty"`
}

// ── List params ───────────────────────────────────────────────────────────────

type ListEventsParams struct {
	SourceID  string
	Status    string
	Direction string
	Limit     int
	After     string // cursor: last event ID seen
}

// ── Wrapped responses ─────────────────────────────────────────────────────────

type dataEnvelope[T any] struct {
	Data T `json:"data"`
}

type errorEnvelope struct {
	Error string `json:"error"`
}
