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
