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

package config

import (
	"github.com/kylelemons/slack/config/configjson"
)

// AppManifest represents the application manifest provided by Slack.
type AppManifest struct {
	DisplayInformation configjson.DisplayInformation `json:"display_information,omitempty"`
	Features           configjson.AppFeatures        `json:"features,omitempty"`
	OAuthConfig        configjson.OAuthConfig        `json:"oauth_config,omitempty"`
	Settings           configjson.AppSettings        `json:"settings,omitempty"`
}

// LoadManifest loads an app manifest in JSON format.
//
// The filename must end in .json.
func LoadManifest(filename string) (*AppManifest, error) {
	return loadFile[AppManifest](filename)
}
