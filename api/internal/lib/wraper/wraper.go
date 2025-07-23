package wraper

import (
	"fmt"
)

type Wraper struct {
	FuncName string
}

const (
	emptyMsg = ""
)

func New(funcName string) Wraper {
	return Wraper{
		FuncName: funcName,
	}
}

func (wp Wraper) WrapMsg(msg string, err error) error {
	return newError(wp.FuncName, msg, err)
}

func (wp Wraper) Wrapf(err error, format string, args ...any) error {
	return wp.WrapMsg(fmt.Sprintf(format, args...), err)
}

func (wp Wraper) Wrap(err error) error {
	return newError(wp.FuncName, emptyMsg, err)
}

func Wrap(fn string, err error) error {
	return New(fn).Wrap(err)
}

func WrapMsg(fn string, msg string, err error) error {
	return New(fn).WrapMsg(msg, err)
}

func Wrapf(fn string, err error, format string, args ...any) error {
	return New(fn).Wrapf(err, format, args...)
}

func IsWraped(err error) *Error {
	seen := make(map[error]struct{})
	return isWraped(err, seen)
}

func isWraped(err error, seen map[error]struct{}) *Error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		if _, ok := seen[err]; ok {
			return nil
		}

		switch x := err.(type) {
		case interface{ Unwrap() *Error }:
			return x.Unwrap()

		case interface{ Unwrap() error }:
			u := x.Unwrap()
			if u == nil {
				return nil
			}

			seen[err] = struct{}{}

			return isWraped(u, seen)

		case interface{ Unwrap() []error }:
			for _, er := range x.Unwrap() {
				if er != nil {
					seen[err] = struct{}{}
					e := isWraped(er, seen)
					if e != nil {
						return e
					}
				}
			}
		}

		return nil
	}

	return e
}
