package wraper

import (
	"fmt"
)

var (
	_ error = &Error{}
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

	return &Error{
		Fn:  fn,
		Msg: msg,
		Err: err,
	}
}

func (e *Error) Error() string {
	if isNil(e.Err) {
		return ""
	}

	if isEmptyMsg(e.Msg) {
		return fmt.Sprintf("%s: %s", e.Fn, e.Err.Error())
	}

	return fmt.Sprintf("%s: %s: %s", e.Fn, e.Msg, e.Err.Error())
}

func (e *Error) Unwrap() error {
	return e.Err
}

func isNil(err error) bool {
	return err == nil
}

func isEmptyMsg(msg string) bool {
	return msg == emptyMsg
}
