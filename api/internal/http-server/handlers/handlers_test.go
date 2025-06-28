package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/api/internal/cache/cachemock"
	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database/mocks"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/api/resp"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/logger/discard"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Equal(t, tt.want, isValidURL(tt.url))
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
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t1 := time.Now()
					discardLogger.Info("start", slog.Time("now", t1))
					defer discardLogger.Info("end", slog.Duration("now", time.Since(t1)))

					next.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodDelete, "/", nil))
				})
			},
		}

		router := h.InitRoutes(mws...)
		assert.NotNil(t, router)

		inUse := router.Middlewares()
		assert.Equal(t, len(mws), len(inUse))

		ctrl.Finish()
	})
}

func Test_badConfigurate(t *testing.T) {
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
			name:   "witout logger",
			db:     mocks.NewMockDatabase(ctrl),
			cache:  cachemock.NewMockCache(ctrl),
			logger: nil,
			cfg:    discardCfg,
			want:   true,
		},
		{
			name:   "witout db and cache",
			db:     nil,
			cache:  nil,
			logger: discardLogger,
			cfg:    discardCfg,
			want:   true,
		},
		{
			name:   "witout all",
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
			assert.Equal(t, tt.want, h.badConfigurate())
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

func TestNewSave(t *testing.T) {
	t.Run("new saver", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockDB := mocks.NewMockDatabase(ctrl)
		mockCache := cachemock.NewMockCache(ctrl)

		h := New(mockDB, mockCache, discardCfg, discardLogger)

		saver := h.NewSave()
		assert.NotNil(t, saver)

		ctrl.Finish()
	})
}

func TestNewDelete(t *testing.T) {
	t.Run("new delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockDB := mocks.NewMockDatabase(ctrl)
		mockCache := cachemock.NewMockCache(ctrl)

		h := New(mockDB, mockCache, discardCfg, discardLogger)

		saver := h.NewDelete()
		assert.NotNil(t, saver)

		ctrl.Finish()
	})
}

func TestNewRedirect(t *testing.T) {
	t.Run("new redirest", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockDB := mocks.NewMockDatabase(ctrl)
		mockCache := cachemock.NewMockCache(ctrl)

		h := New(mockDB, mockCache, discardCfg, discardLogger)

		saver := h.NewRedirect()
		assert.NotNil(t, saver)

		ctrl.Finish()
	})
}

func TestGetURLWithCache(t *testing.T) {
	testCases := []struct {
		name          string
		alias         string
		dbBehavior    func(m *mocks.MockDatabase, alias string)
		cacheBehavior func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup)
		errType       error
		wantErr       bool
	}{
		{
			name: "base case(cache hit)",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(0)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("http://go.com", nil)
				c.EXPECT().Expire(gomock.Any(), alias).Times(1).Return(nil)
			},
			wantErr: false,
		},

		{
			name: "base case(cache miss)",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(1).Return("http://go.com", nil)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				wg.Add(1)

				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("", cache.ErrKeyNotExist)
				c.EXPECT().Expire(gomock.Any(), alias).Times(0)

				c.EXPECT().Set(gomock.Any(), alias, gomock.Any()).
					Times(1).
					Do(func(_ context.Context, _ string, _ any) {
						defer wg.Done()
					}).
					Return(nil)
			},
			wantErr: false,
		},

		{
			name: "expire error",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(0)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("http://go.com", nil)
				c.EXPECT().Expire(gomock.Any(), alias).Times(1).Return(cache.ErrKeyNotExist)
			},
			wantErr: false,
		},

		{
			name: "get from cache error",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(1).Return("http://go.com", nil)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				wg.Add(1)

				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("", ErrInternal)
				c.EXPECT().Expire(gomock.Any(), alias).Times(0)

				c.EXPECT().Set(gomock.Any(), alias, gomock.Any()).
					Times(1).
					Do(func(_ context.Context, _ string, _ any) {
						defer wg.Done()
					}).
					Return(nil)
			},
		},

		{
			name: "url not found in db",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(1).Return("", database.ErrURLNotFound)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("", cache.ErrKeyNotExist)
			},
			wantErr: true,
			errType: database.ErrURLNotFound,
		},

		{
			name: "internal database error",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(1).Return("", ErrInternal)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("", cache.ErrKeyNotExist)
			},
			wantErr: true,
		},

		{
			name: "url not found in db but exist in cache",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(0).Return("", database.ErrURLNotFound)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("http://go.com", nil)
				c.EXPECT().Expire(gomock.Any(), alias).Times(1).Return(nil)
			},
			wantErr: false,
		},

		{
			name: "internal database error but exist in cache",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(0).Return("", ErrInternal)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("http://go.com", nil)
				c.EXPECT().Expire(gomock.Any(), alias).Times(1).Return(nil)
			},
			wantErr: false,
		},

		{
			name: "set error",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Times(1).Return("http://go.com", nil)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				wg.Add(1)

				c.EXPECT().Get(gomock.Any(), alias).Times(1).Return("", cache.ErrKeyNotExist)
				c.EXPECT().Expire(gomock.Any(), alias).Times(0)

				c.EXPECT().Set(gomock.Any(), alias, gomock.Any()).
					Times(1).
					Do(func(_ context.Context, _ string, _ any) {
						defer wg.Done()
					}).
					Return(cache.ErrKeyNotExist)
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var wg = &sync.WaitGroup{}
			ctrl := gomock.NewController(t)

			dbMock := mocks.NewMockDatabase(ctrl)
			mockCache := cachemock.NewMockCache(ctrl)

			if tt.dbBehavior != nil {
				tt.dbBehavior(dbMock, tt.alias)
			}

			if tt.cacheBehavior != nil {
				tt.cacheBehavior(mockCache, tt.alias, wg)
			}

			h := New(dbMock, mockCache, discardCfg, discardLogger)

			_, err := h.GetURLWithCache(context.Background(), tt.alias)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Error(t, err)
			}

			wg.Wait()
		})
	}
}

func TestSave(t *testing.T) {
	testCases := []struct {
		name           string
		dbBehavior     func(m *mocks.MockDatabase, req Request)
		request        Request
		expectedStatus int
	}{
		{
			name: "happy path",
			dbBehavior: func(m *mocks.MockDatabase, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, gomock.Any()).Return(nil)
			},
			request:        Request{URL: "https://www.google.com", Alias: "google"},
			expectedStatus: http.StatusCreated,
		},

		{
			name:    "invalid request body",
			request: Request{URL: "http://invalid"},
			dbBehavior: func(m *mocks.MockDatabase, req Request) {
				m.EXPECT().SaveGeneratedURl(gomock.Any(), req.URL, gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
		},

		{
			name: "duplicate alias",
			dbBehavior: func(m *mocks.MockDatabase, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, gomock.Any()).Return(database.ErrURLExist)
			},
			request:        Request{URL: "https://exmaple.com", Alias: "already_exist"},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:           "empty url",
			request:        Request{Alias: "empty URL"},
			dbBehavior:     func(m *mocks.MockDatabase, req Request) {},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:    "empty alias, happy path(SaveGeneratedURl)",
			request: Request{URL: "http://url.witout.alias"},
			dbBehavior: func(m *mocks.MockDatabase, req Request) {
				m.EXPECT().
					SaveGeneratedURl(gomock.Any(), req.URL, gomock.Any(), gomock.Any()).
					Times(1).
					Return("http://good", nil)
			},
			expectedStatus: http.StatusCreated,
		},

		{
			name:    "empty alias, max retries error(SaveGeneratedURl)",
			request: Request{URL: "http://url.witout.alias"},
			dbBehavior: func(m *mocks.MockDatabase, req Request) {
				m.EXPECT().
					SaveGeneratedURl(gomock.Any(), req.URL, gomock.Any(), gomock.Any()).
					Times(1).
					Return("", database.ErrMaxRetriesForGenerate)
			},
			expectedStatus: http.StatusInternalServerError,
		},

		{
			name:    "empty alias, interal database error(SaveGeneratedURl)",
			request: Request{URL: "http://url.witout.alias"},
			dbBehavior: func(m *mocks.MockDatabase, req Request) {
				m.EXPECT().
					SaveGeneratedURl(gomock.Any(), req.URL, gomock.Any(), gomock.Any()).
					Times(1).
					Return("", ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
		},

		{
			name:    "internal database error",
			request: Request{URL: "http://valid/url", Alias: "real valid"},
			dbBehavior: func(m *mocks.MockDatabase, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, gomock.Any()).Return(ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			dbMock := mocks.NewMockDatabase(ctrl)
			mockCache := cachemock.NewMockCache(ctrl)

			if tt.dbBehavior != nil {
				tt.dbBehavior(dbMock, tt.request)
			}

			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			if tt.name == "invalid request body" {
				body, _ = json.Marshal(fmt.Sprintf(`"url":"%s";;;;`, tt.request.URL))
			}

			r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
			w := httptest.NewRecorder()

			h := New(dbMock, mockCache, discardCfg, discardLogger)

			saver := h.NewSave()

			saver(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			ctrl.Finish()
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name           string
		alias          string
		dbBehavior     func(m *mocks.MockDatabase, alias string)
		caheBehavior   func(c *cachemock.MockCache, alias string)
		expectedStatus int
	}{
		{
			name:  "happy path",
			alias: "alias",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(hasDel, nil)
			},
			caheBehavior: func(c *cachemock.MockCache, alias string) {
				c.EXPECT().Delete(gomock.Any(), alias).
					Times(1).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},

		{
			name:  "unknown alias",
			alias: "unknown",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(noDel, database.ErrURLNotFound)
			},
			caheBehavior: func(c *cachemock.MockCache, alias string) {
				c.EXPECT().Delete(gomock.Any(), alias).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:  "internal database error",
			alias: "valid",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(noDel, ErrInternal)
			},
			caheBehavior: func(c *cachemock.MockCache, alias string) {
				c.EXPECT().Delete(gomock.Any(), alias).Times(0)
			},
			expectedStatus: http.StatusInternalServerError,
		},

		{
			name:  "internal cache error",
			alias: "alias",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(hasDel, nil)
			},
			caheBehavior: func(c *cachemock.MockCache, alias string) {
				c.EXPECT().Delete(gomock.Any(), alias).Times(1).Return(ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
		},

		{
			name:  "key not exist in cache",
			alias: "key",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(hasDel, nil)
			},
			caheBehavior: func(c *cachemock.MockCache, alias string) {
				c.EXPECT().Delete(gomock.Any(), alias).Return(cache.ErrKeyNotExist)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			dbMock := mocks.NewMockDatabase(ctrl)
			mockCache := cachemock.NewMockCache(ctrl)

			if tt.dbBehavior != nil {
				tt.dbBehavior(dbMock, tt.alias)
			}

			if tt.caheBehavior != nil {
				tt.caheBehavior(mockCache, tt.alias)
			}

			q := url.Values{}
			q.Add("alias", tt.alias)

			r := httptest.NewRequest(http.MethodDelete, path+"?"+q.Encode(), nil)
			w := httptest.NewRecorder()

			h := New(dbMock, mockCache, discardCfg, discardLogger)

			deleter := h.NewDelete()

			deleter(w, r)
			assert.Equal(t, tt.expectedStatus, w.Code)

			ctrl.Finish()
		})
	}
}

func TestRedirect(t *testing.T) {
	testCases := []struct {
		name          string
		alias         string
		dbBehavior    func(m *mocks.MockDatabase, alias string)
		cacheBehavior func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup)
		wantStatus    int
		wantRedirect  bool
	}{
		{
			name:  "happy path - cache miss, db hit",
			alias: "valid",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Return("http://something.com", nil)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				wg.Add(1)

				c.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Set(gomock.Any(), alias, "http://something.com").
					Do(func(_ context.Context, _ string, _ any) {
						defer wg.Done()
					}).
					Return(nil)
			},
			wantStatus:   http.StatusFound,
			wantRedirect: true,
		},
		{
			name:  "happy path - cache hit",
			alias: "cached",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), gomock.Any()).Times(0)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), alias).Return("http://cached.com", nil)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				c.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantStatus:   http.StatusFound,
			wantRedirect: true,
		},
		{
			name:  "empty alias",
			alias: "",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), gomock.Any()).Times(0)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				c.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantStatus:   http.StatusBadRequest,
			wantRedirect: false,
		},
		{
			name:  "url not found",
			alias: "notfound",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Return("", database.ErrURLNotFound)
			},
			cacheBehavior: func(m *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				m.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				m.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantStatus:   http.StatusNotFound,
			wantRedirect: false,
		},
		{
			name:  "database error",
			alias: "dberror",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Return("", errors.New("db error"))
			},
			cacheBehavior: func(m *cachemock.MockCache, alias string, wg *sync.WaitGroup) {
				m.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				m.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantStatus:   http.StatusInternalServerError,
			wantRedirect: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dbMock := mocks.NewMockDatabase(ctrl)
			cacheMock := cachemock.NewMockCache(ctrl)
			var wg = &sync.WaitGroup{}

			if tt.dbBehavior != nil {
				tt.dbBehavior(dbMock, tt.alias)
			}

			if tt.cacheBehavior != nil {
				tt.cacheBehavior(cacheMock, tt.alias, wg)
			}

			q := url.Values{}
			if tt.alias != "" {
				q.Add("alias", tt.alias)
			}

			h := New(dbMock, cacheMock, discardCfg, discardLogger)

			r := httptest.NewRequest(http.MethodGet, path+"?"+q.Encode(), nil)
			w := httptest.NewRecorder()

			redirector := h.NewRedirect()
			redirector(w, r)

			wg.Wait()

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantRedirect, wasRedirect(w))
		})
	}
}

func Test_helthy(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()

		ctrl := gomock.NewController(t)

		dbMock := mocks.NewMockDatabase(ctrl)
		cacheMock := cachemock.NewMockCache(ctrl)

		h := New(dbMock, cacheMock, discardCfg, discardLogger)

		h.helthy(w, r)

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

		h.helthy(w, r)

		body, err := io.ReadAll(w.Result().Body)
		require.NoError(t, err, "read responce body")

		var responce resp.Response
		err = json.Unmarshal(body, &responce)
		require.NoError(t, err, "json unmarshaling")

		assert.Equal(t, resp.Error(ErrInternalServer), responce)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func wasRedirect(w *httptest.ResponseRecorder) bool {
	return w.Code == http.StatusFound
}

func TestHandlers_SaveRedirectDelete_HappyPath(t *testing.T) {
	t.Run("Save_Redirect_Delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		dbMock := mocks.NewMockDatabase(ctrl)
		cacheMock := cachemock.NewMockCache(ctrl)

		request := Request{URL: "http://www.google.com", Alias: "google"}
		h := New(dbMock, cacheMock, discardCfg, discardLogger)

		{
			body, err := json.Marshal(request)
			require.NoError(t, err, "json.Marshal")

			r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
			w := httptest.NewRecorder()

			saver := h.NewSave()

			dbMock.EXPECT().SaveURL(gomock.Any(), request.URL, gomock.Any()).Times(1).Return(nil)

			saver(w, r)

			assert.Equal(t, w.Code, http.StatusCreated, "Saver")

		}

		{
			q := url.Values{}
			q.Add("alias", request.Alias)

			r := httptest.NewRequest(http.MethodPost, path+"?"+q.Encode(), nil)
			w := httptest.NewRecorder()

			redirector := h.NewRedirect()

			cacheMock.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return(request.URL, nil)
			cacheMock.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			cacheMock.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			dbMock.EXPECT().GetURl(gomock.Any(), gomock.Any()).Times(0)

			redirector(w, r)

			assert.Equal(t, true, wasRedirect(w), "Redirect")
		}

		{

			q := url.Values{}
			q.Add("alias", request.Alias)

			r := httptest.NewRequest(http.MethodPost, path+"?"+q.Encode(), nil)
			w := httptest.NewRecorder()

			deleter := h.NewDelete()

			dbMock.EXPECT().DeleteURL(gomock.Any(), gomock.Any()).Times(1).Return(hasDel, nil)
			cacheMock.EXPECT().Delete(gomock.Any(), gomock.Any()).
				Times(1).
				Return(nil)

			deleter(w, r)

			assert.Equal(t, w.Code, http.StatusOK)
		}

		ctrl.Finish()
	})
}

func TestHandlers_SaveRedirectDelete_BadRequests(t *testing.T) {
	testCases := []struct {
		name                   string
		request                Request // For save
		alias                  string  // For redirect/delete
		saveBehavior           func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request)
		redirectBehavior       func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string)
		deleteBehavior         func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string)
		expectedSaveStatus     int
		expectedRedirectStatus int
		expectedDeleteStatus   int
	}{
		{
			name:    "empty URL",
			request: Request{Alias: "empty_url"},
			alias:   "empty_url",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				// No expectations - should fail before DB call
			},
			redirectBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				// Shouldn't be called in this test case
			},
			deleteBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				// Shouldn't be called in this test case
			},
			expectedSaveStatus: http.StatusBadRequest,
		},
		{
			name:    "invalid URL format",
			request: Request{URL: "invalid-url", Alias: "bad_format"},
			alias:   "bad_format",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				// No expectations - should fail validation
			},
			expectedSaveStatus: http.StatusBadRequest,
		},
		{
			name:    "empty alias for redirect",
			request: Request{URL: "http://valid.com", Alias: "valid"},
			alias:   "",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(nil)
			},
			redirectBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				// No expectations - should fail before DB/cache call
			},
			expectedSaveStatus:     http.StatusCreated,
			expectedRedirectStatus: http.StatusBadRequest,
		},
		{
			name:    "non-existent alias for redirect",
			request: Request{URL: "http://valid.com", Alias: "will_not_exist"},
			alias:   "will_not_exist",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				// Don't save so it won't exist for redirect
			},
			redirectBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				c.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(0)
				m.EXPECT().GetURl(gomock.Any(), alias).Return("", database.ErrURLNotFound)
			},
			expectedRedirectStatus: http.StatusNotFound,
		},
		{
			name:    "empty alias for delete",
			request: Request{URL: "http://valid.com", Alias: "valid"},
			alias:   "",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(nil)
			},
			deleteBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				// No expectations - should fail before DB call
			},
			expectedSaveStatus:   http.StatusCreated,
			expectedDeleteStatus: http.StatusBadRequest,
		},
		{
			name:    "database error during save",
			request: Request{URL: "http://valid.com", Alias: "db_error"},
			alias:   "db_error",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(ErrInternal)
			},
			expectedSaveStatus: http.StatusInternalServerError,
		},
		{
			name:    "database error during redirect",
			request: Request{URL: "http://valid.com", Alias: "db_error_redirect"},
			alias:   "db_error_redirect",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(nil)
			},
			redirectBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				c.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(0)
				m.EXPECT().GetURl(gomock.Any(), alias).Return("", ErrInternal)
			},
			expectedSaveStatus:     http.StatusCreated,
			expectedRedirectStatus: http.StatusInternalServerError,
		},
		{
			name:    "database error during delete",
			request: Request{URL: "http://valid.com", Alias: "db_error_delete"},
			alias:   "db_error_delete",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(nil)
			},
			deleteBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(noDel, ErrInternal)
			},
			expectedSaveStatus:   http.StatusCreated,
			expectedDeleteStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dbMock := mocks.NewMockDatabase(ctrl)
			cacheMock := cachemock.NewMockCache(ctrl)

			h := New(dbMock, cacheMock, discardCfg, discardLogger)

			// Test Save if expected
			if tt.expectedSaveStatus != 0 {
				body, err := json.Marshal(tt.request)
				require.NoError(t, err)

				r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
				w := httptest.NewRecorder()

				if tt.saveBehavior != nil {
					tt.saveBehavior(dbMock, cacheMock, tt.request)
				}

				saver := h.NewSave()
				saver(w, r)
				assert.Equal(t, tt.expectedSaveStatus, w.Code)
			}

			// Test Redirect if expected
			if tt.expectedRedirectStatus != 0 {
				q := url.Values{}
				if tt.alias != "" {
					q.Add("alias", tt.alias)
				}

				r := httptest.NewRequest(http.MethodGet, path+"?"+q.Encode(), nil)
				w := httptest.NewRecorder()

				if tt.redirectBehavior != nil {
					tt.redirectBehavior(dbMock, cacheMock, tt.alias)
				}

				redirector := h.NewRedirect()
				redirector(w, r)
				assert.Equal(t, tt.expectedRedirectStatus, w.Code)
			}

			// Test Delete if expected
			if tt.expectedDeleteStatus != 0 {
				q := url.Values{}
				if tt.alias != "" {
					q.Add("alias", tt.alias)
				}

				r := httptest.NewRequest(http.MethodDelete, path+"?"+q.Encode(), nil)
				w := httptest.NewRecorder()

				if tt.deleteBehavior != nil {
					tt.deleteBehavior(dbMock, cacheMock, tt.alias)
				}

				deleter := h.NewDelete()
				deleter(w, r)
				assert.Equal(t, tt.expectedDeleteStatus, w.Code)
			}
		})
	}
}
