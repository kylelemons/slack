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
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kylelemons/slack"
	"github.com/kylelemons/slack/api"
	"github.com/kylelemons/slack/config"
	"github.com/kylelemons/slack/slacktest"
)

var (
	replay = flag.Bool("replay", true, "If true, replay saved envelopes instead of connecting to slack")

	// Non-replay mode:
	websocketTokenFile = flag.String("websocket-token-file", "", "File referencing a websocket token")
	saveEnvelopes      = flag.Bool("save-envelopes", false, "If true, save envelopes to testdata for replaying")

	// General:
	verbose = flag.Bool("verbose", false, "If true, print out verbose logging")
)

//go:embed testdata/envelopes/*.json
var testdataFS embed.FS

func TestListenWebsocket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*60*time.Second)
	defer cancel()

	fix := setup(t, ctx)

	ws, err := ListenWebsocket(ctx, fix.client, fix.tok)
	if err != nil {
		t.Fatalf("ListenWebsocket failed: %s", err)
	}
	defer ws.Close()
	ws.Warnf = t.Errorf

	if *verbose {
		ws.Debugf = t.Logf
	}

	if *replay {
		go func() {
			fix.fake.Replay(t, "00_added_to_channel.json")
			fix.fake.Replay(t, "01_channel_self_joined_message.json")
			fix.fake.Replay(t, "01_channel_message.json")
			fix.fake.Replay(t, "01_channel_reaction.json")
			fix.fake.Replay(t, "01_channel_message_with_mention.json")
			fix.fake.Replay(t, "02_app_mention.json")
			fix.fake.Replay(t, "00_removed_from_channel.json")
			fix.fake.Replay(t, "03_slackbot_removal_notification.json")
			fix.fake.DisconnectClients()
			t.Logf("Done replaying!")
		}()
	} else if *saveEnvelopes {
		ws.Tee = func(env *api.Envelope) {
			f, err := os.Create(filepath.Join("testdata", "envelopes", env.ID+".json"))
			if err != nil {
				t.Errorf("Failed to open file for envelope: %s", err)
				return
			}
			enc := json.NewEncoder(f)
			enc.SetIndent("", "    ")
			if err := enc.Encode(env); err != nil {
				t.Errorf("Failed to encode envelope: %s", err)
				return
			}
		}
	}

	// TODO: Use registry.Handle instead of the unexported helper
	//   This is here for the moment to allow treating warnings as errors.
	for _, typ := range []string{
		"member_joined_channel",
		"message",
		"reaction_added",
		"app_mention",
		"channel_left",
	} {
		typ := typ
		ws.NotificationTypes().handle(typ, func(message json.RawMessage) error {
			t.Logf("Handling %q: %#q", typ, message)
			return nil
		})
	}

	if err := ws.Serve(ctx); err != nil {
		t.Errorf("Serve failed: %s", err)
	}
}

type fixtures struct {
	fake   *slacktest.Server
	client *slack.Client
	tok    config.AppToken
}

func setup(t *testing.T, ctx context.Context) (fix fixtures) {
	tok, err := config.LoadAppToken(*websocketTokenFile)
	if err != nil {
		t.Skipf("SKIP: LoadAppToken: %s", err)
	}
	fix.tok = tok

	// TODO: App manifest
	slackURL := slack.PublicURL
	if *replay {
		envelopes, err := fs.Sub(testdataFS, "testdata/envelopes")
		if err != nil {
			t.Fatalf("SETUP: Envelope subdir: %s", err)
		}
		fix.fake = slacktest.SetupServer(t, ctx, envelopes)
		slackURL = fix.fake.URL()
	}

	client, err := slack.NewClient(slackURL)
	if err != nil {
		t.Fatalf("SETUP: NewClient: %s", err)
	}
	client.AppToken = &tok
	if *verbose {
		client.Debugf = t.Logf
	}

	fix.client = client
	return fix
}
