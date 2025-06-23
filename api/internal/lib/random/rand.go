package random

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	v2 "math/rand/v2"

	"github.com/Pshimaf-Git/url-shortener/internal/lib/wraper"
)

var InvalidMax = errors.New("negative or zero max")

// all symbols from english alphabet with numbers
var chars = []rune("DdNOegJKLMfPQabvRSTEFwc56hijZqrsUnV789WXYAC2tklGHImoBp10uxyz34")

// StringRandV2 returns a random string with given length (using a Int64RandomV2 for
// generate random index for char)
func StringRandV2(length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]rune, length)
	for i := range result {
		n := Int64RandV2(int64(length))
		result[i] = chars[n]
	}

	return string(result)
}

// StringCrypto returns a random string with given length (using a Int64Crypto for
// generate random index for char)
func StringCrypto(length int) (string, error) {
	if length <= 0 {
		return "", nil
	}

	result := make([]rune, length)
	for i := range result {
		n, err := Int64Crypto(int64(length))
		if err != nil {
			return "", wraper.Wrapf("StringCrypto", err, "length: %d", length)
		}

		result[i] = chars[n]
	}

	return string(result), nil
}

// MustStringCrypto usong StringCrypto for generate string but it panic if err don't qual nil
func MustStringCrypto(length int) string {
	s, err := StringCrypto(length)
	if err != nil {
		panic(fmt.Sprintf("\\MustStringCrypto: %s\\", err.Error()))
	}

	return s
}

// Int64Crypto return cryptographically strong random number (using crypto/rand).
func Int64Crypto(max int64) (int64, error) {
	if max <= 0 {
		return 0, wraper.Wrap("Int64Crypto", InvalidMax)
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return -1, wraper.Wrap("Int64Crypto", err)
	}

	return nBig.Int64(), nil
}

// Int64RandV2 return non cryptographically strong pseudo
// random number (using math/rand/v2).
func Int64RandV2(max int64) int64 {
	if max <= 0 {
		return 0
	}

	return v2.Int64N(max)
}
