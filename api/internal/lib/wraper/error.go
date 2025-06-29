package wraper

import (
	"fmt"
)

type Error struct {
	Fn  string
	Msg string
	Err error
}

func newError(fn string, msg string, err error) error {
	if isNil(err) {
		return nil
	}

	return Error{
		Fn:  fn,
		Msg: msg,
		Err: err,
	}
}

func (e Error) Error() string {
	return e.String()
}

func (e Error) String() string {
	if isNil(e.Err) {
		return ""
	}

	if isEmptyMsg(e.Msg) {
		return fmt.Errorf("%s: %w", e.Fn, e.Err).Error()
	}

	return fmt.Errorf("%s: %s: %w", e.Fn, e.Msg, e.Err).Error()
}

func (e Error) Unwrap() error {
	return e.Err
}
