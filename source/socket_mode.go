package source

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/kylelemons/slack"
	"github.com/kylelemons/slack/api"
	"github.com/kylelemons/slack/config"
	"github.com/kylelemons/slack/internal/closer"

	"golang.org/x/net/websocket"
)

type Websocket struct {
	envelopeTypes     Registry
	eventTypes        Registry
	notificationTypes Registry

	closer closer.Once
	sock   *websocket.Conn

	read  *json.Decoder
	write *json.Encoder

	Warnf  func(string, ...interface{})
	Debugf func(string, ...interface{})
}

func (ws *Websocket) Close() error {
	return ws.closer.Close()
}

func (ws *Websocket) EnvelopeTypes() *Registry {
	return &ws.envelopeTypes
}

func (ws *Websocket) EventTypes() *Registry {
	return &ws.eventTypes
}

func (ws *Websocket) NotificationTypes() *Registry {
	return &ws.notificationTypes
}

func (ws *Websocket) warnf(format string, args ...interface{}) {
	if ws.Warnf != nil {
		ws.Warnf(format, args...)
	} else if ws.Debugf != nil {
		ws.Debugf(format, args...)
	}
}

func (ws *Websocket) debugf(format string, args ...interface{}) {
	if ws.Debugf != nil {
		ws.Debugf(format, args...)
	}
}

func ListenWebsocket(ctx context.Context, client *slack.Client, token config.AppToken) (*Websocket, error) {
	req := &api.Empty{}
	resp, err := slack.PostJSON[api.ConnectionResponse](ctx, client, client.Tier1, "apps.connections.open", req)
	if err != nil {
		return nil, fmt.Errorf("ListenWebsocket: %w", err)
	}

	loc, err := url.Parse(resp.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q returned from Slack: %w", resp.URL, err)
	}

	conn, err := websocket.DialConfig(&websocket.Config{
		Version:  websocket.ProtocolVersionHybi,
		Location: loc,
		Origin:   client.URL("apps.connections.open"),
		Dialer:   nil, // TODO: use a dialer from the client?
	})
	if err != nil {
		return nil, fmt.Errorf("connecting to %q: %w", loc, err)
	}

	dec := json.NewDecoder(conn)

	var hello api.WebsocketHello
	if err := dec.Decode(&hello); err != nil {
		conn.Close()
		return nil, fmt.Errorf("receiving hello: %w", err)
	}
	if got, want := hello.Type, "hello"; got != want {
		conn.Close()
		return nil, fmt.Errorf("initial frame of type %q, want %q", got, want)
	}
	if client.Debugf != nil {
		client.Debugf("Websocket connected to app %q (connection #%d)", hello.Info.AppID, hello.Connections)
	}

	// TODO: pre-create a new socket before disconnection

	ws := &Websocket{
		closer: closer.For(conn),
		sock:   conn,
		read:   dec,
		write:  json.NewEncoder(conn),
		Warnf:  client.Debugf,
		Debugf: client.Debugf,
	}
	ws.envelopeTypes.Fallback = func(typ string, message json.RawMessage) error {
		ws.warnf("unhandled %q envelope: %#q", typ, message)
		return nil
	}
	ws.eventTypes.Fallback = func(typ string, message json.RawMessage) error {
		ws.warnf("unhandled %q event: %#q", typ, message)
		return nil
	}
	ws.notificationTypes.Fallback = func(typ string, message json.RawMessage) error {
		ws.warnf("unhandled %q notification: %#q", typ, message)
		return nil
	}

	Handle(ws.EnvelopeTypes(), "events_api", func(event *api.Event) error {
		return ws.eventTypes.Dispatch(event.Type, event.Payload)
	})
	Handle(ws.EventTypes(), "event_callback", func(raw *json.RawMessage) error {
		var wrapper struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(*raw, &wrapper); err != nil {
			return err
		}
		return ws.notificationTypes.Dispatch(wrapper.Type, *raw)
	})

	return ws, nil
}

func (w *Websocket) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		w.Close()
	}()

	buf := new(bytes.Buffer)
	asyncErr := make(chan error, 1)
	for {
		env, raw, err := w.readEnvelope()
		// if there was an error asynchronously, return it instead
		select {
		case err := <-asyncErr:
			return err
		default:
		}
		// otherwise report any error from reading the envelope
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			return err
		}

		// Print out a nicely formatted version of the envelope to aid in debugging
		if w.Debugf != nil {
			buf.Reset()
			json.Indent(buf, raw, "", "    ")
			w.Debugf("Envelope received:\n%s", buf.String())
		}

		// Handle messages asynchronously to avoid blocking future messages
		go func() {
			if err := w.envelopeTypes.Dispatch(env.Type, env.Payload); err != nil {
				w.warnf("Dispatch %q: %s", env.Type, err)
			}

			ack := &api.AckEnvelope{
				ID: env.ID,
			}
			if err := w.write.Encode(ack); err != nil {
				select {
				case asyncErr <- err:
					// Close the socket to unblock readEnvelope
					w.Close()
				default:
					// drop the error if there's already one pending
				}
				return
			}

			w.debugf("Envelope acknowledged: %s", env.ID)
		}()
	}
}

func (w *Websocket) readEnvelope() (*api.Envelope, json.RawMessage, error) {
	var raw json.RawMessage
	if err := w.read.Decode(&raw); err != nil {
		return nil, nil, fmt.Errorf("nextEvent: %w", err)
	}
	var env api.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, nil, fmt.Errorf("nextEvent: decode: %w", err)
	}
	return &env, raw, nil
}
