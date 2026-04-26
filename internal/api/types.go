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
	IngestURL string    `json:"ingest_url"` // pre-built by backend
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

// Event matches the backend events.Event JSON shape exactly.
type Event struct {
	ID              string                 `json:"id"`
	AccountID       string                 `json:"account_id"`
	Direction       string                 `json:"direction"`
	SourceID        *string                `json:"source_id,omitempty"`
	EventType       *string                `json:"event_type,omitempty"`
	ExternalEventID *string                `json:"external_event_id,omitempty"`
	ReceivedAt      time.Time              `json:"received_at"`
	Headers         map[string]interface{} `json:"headers"`
	Body            string                 `json:"body"`
	Status          string                 `json:"status"`
	AttemptsCount   int                    `json:"attempts_count"`
	CreatedAt       time.Time              `json:"created_at"`
}

// EventType returns the event type string, or "unknown" if nil.
func (e Event) EventTypeStr() string {
	if e.EventType != nil {
		return *e.EventType
	}
	return "unknown"
}

// DeliveryAttempt matches the backend events.DeliveryAttempt JSON shape.
type DeliveryAttempt struct {
	ID             string     `json:"id"`
	EventID        string     `json:"event_id"`
	DestinationID  string     `json:"destination_id"`
	AttemptNumber  int        `json:"attempt_number"`
	StartedAt      time.Time  `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	ResponseStatus *int       `json:"response_status,omitempty"`
	ResponseBody   *string    `json:"response_body,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	LatencyMS      *int       `json:"latency_ms,omitempty"`
	Outcome        string     `json:"outcome"`
}

// EventDetail matches { "event": {...}, "attempts": [...] } from GET /v1/events/{id}.
type EventDetail struct {
	Event    *Event            `json:"event"`
	Attempts []DeliveryAttempt `json:"attempts"`
}

// ── List params ───────────────────────────────────────────────────────────────

type ListEventsParams struct {
	SourceID  string
	Status    string
	Direction string
	Limit     int
}

// ── Wrapped responses ─────────────────────────────────────────────────────────

// dataEnvelope is used when the backend returns { "data": T }.
type dataEnvelope[T any] struct {
	Data T `json:"data"`
}

// eventListData matches { "data": { "items": [...], "total": N } }.
type eventListData struct {
	Items []*Event `json:"items"`
	Total int      `json:"total"`
}

type errorEnvelope struct {
	Error string `json:"error"`
}
