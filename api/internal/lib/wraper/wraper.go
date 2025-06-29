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
	if isNil(err) {
		return nil
	}

	if isEmptyMsg(msg) {
		return wp.Wrap(err)
	}

	return newError(wp.FuncName, msg, err)
}

func (wp Wraper) Wrapf(err error, format string, args ...any) error {
	return wp.WrapMsg(fmt.Sprintf(format, args...), err)
}

func (wp Wraper) Wrap(err error) error {
	if isNil(err) {
		return nil
	}

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

func isNil(err error) bool {
	return err == nil
}

func isEmptyMsg(msg string) bool {
	return msg == emptyMsg
}
