package grab

import "errors"

// ErrMismatchedFlags is thrown if the flags for one module type are
// passed to an incompatible module type.
var ErrMismatchedFlags = errors.New("mismatched flag/module")

// ErrInvalidArguments is thrown if the command-line arguments invalid.
var ErrInvalidArguments = errors.New("invalid arguments")

// ErrInvalidResponse is returned when the server returns a syntactically-invalid response.
var ErrInvalidResponse = errors.New("invalid response")

// ErrUnexpectedResponse is returned when the server returns a syntactically-valid but unexpected response.
var ErrUnexpectedResponse = errors.New("unexpected response")

type errTotalTimeout string

const (
	ErrTotalTimeout = errTotalTimeout("timeout")
)

func (err errTotalTimeout) Error() string {
	return string(err)
}

func (err errTotalTimeout) Timeout() bool {
	return true
}

func (err errTotalTimeout) Temporary() bool {
	return false
}