// Copyright 2022 Kyle Lemons
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
)

type Empty struct {
}

type PostResponse struct {
	OK        bool   `json:"ok"`
	Warning   string `json:"warning,omitempty"`
	ErrorCode string `json:"error,omitempty"`
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
