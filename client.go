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

package slack

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/time/rate"

	"github.com/kylelemons/slack/api"
	"github.com/kylelemons/slack/config"
	"github.com/kylelemons/slack/internal/lru"
)

var PublicURL = &url.URL{
	Scheme: "https",
	Host:   "slack.com",
	Path:   "/api/",
}

type RateLimits struct {
	WebTier     [5]*rate.Limiter
	PostTier    *rate.Limiter
	HookTier    *rate.Limiter
	ProfileTier *rate.Limiter
}

func DefaultRateLimits() *RateLimits {
	return &RateLimits{
		// https://api.slack.com/docs/rate-limits
		WebTier: [5]*rate.Limiter{
			rate.NewLimiter(rate.Every(1*time.Second), 1), // Not a real tier, this is to make WebTier[1] mean Tier 1
			rate.NewLimiter(rate.Every(1*time.Minute), 1),
			rate.NewLimiter(rate.Every(1*time.Minute), 20),
			rate.NewLimiter(rate.Every(1*time.Minute), 50),
			rate.NewLimiter(rate.Every(1*time.Minute), 100),
		},
		PostTier:    rate.NewLimiter(rate.Every(1*time.Second), 1),
		HookTier:    rate.NewLimiter(rate.Every(1*time.Second), 1),
		ProfileTier: rate.NewLimiter(rate.Every(1*time.Minute), 10),
	}
}

type Client struct {
	BaseURL *url.URL
	HTTP    *http.Client
	Limits  *RateLimits

	Manifest *config.AppManifest
	Token    AuthToken

	LogSuccess func(method string, req, resp interface{})
	LogFailure func(method string, req interface{}, err error)
	LogWarning func(method string, req, resp interface{}, warnings []string)

	Debugf func(format string, args ...interface{})

	users *lru.Cache[string, *User]
}

type AuthToken interface {
	SetHeader(header http.Header)
}

func newClient(manifest *config.AppManifest, baseURL *url.URL) *Client {
	return &Client{
		BaseURL:  baseURL,
		HTTP:     http.DefaultClient,
		Limits:   DefaultRateLimits(),
		Manifest: manifest,

		users: lru.NewCache[string, *User](1000),
	}
}

func NewAppClient(manifest *config.AppManifest, baseURL *url.URL, tok config.AppToken) (*Client, error) {
	return newClient(manifest, baseURL).WithToken(tok), nil
}

func NewBotClient(manifest *config.AppManifest, baseURL *url.URL, tok config.BotToken) (*Client, error) {
	return newClient(manifest, baseURL).WithToken(tok), nil
}

func (c *Client) WithToken(tok AuthToken) *Client {
	client := *c
	client.Token = tok
	return &client
}

func (c *Client) URL(method string) *url.URL {
	return c.BaseURL.ResolveReference(&url.URL{
		Path: method,
	})
}

func (c *Client) debugf(format string, args ...interface{}) {
	if c.Debugf == nil {
		return
	}
	c.Debugf(format, args...)
}

func (c *Client) debugHTTPRequest(req *http.Request, body []byte) {
	if c.Debugf == nil {
		return
	}
	c.Debugf("%s %s", req.Method, req.URL)
	for k, v := range req.Header {
		c.Debugf("  Header[%q] = %q", k, v)
	}
	if req.ContentLength > 0 {
		c.Debugf("  Content-Length: %d", req.ContentLength)
	}
	if len(body) > 0 {
		c.Debugf("  Body: (%d bytes) %#q", len(body), body)
	}
}

func (c *Client) debugHTTPResponse(resp *http.Response, body []byte) {
	if c.Debugf == nil {
		return
	}
	c.Debugf("%s", resp.Status)
	for k, v := range resp.Header {
		c.Debugf("  Header[%q] = %q", k, v)
	}
	if resp.ContentLength > 0 {
		c.Debugf("  Content-Length: %d", resp.ContentLength)
	}
	if len(body) > 0 {
		c.Debugf("  Body: (%d bytes) %#q", len(body), body)
	}
}

type User struct {
	// https://api.slack.com/types/user

	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// ...
}

func (c *Client) LookupUser(ctx context.Context, userID string) (*User, error) {
	return c.users.Query(userID, func() (*User, error) {
		if t, ok := c.Token.(config.BotToken); !ok {
			return nil, fmt.Errorf("LookupUser client has %T token, want %T", c.Token, t)
		}
		if !slices.Contains(c.Manifest.OAuthConfig.Scopes.Bot, "users:read") {
			return nil, fmt.Errorf(`LookupUser requires "users:read" scope`)
		}

		type UserResp struct {
			api.PostResponse
			User *User `json:"user"`
		}
		resp, err := GetJSON[UserResp](ctx, c, c.Limits.WebTier[4], "/api/users.info", url.Values{
			"user": {userID},
		})
		if err != nil {
			return nil, err
		}
		return resp.User, nil
	})
}
