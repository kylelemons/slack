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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/time/rate"

	"github.com/kylelemons/slack/api"
)

// TODO: Handle HTTP 429 responses

func GetJSON[Resp any](ctx context.Context, client *Client, tier *rate.Limiter, method string, params url.Values) (resp *Resp, err error) {
	if client.LogFailure != nil {
		defer func() {
			if err != nil {
				client.LogFailure(method, params, err)
			}
		}()
	}

	url := client.URL(method)
	url.RawQuery = params.Encode()

	httpReq := &http.Request{
		Method: http.MethodGet,
		URL:    url,
		Header: http.Header{
			// TODO: Accept?
		},
	}
	if client.Token != nil {
		client.Token.SetHeader(httpReq.Header)
	}

	if err := tier.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%s: timeout waiting for rate limit: %w", method, err)
	}

	client.debugHTTPRequest(httpReq, nil)
	return doJSON[Resp](ctx, client, params, httpReq)
}

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
			// TODO: Accept?
			"Content-Type": {"application/json;charset=UTF-8"},
		},

		ContentLength: int64(buf.Len()),
		Body:          ioutil.NopCloser(buf),
	}
	if client.Token != nil {
		client.Token.SetHeader(httpReq.Header)
	}

	if err := tier.Wait(ctx); err != nil {
		return nil, fmt.Errorf("timeout waiting for rate limit: %w", method, err)
	}
	client.debugHTTPRequest(httpReq, buf.Bytes())

	return doJSON[Resp](ctx, client, req, httpReq)
}

func doJSON[Resp any](ctx context.Context, client *Client, req any, httpReq *http.Request) (*Resp, error) {
	method := httpReq.Method

	httpResp, err := client.HTTP.Do(httpReq.WithContext(ctx))
	if err != nil {
		client.debugf("http.Do error: (%T) %q", err, err)
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
		return nil, fmt.Errorf("%s: response not OK: %w",
			method, &APIError{
				Code:     ErrorCode(apiResp.ErrorCode),
				Request:  req,
				Response: respJSON,
			})
	}

	resp := new(Resp)
	if err := json.NewDecoder(bytes.NewReader(respJSON)).Decode(resp); err != nil {
		return nil, fmt.Errorf("%s: decoding response as %T: %w",
			method, resp, &ParseError{Raw: respJSON, Err: err})
	}

	if len(apiResp.Warning) > 0 {
		// TODO: have a client.Warnf like we have for websocket?
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
