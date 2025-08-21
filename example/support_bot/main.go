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

package main

import (
	"context"
	"flag"
	"log"

	"github.com/kylelemons/slack"
	"github.com/kylelemons/slack/api"
	"github.com/kylelemons/slack/config"
	"github.com/kylelemons/slack/slackapp"
	"github.com/kylelemons/slack/source"
)

var (
	manifest = flag.String("manifest", "./manifest.json", "Path to manifest JSON")
	appToken = flag.String("app-token", "./app_token.txt", "Path to app token file")
	botToken = flag.String("bot-token", "./bot_token.txt", "Path to bot token file")
	debug    = flag.Bool("debug", false, "If true, print debugging")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	manifest, err := config.LoadManifest(*manifest)
	if err != nil {
		log.Fatalf("LoadManifest: %s", err)
	}
	log.Printf("Launching app: %s", manifest.DisplayInformation.Name)

	appTok, err := config.LoadAppToken(*appToken)
	if err != nil {
		log.Fatalf("LoadAppToken: %s", err)
	}
	botTok, err := config.LoadBotToken(*botToken)
	if err != nil {
		log.Fatalf("LoadBotToken: %s", err)
	}

	client, err := slack.NewBotClient(manifest, slack.PublicURL, botTok)
	if err != nil {
		log.Fatalf("NewAppClient: %s", err)
	}
	if *debug {
		client.Debugf = log.Printf
	}
	client.LogWarning = func(method string, req, resp interface{}, warnings []string) {
		for _, warning := range warnings {
			log.Printf("WARN [%s] (%+v) %s", method, req, warning)
		}
	}
	client.LogFailure = func(method string, req interface{}, err error) {
		log.Printf("ERROR [%s] (%+v) %s", method, req, err)
	}

	ws, err := source.ListenWebsocket(ctx, client, appTok)
	if err != nil {
		log.Fatalf("ListenWebsocket: %s", err)
	}
	if *debug {
		ws.Debugf = log.Printf
	} else {
		ws.Warnf = log.Printf
	}

	app := slackapp.New(manifest, client, ws)
	app.OnChannelMessage(func(ev *api.MessageEvent, channelID, userID, message string) {
		if u, err := client.WithToken(botTok).LookupUser(ctx, userID); err == nil {
			log.Printf("%s <%s : %s> %s", ev.TS, channelID, u.Name, message)
		} else {
			log.Printf("%s <%s : %s> %s (err: %s)", ev.TS, channelID, userID, message, err)
		}
	})
	app.OnChannelThreadedMessage(func(ev *api.MessageEvent, channelID, userID, threadTS, message string) {
		log.Printf("--> %s <%s : %s> %s", threadTS, channelID, userID, message)
	})

	app.Run(ctx)
}
