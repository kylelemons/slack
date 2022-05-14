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
	"context"
	"flag"
	"testing"
	"time"

	"github.com/kylelemons/slack"
	"github.com/kylelemons/slack/config"
)

var (
	websocketTokenFile = flag.String("websocket-token-file", "", "File referencing a websocket token")
)

func TestListenWebsocket(t *testing.T) {
	tok, err := config.LoadAppToken(*websocketTokenFile)
	if err != nil {
		t.Skipf("SKIP: LoadAppToken: %s", err)
	}

	client, err := slack.NewClient(slack.PublicURL)
	if err != nil {
		t.Fatalf("SETUP: NewClient: %s", err)
	}
	client.AppToken = &tok
	client.Debugf = t.Logf

	ctx := context.Background()
	ws, err := ListenWebsocket(ctx, client, tok)
	if err != nil {
		t.Fatalf("ListenWebsocket failed: %s", err)
	}
	defer ws.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := ws.Serve(ctx); err != nil {
		t.Errorf("Serve failed: %s", err)
	}
}
