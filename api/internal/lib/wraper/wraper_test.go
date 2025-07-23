package wraper

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		const fn = "TestFunc"
		wp := New(fn)
		assert.NotNil(t, wp)
		assert.Equal(t, wp.FuncName, fn)
	})
}

func TestWraper_Wrap(t *testing.T) {
	testCases := []struct {
		name     string
		fn       string
		err      error
		wantErr  bool
		wantCont string
	}{
		{
			name:    "nil error returns nil",
			fn:      "TestFunc",
			err:     nil,
			wantErr: false,
		},
		{
			name:     "wraps error correctly",
			fn:       "TestFunc",
			err:      errors.New("original error"),
			wantErr:  true,
			wantCont: "TestFunc: original error",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Test Wraper struct method
			wp := New(tt.fn)
			err := wp.Wrap(tt.err)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCont, err.Error())
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}
}

func TestWraper_WrapMsg(t *testing.T) {
	testCases := []struct {
		name        string
		funcName    string
		msg         string
		err         error
		wantErr     bool
		wantContain []string // expected substrings in error message
	}{
		{
			name:     "nil error returns nil",
			funcName: "TestFunc",
			msg:      "some message",
			err:      nil,
			wantErr:  false,
		},
		{
			name:        "with message wraps correctly",
			funcName:    "TestFunc",
			msg:         "something failed",
			err:         errors.New("original error"),
			wantErr:     true,
			wantContain: []string{"TestFunc", "something failed", "original error"},
		},
		{
			name:        "empty message omits message part",
			funcName:    "TestFunc",
			msg:         "",
			err:         errors.New("original error"),
			wantErr:     true,
			wantContain: []string{"TestFunc", "original error"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			wp := New(tt.funcName)

			err := wp.WrapMsg(tt.msg, tt.err)
			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			for _, s := range tt.wantContain {
				assert.Contains(t, err.Error(), s)
			}
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestWraper_Wrapf(t *testing.T) {
	type input struct {
		err    error
		format string
		args   []any
	}
	testCases := []struct {
		name     string
		fn       string
		input    input
		wantErr  bool
		wantCont string
	}{
		{
			name: "base case",
			fn:   "TestFunc",
			input: input{
				err:    errors.New("error"),
				format: "format -> %s, %d",
				args:   []any{"hello", 200},
			},
			wantErr:  true,
			wantCont: "TestFunc: format -> hello, 200: error",
		},

		{
			name: "nil input error",
			fn:   "TestFunc",
			input: input{
				err:    nil,
				format: "%s",
				args:   []any{"<nil>"},
			},
			wantErr:  false,
			wantCont: "",
		},

		{
			name: "empty format string",
			fn:   "TestFunc",
			input: input{
				err:    errors.New("error"),
				format: "",
				args:   []any{},
			},
			wantErr:  true,
			wantCont: "TestFunc: error",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			wp := New(tt.fn)
			err := wp.Wrapf(tt.input.err, tt.input.format, tt.input.args...)

			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCont, err.Error())
				assert.ErrorIs(t, err, tt.input.err)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	testCases := []struct {
		name     string
		fn       string
		err      error
		wantErr  bool
		wantCont string
	}{
		{
			name:    "nil error returns nil",
			fn:      "TestFunc",
			err:     nil,
			wantErr: false,
		},
		{
			name:     "wraps error correctly",
			fn:       "TestFunc",
			err:      errors.New("original error"),
			wantErr:  true,
			wantCont: "TestFunc: original error",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.fn, tt.err)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCont, err.Error())
				assert.ErrorIs(t, err, tt.err)
			}
		})
	}
}

func TestWrapMsg(t *testing.T) {
	testCases := []struct {
		name        string
		funcName    string
		msg         string
		err         error
		wantErr     bool
		wantContain []string // expected substrings in error message
	}{
		{
			name:     "nil error returns nil",
			funcName: "TestFunc",
			msg:      "some message",
			err:      nil,
			wantErr:  false,
		},
		{
			name:        "with message wraps correctly",
			funcName:    "TestFunc",
			msg:         "something failed",
			err:         errors.New("original error"),
			wantErr:     true,
			wantContain: []string{"TestFunc", "something failed", "original error"},
		},
		{
			name:        "empty message omits message part",
			funcName:    "TestFunc",
			msg:         "",
			err:         errors.New("original error"),
			wantErr:     true,
			wantContain: []string{"TestFunc", "original error"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapMsg(tt.funcName, tt.msg, tt.err)
			if !tt.wantErr {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			for _, s := range tt.wantContain {
				assert.Contains(t, err.Error(), s)
			}
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestWrapf(t *testing.T) {
	type input struct {
		fn     string
		err    error
		format string
		args   []any
	}
	testCases := []struct {
		name     string
		input    input
		wantErr  bool
		wantCont string
	}{
		{
			name: "base case",
			input: input{
				fn:     "TestFunc",
				err:    errors.New("error"),
				format: "format -> %s, %d",
				args:   []any{"hello", 200},
			},
			wantErr:  true,
			wantCont: "TestFunc: format -> hello, 200: error",
		},

		{
			name: "nil input error",
			input: input{
				fn:     "TestFunc",
				err:    nil,
				format: "%s",
				args:   []any{"<nil>"},
			},
			wantErr:  false,
			wantCont: "",
		},

		{
			name: "empty format string",
			input: input{
				fn:     "TestFunc",
				err:    errors.New("error"),
				format: "",
				args:   []any{},
			},
			wantErr:  true,
			wantCont: "TestFunc: error",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrapf(tt.input.fn, tt.input.err, tt.input.format, tt.input.args...)

			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantCont, err.Error())
				assert.ErrorIs(t, err, tt.input.err)
			}
		})
	}
}

func Test_newError(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		const fn = "TestFunc"

		var msg = "msg"
		var originalError error = errors.New("org")

		err := newError(fn, msg, originalError)
		require.Error(t, err)

		e, ok := err.(*Error)
		require.True(t, ok)

		assert.Equal(t, fn, e.Fn)
		assert.Equal(t, msg, e.Msg)
		assert.ErrorIs(t, e.Err, originalError)

		assert.ErrorIs(t, e, originalError)
	})

	t.Run("without_msg", func(t *testing.T) {
		const fn = "TestFunc"

		var originalError error = errors.New("org")

		err := newError(fn, emptyMsg, originalError)
		require.Error(t, err)

		e, ok := err.(*Error)
		require.True(t, ok)

		assert.Equal(t, fn, e.Fn)
		assert.Equal(t, emptyMsg, e.Msg)
		assert.ErrorIs(t, e.Err, originalError)

		assert.ErrorIs(t, e, originalError)
	})

	t.Run("nil_error", func(t *testing.T) {
		err := newError("fn", "msg", nil)
		assert.NoError(t, err)
	})
}

func TestError(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		const fn = "TestFunc"

		var msg = "msg"
		var originalError error = errors.New("org")

		err := newError(fn, msg, originalError)
		require.Error(t, err)

		wantContains := []string{fn, msg, originalError.Error(), ":"}

		for _, s := range wantContains {
			assert.Contains(t, err.Error(), s)
		}

		want := fmt.Sprintf("%s: %s: %s", fn, msg, originalError.Error())
		assert.Equal(t, want, err.Error())
	})

	t.Run("without_msg", func(t *testing.T) {
		const fn = "TestFunc"

		var originalError error = errors.New("org")

		err := newError(fn, emptyMsg, originalError)
		require.Error(t, err)

		wantContains := []string{fn, emptyMsg, originalError.Error(), ":"}

		for _, s := range wantContains {
			assert.Contains(t, err.Error(), s)
		}

		want := fmt.Sprintf("%s: %s", fn, originalError.Error())
		assert.Equal(t, want, err.Error())
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		const fn = "TestFunc"
		var originalError error = errors.New("org")

		err := newError(fn, emptyMsg, originalError)
		require.Error(t, err)

		e, ok := err.(*Error)
		require.True(t, ok)

		assert.ErrorIs(t, e.Unwrap(), originalError)
	})
}

func Test_isNil(t *testing.T) {
	t.Run("not_nil", func(t *testing.T) {
		assert.False(t, isNil(assert.AnError))
	})

	t.Run("nil", func(t *testing.T) {
		assert.True(t, isNil(nil))
	})
}

func Test_isEmptyMsg(t *testing.T) {
	t.Run("non_empty", func(t *testing.T) {
		assert.False(t, isEmptyMsg("not empty"))
	})

	t.Run("empty", func(t *testing.T) {
		assert.True(t, isEmptyMsg(emptyMsg))
	})
}

func TestIsWraped(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		const fn = "Test"
		msg := "message"
		wp := New(fn)
		werr := wp.WrapMsg(msg, assert.AnError)

		e := IsWraped(werr)
		require.NotNil(t, e)

		assert.ErrorIs(t, werr, assert.AnError)
		assert.ErrorIs(t, e, assert.AnError)

		assert.ErrorIs(t, assert.AnError, e.Err)
		assert.ErrorIs(t, assert.AnError, e.Unwrap())

		assert.Contains(t, e.Error(), fn)
		assert.Contains(t, e.Error(), msg)
		assert.Contains(t, e.Error(), assert.AnError.Error())

		assert.Equal(t, assert.AnError.Error(), e.Unwrap().Error())
		assert.Equal(t, assert.AnError.Error(), e.Err.Error())

		assert.Equal(t, msg, e.Msg)
		assert.Equal(t, fn, e.Fn)

		assert.Equal(t, fmt.Sprintf("%s: %s: %s", fn, msg, assert.AnError.Error()), e.Error())
	})

	t.Run("nil_error", func(t *testing.T) {
		werr := IsWraped(New("Test").Wrap(nil))
		assert.Nil(t, werr)
	})

	t.Run("bad_type", func(t *testing.T) {
		var pathErr error = &os.PathError{
			Op:   "MyFunc",
			Path: "D:/usr/go/src/os",
			Err:  assert.AnError,
		}

		var syscallErr error = &os.SyscallError{
			Syscall: "syscall",
			Err:     pathErr,
		}

		err := fmt.Errorf("%w: %w", pathErr, syscallErr)

		e := IsWraped(err)
		assert.Nil(t, e)
	})

	t.Run("long_error_wrap", func(t *testing.T) {
		baseFN := "fn"
		baseErr := assert.AnError
		baseWrapedErr := Wrap(baseFN, baseErr)

		var pathErr error = &os.PathError{
			Op:   "MyFunc",
			Path: "D:/usr/go/src/os",
			Err:  baseWrapedErr,
		}

		var syscallErr error = &os.SyscallError{
			Syscall: "syscall",
			Err:     pathErr,
		}

		var linkErr error = &os.LinkError{
			Op:  "MyOp",
			Old: "old",
			New: "new",
			Err: syscallErr,
		}

		var ddlErr error = &syscall.DLLError{
			Err:     linkErr,
			ObjName: "obj",
			Msg:     emptyMsg,
		}

		fmtErr := fmt.Errorf("%w: some string", fmt.Errorf("%w: %w: %w: %w", os.ErrClosed, bufio.ErrBufferFull, assert.AnError, ddlErr))

		e := IsWraped(fmtErr)
		require.NotNil(t, e)
		require.NotNil(t, e.Err)

		assert.ErrorIs(t, e, baseErr)

		assert.Equal(t, baseErr.Error(), e.Err.Error())
		assert.Equal(t, baseFN, e.Fn)
	})

	t.Run("round", func(t *testing.T) {
		first := &os.PathError{}
		second := &os.SyscallError{}

		first.Err = second
		second.Err = first

		e := IsWraped(second)
		assert.Nil(t, e)
	})
}
