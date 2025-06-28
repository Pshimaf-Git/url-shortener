package random

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInt64Crypto(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		validate func(int64) bool
		wantErr  bool
		errType  error
	}{
		{
			name:  "zero max",
			input: 0,
			validate: func(result int64) bool {
				return result == 0
			},
			wantErr: true,
			errType: ErrInvalidMax,
		},
		{
			name:  "negative max",
			input: -100,
			validate: func(result int64) bool {
				return result == 0
			},
			wantErr: true,
			errType: ErrInvalidMax,
		},
		{
			name:  "small positive max",
			input: 1,
			validate: func(result int64) bool {
				return result >= 0 && result < 1
			},
		},
		{
			name:  "medium positive max",
			input: 100,
			validate: func(result int64) bool {
				return result >= 0 && result < 100
			},
		},
		{
			name:  "large positive max",
			input: 1_000_000_000,
			validate: func(result int64) bool {
				return result >= 0 && result < 1_000_000_000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				result, err := Int64Crypto(tt.input)
				assert.True(t, tt.validate(result), "result %d failed validation for input %d", result, tt.input)

				if !tt.wantErr {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}

				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			}
		})
	}
}

func TestInt64RandV2(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		validate func(int64) bool
	}{
		{
			name:  "zero max",
			input: 0,
			validate: func(result int64) bool {
				return result == 0
			},
		},
		{
			name:  "negative max",
			input: -100,
			validate: func(result int64) bool {
				return result == 0
			},
		},
		{
			name:  "small positive max",
			input: 1,
			validate: func(result int64) bool {
				return result >= 0 && result < 1
			},
		},
		{
			name:  "medium positive max",
			input: 100,
			validate: func(result int64) bool {
				return result >= 0 && result < 100
			},
		},
		{
			name:  "large positive max",
			input: 1_000_000_000,
			validate: func(result int64) bool {
				return result >= 0 && result < 1_000_000_000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				result := Int64RandV2(tt.input)
				assert.True(t, tt.validate(result), "result %d failed validation for input %d", result, tt.input)
			}
		})
	}
}

func TestStringRandV2(t *testing.T) {
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

func TestStringCrypto(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
		errType error
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
			wantErr: true,
			errType: ErrInvalidMax,
		},

		{
			name: "zero length",
			args: args{
				length: 0,
			},
			wantLen: 0,
			wantErr: true,
			errType: ErrInvalidMax,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringCrypto(tt.args.length)

			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.wantLen, utf8.RuneCountInString(got))

			if tt.errType != nil {
				assert.ErrorIs(t, err, tt.errType)
			}
		})
	}
}

func TestMustStringCrypto(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectPanic bool
		validate    func(string) bool
	}{
		{
			name:        "zero length",
			length:      0,
			expectPanic: true,
		},
		{
			name:        "negative length",
			length:      -5,
			expectPanic: true,
		},
		{
			name:        "valid small length",
			length:      5,
			expectPanic: false,
			validate: func(s string) bool {
				return utf8.RuneCountInString(s) == 5
			},
		},
		{
			name:        "valid medium length",
			length:      32,
			expectPanic: false,
			validate: func(s string) bool {
				return utf8.RuneCountInString(s) == 32
			},
		},
		{
			name:        "valid large length",
			length:      1024,
			expectPanic: false,
			validate: func(s string) bool {
				return utf8.RuneCountInString(s) == 1024
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					MustStringCrypto(tt.length)
				}, "expected panic for length %d", tt.length)
				return
			}

			var result string
			require.NotPanics(t, func() {
				result = MustStringCrypto(tt.length)
			}, "unexpected panic for length %d", tt.length)

			assert.True(t, tt.validate(result), "invalid result for length %d", tt.length)

			if tt.length > 0 {
				for _, r := range result {
					assert.Contains(t, chars, r, "generated string contains invalid character")
				}
			}
		})
	}
}

func BenchmarkMustStringCrypto(b *testing.B) {
	lengths := []int{8, 32, 256}
	for _, length := range lengths {
		b.Run(fmt.Sprintf("length-%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = MustStringCrypto(length)
			}
		})
	}
}
