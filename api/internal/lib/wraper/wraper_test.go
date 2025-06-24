package wraper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWraper_WrapMsg(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
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

			// Test standalone functions
			err = WrapMsg(tt.funcName, tt.msg, tt.err)
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

func TestWraper_Wrap(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
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

			// Test standalone function
			err = Wrap(tt.fn, tt.err)
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

func TestWrapf(t *testing.T) {
	var ErrOriginal = errors.New("original error")

	t.Run("formats message correctly", func(t *testing.T) {
		err := Wrapf("TestFunc", ErrOriginal, "formatted %d %s", 42, "message")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "TestFunc")
		assert.Contains(t, err.Error(), "formatted 42 message")
		assert.Contains(t, err.Error(), "original error")
		assert.ErrorIs(t, err, ErrOriginal)
	})

	t.Run("nil error", func(t *testing.T) {
		err := Wrapf("TestFunc", nil, "formatted %v", "hello world")
		assert.NoError(t, err)
	})
}
