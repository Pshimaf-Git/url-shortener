package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE:
//
// BEFORE EXEC THESE COMMAND IN TETMINAL
//
// docker run --rm --name TEST-POSTGRES -e POSTGRES_PASSWORD=PASSWORD -p 5432:5432 -d postgres:13-alpine
//
// these command create a docker container with pistgres database for test
//
// AFTER EXEC THESE COMMAND
//
// docker stop TEST-POSTGRES && docker rm TEST-POSTGRES

const (
	testTimeout = 5 * time.Second
)

var (
	POSTGRES_CFG = &config.PostreSQLConfig{
		Host:     "127.0.0.1",
		Port:     "5432",
		User:     "postgres",
		Name:     "postgres",
		Password: "PASSWORD",
		SSLMode:  "disable",
	}
)

func setupTestDB(t *testing.T) (*storage, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	db, err := New(ctx, POSTGRES_CFG)
	require.NoError(t, err)

	cleanup := func() {
		_, err := db.pool.Exec(context.Background(), "DELETE FROM urls")
		require.NoError(t, err)
		db.Close()
	}

	return db, cleanup
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *config.PostreSQLConfig
		ctxDuration time.Duration
		wantErr     bool
		errType     error
	}{
		{
			name:        "successful connection",
			cfg:         POSTGRES_CFG,
			ctxDuration: testTimeout,
			wantErr:     false,
		},
		{
			name:        "invalid config",
			cfg:         &config.PostreSQLConfig{},
			ctxDuration: testTimeout,
			wantErr:     true,
		},
		{
			name:        "context timeout",
			cfg:         POSTGRES_CFG,
			ctxDuration: time.Nanosecond,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxDuration)
			defer cancel()

			db, err := New(ctx, tt.cfg)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, db)
				require.NoError(t, db.Close())
			}
		})
	}
}

func TestSaveURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name    string
		url     string
		alias   string
		wantErr bool
		errType error
		setup   func() // Optional setup function
	}{
		{
			name:    "save new url",
			url:     "https://example.com",
			alias:   "example",
			wantErr: false,
		},
		{
			name:    "duplicate alias",
			url:     "https://another.com",
			alias:   "dupl",
			wantErr: true,
			errType: database.ErrURLExist,
			setup: func() {
				// Pre-create the alias
				ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
				defer cancel()
				err := db.SaveURL(ctx, "https://original.com", "dupl")
				require.NoError(t, err)
			},
		},
		{
			name:    "empty url",
			url:     "",
			alias:   "empty",
			wantErr: false,
		},
		{
			name:    "empty alias",
			url:     "https://example.com",
			alias:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			if tt.setup != nil {
				tt.setup()
			}

			err := db.SaveURL(ctx, tt.url, tt.alias)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify the URL was actually saved
				if tt.alias != "" {
					url, err := db.GetURl(ctx, tt.alias)
					require.NoError(t, err)
					assert.Equal(t, tt.url, url)
				}
			}
		})
	}
}

func TestGetURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Setup test data
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	testURL := "https://example.com"
	testAlias := "example"
	err := db.SaveURL(ctx, testURL, testAlias)
	require.NoError(t, err)

	tests := []struct {
		name    string
		alias   string
		wantURL string
		wantErr bool
		errType error
	}{
		{
			name:    "get existing url",
			alias:   testAlias,
			wantURL: testURL,
			wantErr: false,
		},
		{
			name:    "non-existent alias",
			alias:   "nonexistent",
			wantErr: true,
			errType: database.ErrURLNotFound,
		},
		{
			name:    "empty alias",
			alias:   "",
			wantErr: true,
			errType: database.ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			url, err := db.GetURl(ctx, tt.alias)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantURL, url)
			}
		})
	}
}

func TestDeleteURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Setup test data
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	testURL := "https://example.com"
	testAlias := "example"
	err := db.SaveURL(ctx, testURL, testAlias)
	require.NoError(t, err)

	tests := []struct {
		name         string
		alias        string
		wantAffected int64
		wantErr      bool
		setup        func() // Optional setup function
		verify       func() // Optional verification function
	}{
		{
			name:         "delete existing url",
			alias:        testAlias,
			wantAffected: 1,
			wantErr:      false,
			verify: func() {
				_, err := db.GetURl(ctx, testAlias)
				assert.ErrorIs(t, err, database.ErrURLNotFound)
			},
		},
		{
			name:         "delete non-existent url",
			alias:        "nonexistent",
			wantAffected: 0,
			wantErr:      true,
		},
		{
			name:    "delete empty alias",
			alias:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			if tt.setup != nil {
				tt.setup()
			}

			affected, err := db.DeleteURL(ctx, tt.alias)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAffected, affected)
			}

			if tt.verify != nil {
				tt.verify()
			}
		})
	}
}

func TestSaveGeneratedURL(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tests := []struct {
		name        string
		url         string
		length      int
		maxAttempts int
		wantErr     bool
		errType     error
		setup       func() // Optional setup function
	}{
		{
			name:        "successful generation",
			url:         "https://example.com",
			length:      6,
			maxAttempts: 3,
			wantErr:     false,
		},
		{
			name:        "empty url",
			url:         "",
			length:      6,
			maxAttempts: 3,
			wantErr:     true,
		},
		{
			name:        "invalid length",
			url:         "https://example.com",
			length:      0,
			maxAttempts: 3,
			wantErr:     true,
		},
		{
			name:        "invalid attempts",
			url:         "https://example.com",
			length:      6,
			maxAttempts: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			if tt.setup != nil {
				tt.setup()
			}

			alias, err := db.SaveGeneratedURl(ctx, tt.url, tt.length, tt.maxAttempts)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Empty(t, alias)
			} else {
				assert.NoError(t, err)
				assert.Len(t, alias, tt.length)

				// Verify the URL was actually saved
				url, err := db.GetURl(ctx, alias)
				require.NoError(t, err)
				assert.Equal(t, tt.url, url)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) *storage
		wantErr bool
	}{
		{
			name: "close valid db",
			setup: func(t *testing.T) *storage {
				db, err := New(context.Background(), POSTGRES_CFG)
				require.NoError(t, err)
				return db
			},
			wantErr: false,
		},
		{
			name: "close nil pool",
			setup: func(t *testing.T) *storage {
				return &storage{pool: nil}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setup(t)
			err := db.Close()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// For valid db case, verify pool is closed
			if db.pool != nil {
				_, err = db.pool.Exec(context.Background(), "SELECT 1")
				assert.Error(t, err)
			}
		})
	}
}
