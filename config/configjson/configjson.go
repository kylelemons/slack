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

package configjson

type DisplayInformation struct {
	Name string `json:"name,omitempty"`
}

type AppFeatures struct {
	BotUser BotUserFeatures `json:"bot_user,omitempty"`
}

type BotUserFeatures struct {
	DisplayName  string `json:"display_name,omitempty"`
	AlwaysOnline bool   `json:"always_online,omitempty"`
}

type OAuthConfig struct {
	Scopes OAuthScopes `json:"scopes,omitempty"`
}

type OAuthScopes struct {
	Bot []string `json:"bot,omitempty"`
}

type AppSettings struct {
	EventSubscriptions   EventSubscriptions    `json:"event_subscriptions"`
	Interactivity        InteractivitySettings `json:"interactivity"`
	OrgDeployEnabled     bool                  `json:"org_deploy_enabled,omitempty"`
	SocketModeEnabled    bool                  `json:"socket_mode_enabled,omitempty"`
	TokenRotationEnabled bool                  `json:"token_rotation_enabled,omitempty"`
}

type EventSubscriptions struct {
	BotEvents []string `json:"bot_events,omitempty"`
}

type InteractivitySettings struct {
	IsEnabled bool `json:"is_enabled,omitempty"`
}
