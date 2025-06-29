package wraper

import (
	"fmt"
)

type Wraper struct {
	FuncName string
}

func New(funcName string) Wraper {
	return Wraper{
		FuncName: funcName,
	}
}

func (wp Wraper) WrapMsg(msg string, err error) error {
	if isNil(err) {
		return nil
	}

	if msg == "" {
		return wp.Wrap(err)
	}

	return fmt.Errorf("%s: %s: %w", wp.FuncName, msg, err)
}

func (wp Wraper) Wrapf(err error, format string, args ...any) error {
	return wp.WrapMsg(fmt.Sprintf(format, args...), err)
}

func (wp Wraper) Wrap(err error) error {
	if isNil(err) {
		return nil
	}

	return fmt.Errorf("%s: %w", wp.FuncName, err)
}

func (wp Wraper) WrapN(errs ...error) error {
	var n int
	for _, err := range errs {
		if !isNil(err) {
			n++
		}
	}

	if n == 0 {
		return nil
	}

	nonNilErrs := make([]error, 0, n)
	for _, err := range errs {
		if !isNil(err) {
			nonNilErrs = append(nonNilErrs, err)
		}
	}

	result := nonNilErrs[0]

	for _, err := range nonNilErrs[1:] {
		result = fmt.Errorf("%w: %w", result, err)
	}

	return wp.Wrap(result)
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

func WrapN(fn string, errs ...error) error {
	return New(fn).WrapN(errs...)
}

func isNil(err error) bool {
	return err == nil
}
