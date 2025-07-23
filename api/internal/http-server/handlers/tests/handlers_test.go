package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/api/internal/cache/cachemock"
	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database/mocks"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/logger/discard"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/Pshimaf-Git/url-shortener/api/internal/http-server/handlers"
)

const path = "/"

const (
	noDel = int64(iota)
	hasDel
)

var (
	ErrInternal = errors.New("internal error")
)

var (
	discardLogger = discard.NewDiscardLogger()
	discardCfg    = &config.ServerConfig{}
)

func Test_isValidURL(t *testing.T) {
	testCases := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "base case",
			url:  "http://www.example.com",
			want: true,
		},

		{
			name: "empty url",
			url:  "",
			want: false,
		},

		{
			name: "invalid url format",
			url:  "invalid://http",
			want: false,
		},

		{
			name: "unkown schema",
			url:  "unkown://www.example.com",
			want: false,
		},

		{
			name: "invalid control character",
			url:  "%%%s&@&$$$~~~````/\\\\(^__6__^)###№№№№№!!\"\"'\r'\b\a",
			want: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidURL(tt.url))
		})
	}
}

func TestInitRoutes(t *testing.T) {
	t.Run("base_case", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockDB := mocks.NewMockDatabase(ctrl)
		mockCache := cachemock.NewMockCache(ctrl)

		h := New(mockDB, mockCache, discardCfg, discardLogger)
		require.NotNil(t, h)

		mws := []func(http.Handler) http.Handler{
			middleware.Recoverer,
			middleware.RequestID,
			middleware.RealIP,
			middleware.Logger,
			httprate.LimitByRealIP(100, time.Minute),
		}

		router := h.InitRoutes(mws...)
		assert.NotNil(t, router)

		inUse := router.Middlewares()
		assert.Equal(t, len(mws), len(inUse))

		ctrl.Finish()
	})
}

func TestBadConfigurate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name   string
		db     database.Database
		cache  cache.Cache
		logger *slog.Logger
		cfg    *config.ServerConfig
		want   bool
	}{
		{
			name:   "happy path",
			db:     mocks.NewMockDatabase(ctrl),
			cache:  cachemock.NewMockCache(ctrl),
			logger: discardLogger,
			cfg:    discardCfg,
			want:   false,
		},
		{
			name:   "without logger",
			db:     mocks.NewMockDatabase(ctrl),
			cache:  cachemock.NewMockCache(ctrl),
			logger: nil,
			cfg:    discardCfg,
			want:   true,
		},
		{
			name:   "without db and cache",
			db:     nil,
			cache:  nil,
			logger: discardLogger,
			cfg:    discardCfg,
			want:   true,
		},
		{
			name:   "without all",
			db:     nil,
			cache:  nil,
			logger: nil,
			cfg:    nil,
			want:   true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.db, tt.cache, tt.cfg, tt.logger)
			assert.Equal(t, tt.want, h.BadConfigurate())
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("new handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockDB := mocks.NewMockDatabase(ctrl)
		mockCache := cachemock.NewMockCache(ctrl)

		h := New(mockDB, mockCache, discardCfg, discardLogger)
		assert.NotNil(t, h)

		ctrl.Finish()
	})
}

func Test_helthy(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()

		ctrl := gomock.NewController(t)

		dbMock := mocks.NewMockDatabase(ctrl)
		cacheMock := cachemock.NewMockCache(ctrl)

		h := New(dbMock, cacheMock, discardCfg, discardLogger)

		h.Helthy(w, r)

		body, err := io.ReadAll(w.Result().Body)
		require.NoError(t, err)

		var responce resp.Response
		err = json.Unmarshal(body, &responce)

		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, resp.OK(), responce)

		ctrl.Finish()
	})

	t.Run("bad_configurate", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()

		h := New(nil, nil, discardCfg, discardLogger)

		h.Helthy(w, r)

		body, err := io.ReadAll(w.Result().Body)
		require.NoError(t, err, "read responce body")

		var responce resp.Response
		err = json.Unmarshal(body, &responce)
		require.NoError(t, err, "json unmarshaling")

		assert.Equal(t, resp.Error(ErrInternalServer), responce)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
