package cache

import (
	"context"

	"github.com/Pshimaf-Git/url-shortener/internal/lib/errors"
)

type Setter interface {
	Set(ctx context.Context, key string, value any) error

	Close() error
}

type Getter interface {
	Get(ctx context.Context, key string) (val string, err error)

	Close() error
}

type Expirer interface {
	Expire(ctx context.Context, key string) error
}

type Cache interface {
	Setter
	Getter
	Expirer

	Close() error
}

var (
	ErrKeyNotExist = errors.New("key does not exist")
	ErrEmptyKey    = errors.New("emprty key")
)
