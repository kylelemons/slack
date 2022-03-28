package slack

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"

	"github.com/kylelemons/slack/config"
)

var PublicURL = &url.URL{
	Scheme: "https",
	Host:   "slack.com",
	Path:   "/api/",
}

type Client struct {
	BaseURL *url.URL
	HTTP    *http.Client

	Tier1 *rate.Limiter

	AppToken *config.AppToken

	LogSuccess func(method string, req, resp interface{})
	LogFailure func(method string, req interface{}, err error)
	LogWarning func(method string, req, resp interface{}, warnings []string)

	Debugf func(format string, args ...interface{})
}

func NewClient(baseURL *url.URL) (*Client, error) {
	return &Client{
		BaseURL: baseURL,
		HTTP:    http.DefaultClient,
		Tier1:   rate.NewLimiter(rate.Every(1*time.Second), 3),
	}, nil
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
