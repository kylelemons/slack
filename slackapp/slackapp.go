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

package slackapp

import (
	"context"
	"log"

	"golang.org/x/exp/slices"

	"github.com/kylelemons/slack"
	"github.com/kylelemons/slack/api"
	"github.com/kylelemons/slack/config"
	"github.com/kylelemons/slack/source"
)

type Source interface {
	NotificationTypes() *source.Registry

	Serve(ctx context.Context) error
}

type SlackApp struct {
	manifest *config.AppManifest
	client   *slack.Client
	source   Source
}

func New(manifest *config.AppManifest, client *slack.Client, source Source) *SlackApp {
	return &SlackApp{
		manifest: manifest,
		client:   client,
		source:   source,
	}
}

func (a *SlackApp) Run(ctx context.Context) {
	if err := a.source.Serve(ctx); err != nil {
		log.Fatalf("Serving ")
	}
}

func (a *SlackApp) OnChannelMessage(call func(event *api.MessageEvent, channelID, userID, message string)) {
	if !slices.Contains(a.manifest.Settings.EventSubscriptions.BotEvents, "message.channels") {
		log.Panicf(`OnChannelMessage requires subscription to "message.channels" event`)
	}

	source.Handle(a.source.NotificationTypes(), "message", func(payload *api.MessageEvent) error {
		if payload.ThreadTS == "" && payload.Subtype == "" {
			call(payload, payload.ChannelID, payload.UserID, payload.Text)
		}
		return nil
	})
}

func (a *SlackApp) OnChannelThreadedMessage(call func(event *api.MessageEvent, channelID, userID, threadTS, message string)) {
	if !slices.Contains(a.manifest.Settings.EventSubscriptions.BotEvents, "message.channels") {
		log.Panicf(`OnChannelMessage requires subscription to "message.channels" event`)
	}

	source.Handle(a.source.NotificationTypes(), "message", func(payload *api.MessageEvent) error {
		if payload.ThreadTS != "" && payload.Subtype == "" {
			call(payload, payload.ChannelID, payload.UserID, payload.ThreadTS, payload.Text)
		}
		return nil
	})
}
