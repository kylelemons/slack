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
