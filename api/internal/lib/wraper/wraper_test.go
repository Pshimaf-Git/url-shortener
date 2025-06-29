package wraper

import (
	"errors"
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

func TestWraper_WrapN(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		errs     []error
		want     string
		wantNil  bool
	}{
		{
			name:     "no errors",
			funcName: "TestFunc",
			errs:     []error{},
			wantNil:  true,
		},
		{
			name:     "all nil errors",
			funcName: "TestFunc",
			errs:     []error{nil, nil, nil},
			wantNil:  true,
		},
		{
			name:     "single error",
			funcName: "TestFunc",
			errs:     []error{errors.New("first error")},
			want:     "TestFunc: first error",
		},
		{
			name:     "multiple errors",
			funcName: "ProcessData",
			errs: []error{
				errors.New("db connection failed"),
				errors.New("query failed"),
				errors.New("marshal failed"),
			},
			want: "ProcessData: db connection failed: query failed: marshal failed",
		},
		{
			name:     "mixed nil and non-nil errors",
			funcName: "ValidateInput",
			errs: []error{
				nil,
				errors.New("invalid email"),
				nil,
				errors.New("missing name"),
				nil,
			},
			want: "ValidateInput: invalid email: missing name",
		},
		{
			name:     "wrapped errors",
			funcName: "ComplexOperation",
			errs: []error{
				errors.New("initial error"),
				Wrap("SubOperation", errors.New("sub error")),
				errors.New("final error"),
			},
			want: "ComplexOperation: initial error: SubOperation: sub error: final error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := New(tt.funcName)
			got := wp.WrapN(tt.errs...)

			if tt.wantNil {
				assert.Nil(t, got, "should return nil")
				return
			}

			if !assert.NotNil(t, got, "should return error") {
				return
			}

			assert.Equal(t, tt.want, got.Error(), "error message mismatch")
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

func TestWraperWrapN(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		errs     []error
		want     string
		wantNil  bool
	}{
		{
			name:     "no errors",
			funcName: "TestFunc",
			errs:     []error{},
			wantNil:  true,
		},
		{
			name:     "all nil errors",
			funcName: "TestFunc",
			errs:     []error{nil, nil, nil},
			wantNil:  true,
		},
		{
			name:     "single error",
			funcName: "TestFunc",
			errs:     []error{errors.New("first error")},
			want:     "TestFunc: first error",
		},
		{
			name:     "multiple errors",
			funcName: "ProcessData",
			errs: []error{
				errors.New("db connection failed"),
				errors.New("query failed"),
				errors.New("marshal failed"),
			},
			want: "ProcessData: db connection failed: query failed: marshal failed",
		},
		{
			name:     "mixed nil and non-nil errors",
			funcName: "ValidateInput",
			errs: []error{
				nil,
				errors.New("invalid email"),
				nil,
				errors.New("missing name"),
				nil,
			},
			want: "ValidateInput: invalid email: missing name",
		},
		{
			name:     "wrapped errors",
			funcName: "ComplexOperation",
			errs: []error{
				errors.New("initial error"),
				Wrap("SubOperation", errors.New("sub error")),
				errors.New("final error"),
			},
			want: "ComplexOperation: initial error: SubOperation: sub error: final error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapN(tt.funcName, tt.errs...)

			if tt.wantNil {
				assert.Nil(t, got, "should return nil")
				return
			}

			if !assert.NotNil(t, got, "should return error") {
				return
			}

			assert.Equal(t, tt.want, got.Error(), "error message mismatch")
		})
	}
}
