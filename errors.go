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
	return e.Code.Error()
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
