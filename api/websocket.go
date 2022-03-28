package api

import (
	"encoding/json"
)

type WebsocketHello struct {
	Type        string `json:"type,omitempty"`
	Connections int    `json:"num_connections,omitempty"`
	Info        struct {
		AppID string `json:"app_id,omitempty"`
	} `json:"connection_info,omitempty"`
}

type Envelope struct {
	// Auth
	ProofToken     string            `json:"token,omitempty"`
	Authorizations []json.RawMessage `json:"authorizations,omitempty"`

	// Context
	TeamID   string `json:"team_id,omitempty"`
	APIAppID string `json:"api_app_id,omitempty"`
	Context  string `json:"event_context,omitempty"`

	// Payload
	ID      string          `json:"envelope_id,omitempty"`
	Type    string          `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type AckEnvelope struct {
	ID      string          `json:"envelope_id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
