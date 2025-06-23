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
	if err == nil {
		return nil
	}

	if msg == "" {
		return fmt.Errorf("%s: %w", wp.FuncName, err)
	}

	return fmt.Errorf("%s: %s: %w", wp.FuncName, msg, err)
}

func (wp Wraper) Wrapf(err error, format string, args ...any) error {
	return wp.WrapMsg(fmt.Sprintf(format, args...), err)
}

func (wp Wraper) Wrap(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", wp.FuncName, err)
}

func Wrap(fn string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", fn, err)
}

// WrapMsg annotates err with additional context while preserving the original error.
// The message is formatted as "fn: msg: err" if msg is non-empty,
// or "fn: err" if msg is empty.
//
// The returned error implements an Unwrap method returning the original err,
// making it compatible with errors.Is and errors.As.
//
// If err is nil, Wrap returns nil.
func WrapMsg(fn string, msg string, err error) error {
	if err == nil {
		return nil
	}

	if msg == "" {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return fmt.Errorf("%s: %s: %w", fn, msg, err)
}

func Wrapf(fn string, err error, format string, args ...any) error {
	return WrapMsg(fn, fmt.Sprintf(format, args...), err)
}
