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
	"fmt"
)

type ErrorCode string

func (ec ErrorCode) Error() string {
	return string(ec)
}

type APIError struct {
	Code     error
	Request  interface{}
	Response []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("code %q", e.Code.Error())
}

func (e *APIError) Unwrap() error {
	return e.Code
}

type ParseError struct {
	Raw []byte
	Err error
}

func (e *ParseError) Error() string {
	first := e.Raw
	if len(first) > 13 {
		first = first[:10]
		first = append(first, "..."...)
	}

	return fmt.Sprintf("%s (%d bytes starting with %q)", e.Err, len(e.Raw), first)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
