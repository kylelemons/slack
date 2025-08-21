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

type MessageEvent struct {
	TS        string `json:"event_ts,omitempty"`  // event timestamp / key (used for child threads, etc)
	Subtype   string `json:"subtype,omitempty"`   // event subtype
	UserID    string `json:"user,omitempty"`      // source
	ChannelID string `json:"channel,omitempty"`   // destination
	Text      string `json:"text,omitempty"`      // message text
	Blocks    []any  `json:"blocks,omitempty"`    // message blocks
	ThreadTS  string `json:"thread_ts,omitempty"` // thread timestamp / key, if any
}
