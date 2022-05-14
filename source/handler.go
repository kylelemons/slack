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

package source

import (
	"encoding/json"
	"fmt"
)

type Registry struct {
	handlers map[string][]func(message json.RawMessage) error

	Fallback func(typ string, message json.RawMessage) error
}

func (r *Registry) handle(typ string, cb func(message json.RawMessage) error) {
	if r.handlers == nil {
		r.handlers = make(map[string][]func(message json.RawMessage) error)
	}
	r.handlers[typ] = append(r.handlers[typ], cb)
}

func (r *Registry) Dispatch(typ string, payload json.RawMessage) error {
	handlers := r.handlers[typ]
	if len(handlers) == 0 && r.Fallback != nil {
		return r.Fallback(typ, payload)
	}
	var errors []error
	for _, f := range r.handlers[typ] {
		if err := f(payload); err != nil {
			errors = append(errors, err)
		}
	}
	switch len(errors) {
	case 0:
		return nil
	case 1:
		return errors[0]
	default:
		return fmt.Errorf("%d errors: %q", len(errors), errors)
	}
}

func Handle[Payload any](reg *Registry, typ string, handler func(*Payload) error) {
	reg.handle(typ, func(message json.RawMessage) error {
		var payload Payload
		if err := json.Unmarshal(message, &payload); err != nil {
			return fmt.Errorf("%s: decode %T: %w", typ, payload, err)
		}
		if err := handler(&payload); err != nil {
			return fmt.Errorf("%s: handle %T: %w", typ, payload, err)
		}
		return nil
	})
}
