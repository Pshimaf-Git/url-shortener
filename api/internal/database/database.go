package database

import (
	"context"
	"errors"
)

type URLProvider interface {
	GetURl(ctx context.Context, alias string) (string, error)
}

type URLDeleter interface {
	DeleteURL(ctx context.Context, alias string) (int64, error)
}

type URLSaver interface {
	SaveURL(ctx context.Context, userURl string, alias string) error
	SaveGeneratedURl(ctx context.Context, originalURL string, length, maxAttempts int) (string, error)
}

type Database interface {
	URLProvider
	URLDeleter
	URLSaver

	Close() error
}

var (
	ErrURLNotFound           = errors.New("url not found")
	ErrURLExist              = errors.New("url exists")
	ErrMaxRetriesForGenerate = errors.New("max retries for generate unique alias")
)
