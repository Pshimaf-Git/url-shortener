package resp

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOK(t *testing.T) {
	tests := []struct {
		name string
		want Response
	}{
		{
			name: "base case",
			want: Response{
				Status: StatusOK,
				Error:  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OK()
			assert.Equal(t, tt.want, got)
		})
	}
}

var ErrSomeError = errors.New("ERROR")

func TestError(t *testing.T) {
	tests := []struct {
		name string
		want Response
	}{
		{
			name: "base case",
			want: Response{
				Status: StatusError,
				Error:  ErrSomeError.Error(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Error(ErrSomeError)
			assert.Equal(t, tt.want, got)
		})
	}
}
