package sl

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/wraper"
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

		{
			name: "wrapped error",
			args: args{
				err: fmt.Errorf("%s: %w: %w: blabla", "MyFunc", ErrSomething, assert.AnError),
			},
			want: slog.Attr{
				Key: errorKey,
				Value: slog.StringValue(
					fmt.Sprintf("MyFunc: %s: %s: blabla",
						ErrSomething.Error(), assert.AnError.Error()),
				),
			},
		},

		{
			name: "wrapped error with package wraper",
			args: args{
				err: wraper.WrapMsg("MyFunc", "message", ErrSomething),
			},
			want: slog.Attr{
				Key: errorKey,
				Value: slog.StringValue(
					fmt.Sprintf("MyFunc: message: %s", ErrSomething.Error()),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Error(tt.args.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
