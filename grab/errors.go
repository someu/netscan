package grab

import "errors"

var ErrMismatchedFlags = errors.New("mismatched flag/module")

var ErrInvalidArguments = errors.New("invalid arguments")

var ErrInvalidResponse = errors.New("invalid response")

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
