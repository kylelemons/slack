package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/time/rate"

	"github.com/kylelemons/slack/api"
)

func PostJSON[Resp any](ctx context.Context, client *Client, tier *rate.Limiter, method string, req any) (resp *Resp, err error) {
	if client.LogFailure != nil {
		defer func() {
			if err != nil {
				client.LogFailure(method, req, err)
			}
		}()
	}

	url := client.URL(method)

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(req); err != nil {
		return nil, fmt.Errorf("%s: encoding request as %T: %w", method, req, err)
	}

	httpReq := &http.Request{
		Method: http.MethodPost,
		URL:    url,
		Header: http.Header{
			"Content-Type": {"application/json;charset=UTF-8"},
		},

		ContentLength: int64(buf.Len()),
		Body:          ioutil.NopCloser(buf),
	}
	if client.AppToken != nil {
		httpReq.Header.Set("Authorization", "Bearer "+client.AppToken.String())
	}

	if err := tier.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%s: timeout waiting for rate limit: %w", method, err)
	}

	client.debugHTTPRequest(httpReq, buf.Bytes())
	httpResp, err := client.HTTP.Do(httpReq.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("%s: request failed: %w", method, err)
	}
	defer httpResp.Body.Close()

	// TODO: set upper limit on size?
	respJSON, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: reading response: %w", method, err)
	}
	client.debugHTTPResponse(httpResp, respJSON)

	var apiResp api.PostResponse
	if err := json.NewDecoder(bytes.NewReader(respJSON)).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("%s: decoding json header: %w",
			method, &ParseError{Raw: respJSON, Err: err})
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("%s: request failed: %w",
			method, &APIError{
				Code:     ErrorCode(apiResp.ErrorCode),
				Request:  req,
				Response: respJSON,
			})
	}

	resp = new(Resp)
	if err := json.NewDecoder(bytes.NewReader(respJSON)).Decode(resp); err != nil {
		return nil, fmt.Errorf("%s: decoding response as %T: %w",
			method, resp, &ParseError{Raw: respJSON, Err: err})
	}

	if len(apiResp.Warning) > 0 {
		warnings := strings.Split(apiResp.Warning, ",")
		client.Debugf("Warning(s): %q", warnings)
		if client.LogWarning != nil {
			client.LogWarning(method, req, resp, warnings)
		}
	}
	if client.LogSuccess != nil {
		client.LogSuccess(method, req, resp)
	}

	return resp, nil
}
