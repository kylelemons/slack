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

package slacktest

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/kylelemons/slack/api"
)

type Server struct {
	t    *testing.T
	ws   *websocket.Server
	http *httptest.Server
	fs   fs.FS

	envelopes         chan api.Envelope
	requestDisconnect chan struct{}
}

func SetupServer(t *testing.T, ctx context.Context, envelopes fs.FS) *Server {
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	send := make(chan api.Envelope)
	discon := make(chan struct{})

	ws := &websocket.Server{
		Config: websocket.Config{
			Location: joinURL(t, srv.URL, "apps.connections.open"),
			Origin:   joinURL(t, srv.URL, "apps.connections.open"),
			Version:  websocket.ProtocolVersionHybi,
			Header:   nil, // TODO
		},
		Handler: func(conn *websocket.Conn) {
			t.Logf("Connected!")
			defer conn.Close() // TODO: do I need to do this?

			enc := json.NewEncoder(conn)
			dec := json.NewDecoder(conn)

			if err := enc.Encode(&api.WebsocketHello{
				Type:        "hello",
				Connections: 1, // TODO
				Info: api.WebsocketHelloInfo{
					AppID: "TODO", // TODO
				},
			}); err != nil {
				t.Errorf("Sending hello: %s", err)
				return
			}

			for {
				select {
				case <-discon:
					t.Logf("Disconnecting!")
					return
				case env := <-send:
					if err := enc.Encode(env); err != nil {
						t.Errorf("Encode(%#v): %s", env, err)
						return
					}

					// TODO: async acks
					var ack api.AckEnvelope
					if err := dec.Decode(&ack); err != nil {
						t.Errorf("Decode %T: %s", ack, err)
						return
					}
					if got, want := ack.ID, env.ID; got != want {
						t.Errorf("Envelope ack = %q for message %q, want matching ack", got, want)
					}
				case <-ctx.Done():
					t.Logf("Context closed!")
					return
				}
			}
		},
	}
	mux.HandleFunc("/apps.connections.open", func(w http.ResponseWriter, req *http.Request) {
		u := joinURL(t, srv.URL, "/_/_websocket")
		u.Scheme = "ws"

		resp := struct {
			api.PostResponse
			api.ConnectionResponse
		}{
			PostResponse: api.PostResponse{
				OK: true,
			},
			ConnectionResponse: api.ConnectionResponse{
				URL: u.String(),
			},
		}
		writeJSON(t, w, resp)
	})
	mux.Handle("/_/_websocket", ws)

	return &Server{
		t:                 t,
		ws:                ws,
		http:              srv,
		fs:                envelopes,
		envelopes:         send,
		requestDisconnect: discon,
	}
}

func (s *Server) URL() *url.URL {
	u, err := url.Parse(s.http.URL)
	if err != nil {
		s.t.Errorf("httptest produced invalid URL " + s.http.URL)
		return new(url.URL)
	}
	return u
}

func (s *Server) Replay(t *testing.T, filename string) {
	data, err := fs.ReadFile(s.fs, filename)
	if err != nil {
		s.t.Errorf("Replay(%q): %s", filename, err)
		return
	}

	var env api.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		s.t.Errorf("json.Decode(%q): %s", filename, err)
		return
	}
	s.envelopes <- env
}

func (s *Server) DisconnectClients() {
	// TODO: this is a one-shot, we should go through connections and close them
	close(s.requestDisconnect)
}

func joinURL(t *testing.T, base, path string) *url.URL {
	u, err := url.Parse(base)
	if err != nil {
		t.Errorf("SETUP: invalid base URL %q", base)
		return new(url.URL)
	}
	return u.ResolveReference(&url.URL{Path: path})
}

func writeJSON(t *testing.T, w http.ResponseWriter, obj any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		t.Errorf("Failed to encode response: %s\n%#v", w, obj)
	}
}
