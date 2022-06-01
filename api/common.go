package api

import (
	"encoding/json"
)

type Empty struct {
}

type PostResponse struct {
	OK        bool   `json:"ok"`
	Warning   string `json:"warning"`
	ErrorCode string `json:"error"`
}

type Event struct {
	// Auth
	ProofToken     string            `json:"token,omitempty"`
	Authorizations []json.RawMessage `json:"authorizations,omitempty"`

	// Context
	TeamID   string `json:"team_id,omitempty"`
	APIAppID string `json:"api_app_id,omitempty"`
	Context  string `json:"event_context,omitempty"`

	// Event
	ID       string          `json:"event_id,omitempty"`
	UnixTime int             `json:"event_time,omitempty"`
	Type     string          `json:"type,omitempty"`
	Payload  json.RawMessage `json:"event,omitempty"`

	// Debugging
	Envelope json.RawMessage `json:"-"`
	Raw      json.RawMessage `json:"-"`
}
