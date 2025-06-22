package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	type args struct {
		fn  string
		msg string
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base case",
			args: args{
				fn:  "TestWrap",
				msg: "some message",
				err: errors.New("error"),
			},
			wantErr: true,
		},

		{
			name: "nil error",
			args: args{
				fn:  "TestWrap",
				msg: "nil err",
				err: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.args.fn, tt.args.msg, tt.args.err)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
