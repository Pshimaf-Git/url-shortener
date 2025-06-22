// Package errors provides enhanced error handling utilities that extend the standard
// errors package.
//
// It offers:
// - Standard error functions (New, Join, As, Is, Unwrap)
// - Enhanced error wrapping with context preservation
// - Consistent error formatting
// - Compatibility with standard error inspection
//
// The Wrap function follows Go's error wrapping conventions while adding contextual
// information for better error tracing and debugging.
package errors

import (
	"errors"
	"fmt"
	"strings"
)

// New returns an error that formats as the given text. Each call to New returns a
// distinct error value even if the text is identical.
func New(msg string) error { return errors.New(msg) }

// Join returns an error that wraps the given errors. Any nil error values are discarded
// Join returns nil if every value in errs is nil. The error formats as the concatenation
// of the strings obtained by calling the Error method of each element of errs, with a
// newline between each string.
//
// A non-nil error returned by Join implements the Unwrap() []error method.
func Join(errs ...error) error { return errors.Join(errs...) }

// As finds the first error in err's tree that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The tree consists of err itself, followed by the errors obtained by repeatedly calling
// its Unwrap() error or Unwrap() []error method. When err wraps multiple errors, As
// examines err followed by a depth-first traversal of its children.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(any) bool such that As(target)
// returns true. In the latter case, the As method is responsible for setting target.
//
// An error type might provide an As method so it can be treated as if it were a different
// error type.
//
// As panics if target is not a non-nil pointer to either a type that implements error, or
// to any interface type.
func As(err error, target any) bool { return errors.As(err, target) }

// Is reports whether any error in err's tree matches target.
//
// The tree consists of err itself, followed by the errors obtained by repeatedly
// calling its Unwrap() error or Unwrap() []error method. When err wraps multiple
// errors, Is examines err followed by a depth-first traversal of its children.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See [syscall.Errno.Is] for
// an example in the standard library. An Is method should only shallowly
// compare err and the target and not call [Unwrap] on either.
func Is(err error, target error) bool { return errors.Is(err, target) }

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
//
// Unwrap only calls a method of the form "Unwrap() error".
// In particular Unwrap does not unwrap errors returned by [Join].
func Unwrap(err error) error { return errors.Unwrap(err) }

// Wrap annotates err with additional context while preserving the original error.
// The message is formatted as "fn: msg: err" if msg is non-empty,
// or "fn: err" if msg is empty.
//
// The returned error implements an Unwrap method returning the original err,
// making it compatible with errors.Is and errors.As.
//
// If err is nil, Wrap returns nil.
func Wrap(fn string, msg string, err error) error {
	if err == nil {
		return nil
	}

	if strings.EqualFold(msg, "") {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return fmt.Errorf("%s: %s: %w", fn, msg, err)
}
