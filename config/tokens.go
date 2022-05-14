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
	"bytes"
	"fmt"
	"os"
	"strings"
)

type AppToken string

func (t AppToken) String() string {
	return string(t)
}

func LoadAppToken(filename string) (AppToken, error) {
	const TokenPrefix = "xapp-"

	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("reading app token: %w", err)
	}

	s := string(bytes.TrimSpace(data))
	if !strings.HasPrefix(s, TokenPrefix) {
		return "", fmt.Errorf("token in %q does not have prefix %q", filename, TokenPrefix)
	}

	return AppToken(s), nil
}

type BotToken string

func (t BotToken) String() string {
	return string(t)
}

func LoadBotToken(filename string) (BotToken, error) {
	const TokenPrefix = "xoxb-"

	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("reading app token: %w", err)
	}

	s := string(bytes.TrimSpace(data))
	if !strings.HasPrefix(s, TokenPrefix) {
		return "", fmt.Errorf("token in %q does not have prefix %q", filename, TokenPrefix)
	}

	return BotToken(s), nil
}
