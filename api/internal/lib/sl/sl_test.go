package sl

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ErrSomething = errors.New("something error")

const errorKey = "error"

func TestError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want slog.Attr
	}{
		{
			name: "base case",
			args: args{
				err: ErrSomething,
			},
			want: slog.Attr{
				Key:   errorKey,
				Value: slog.StringValue(ErrSomething.Error()),
			},
		},

		{
			name: "nil error",
			args: args{
				err: nil,
			},
			want: slog.Attr{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Error(tt.args.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
