package random

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestInt64Crypto(t *testing.T) {
	type args struct {
		max int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "base case",
			args: args{
				max: 100,
			},
			wantErr: false,
		},

		{
			name: "negative max",
			args: args{
				max: -10000,
			},
			wantErr: true,
		},

		{
			name: "zaro max",
			args: args{
				max: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Int64Crypto(tt.args.max)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestString(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "base-case",
			args: args{
				length: 10,
			},
			wantLen: 10,
		},

		{
			name: "negative length",
			args: args{
				length: -12,
			},
			wantLen: 0,
		},

		{
			name: "zero length",
			args: args{
				length: 0,
			},
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringRandV2(tt.args.length)
			assert.Equal(t, tt.wantLen, utf8.RuneCountInString(got))
		})
	}
}
