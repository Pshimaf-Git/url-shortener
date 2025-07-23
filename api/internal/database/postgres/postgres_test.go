package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/random"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE:
//
// BEFORE EXEC THESE COMMAND IN TETMINAL
//
// docker run --rm --name TEST-POSTGRES -e POSTGRES_PASSWORD=PASSWORD -p 5432:5432 -d 13-alpine
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
	t.Run("context_deadline_exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		time.Sleep(time.Nanosecond * 10)

		db, err := New(ctx, POSTGRES_CFG)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Nil(t, db)
	})

	t.Run("context_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cancel()
		db, err := New(ctx, POSTGRES_CFG)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, db)
	})

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
	t.Run("context_deadline_exceeded", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		time.Sleep(time.Nanosecond * 10)

		testURL := "https://example.com"
		testAlias := "example"

		err := db.SaveURL(ctx, testURL, testAlias)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("context_canceled", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testURL := "https://example.com"
		testAlias := "example"

		cancel()
		err := db.SaveURL(ctx, testURL, testAlias)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

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
	t.Run("context_deadline_exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		time.Sleep(time.Nanosecond * 10)

		testURL := "https://example.context_deadline_exceeded.com"
		testAlias := "__context_deadline_exceeded__"

		err := db.SaveURL(context.Background(), testURL, testAlias)
		require.NoError(t, err)

		url, err := db.GetURl(ctx, testAlias)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		// because context canceled
		assert.Equal(t, "", url)
	})

	t.Run("context_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testURL := "https://example.context_canceled.ru"
		testAlias := "__context_canceled__"

		err := db.SaveURL(ctx, testURL, testAlias)
		require.NoError(t, err)

		cancel()
		url, err := db.GetURl(ctx, testAlias)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)

		// because context canceled
		assert.Equal(t, "", url)
	})

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

	t.Run("context_deadline_exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		time.Sleep(time.Nanosecond * 10)

		testURL := "https://example.context_deadline_exceeded.com"
		testAlias := "context_deadline_exceeded"
		err := db.SaveURL(context.Background(), testURL, testAlias)
		require.NoError(t, err)

		n, err := db.DeleteURL(ctx, testAlias)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Equal(t, int64(0), n)
	})

	t.Run("context_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testURL := "https://example.context_canceled.com"
		testAlias := "canceled"
		err := db.SaveURL(context.Background(), testURL, testAlias)
		require.NoError(t, err)

		cancel()
		n, err := db.DeleteURL(ctx, testAlias)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, int64(0), n)
	})

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

	t.Run("context_deadline_exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		time.Sleep(time.Nanosecond * 10)

		testURL := "https://context_deadline_exceeded.com"

		alias, err := db.SaveGeneratedURl(ctx, testURL, 10, 10)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Equal(t, "", alias)
	})

	t.Run("context_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		testURL := "https://context_deadline_exceeded.com"

		cancel()
		alias, err := db.SaveGeneratedURl(ctx, testURL, 10, 10)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, "", alias)
	})

	t.Run("max_retries_for_generate", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		conflictAlias := "abc123"
		err := db.SaveURL(ctx, "https://occupied.com", conflictAlias)
		require.NoError(t, err)

		patches := gomonkey.ApplyFunc(random.StringRandV2, func(int) string {
			return conflictAlias
		})
		defer patches.Reset()

		_, err = db.SaveGeneratedURl(ctx, "https://new-url.com", len(conflictAlias), 2)

		assert.ErrorIs(t, err, database.ErrMaxRetriesForGenerate)
	})

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

func TestBuildConnString(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		localCfg := &config.PostreSQLConfig{
			Host:     "89.9.8.7",
			Port:     "9876",
			Name:     "mydb",
			User:     "root",
			SSLMode:  "enable",
			Password: "1234",
		}

		expectedString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			localCfg.User, localCfg.Password, localCfg.Host, localCfg.Port, localCfg.Name, localCfg.SSLMode,
		)

		wantCotains := []string{
			localCfg.Host,
			localCfg.Name,
			localCfg.User,
			localCfg.Port,
			localCfg.SSLMode,
			localCfg.Password,
		}

		connStr := BuildConnString(localCfg)
		for _, s := range wantCotains {
			assert.Contains(t, connStr, s)
		}

		assert.Equal(t, expectedString, connStr)
	})
}

func Test_validateSaveGeneratedURL(t *testing.T) {
	type input struct {
		url         string
		length      int
		maxAttempts int
	}
	testCases := []struct {
		name    string
		input   input
		wantErr bool
	}{
		{
			name: "good input",
			input: input{
				url:         "http://good.com",
				length:      10,
				maxAttempts: 5,
			},
			wantErr: false,
		},

		{
			name: "bad max attempts",
			input: input{
				url:         "http://good.com",
				length:      10,
				maxAttempts: -10,
			},
			wantErr: true,
		},

		{
			name: "bad length",
			input: input{
				url:         "http://good.com",
				length:      0,
				maxAttempts: 1000,
			},
			wantErr: true,
		},

		{
			name: "empty URL",
			input: input{
				url:         "",
				length:      1,
				maxAttempts: 1,
			},
			wantErr: true,
		},

		{
			name:    "bad all inputs",
			input:   input{}, // zero value: "", 0, 0
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSaveGeneratedURl(tt.input.url, tt.input.length, tt.input.maxAttempts)

			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

		})
	}
}
