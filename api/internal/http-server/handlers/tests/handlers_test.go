package handlers_tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Pshimaf-Git/url-shortener/internal/cache"
	"github.com/Pshimaf-Git/url-shortener/internal/cache/cachemock"
	"github.com/Pshimaf-Git/url-shortener/internal/config"
	"github.com/Pshimaf-Git/url-shortener/internal/database"
	"github.com/Pshimaf-Git/url-shortener/internal/database/mocks"
	"github.com/Pshimaf-Git/url-shortener/internal/http-server/handlers"
	"github.com/Pshimaf-Git/url-shortener/internal/lib/logger/discard"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const path = "/"

const (
	noDel = int64(iota)
	hasDel
)

var discardLogger = discard.NewDiscardLogger()

var (
	ErrInternalDatabase = errors.New("internal database error")
)

var (
	discardCfg = &config.Config{}
)

func TestNewHandler(t *testing.T) {
	t.Run("new handler", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockDB := mocks.NewMockDatabase(ctrl)
		mockCache := cachemock.NewMockCache(ctrl)

		h := handlers.New(mockDB, mockCache, discardCfg, discardLogger)
		assert.NotNil(t, h)

		ctrl.Finish()
	})
}

func TestNewSave(t *testing.T) {
	t.Run("new saver", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockDB := mocks.NewMockDatabase(ctrl)
		mockCache := cachemock.NewMockCache(ctrl)

		h := handlers.New(mockDB, mockCache, discardCfg, discardLogger)

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

		h := handlers.New(mockDB, mockCache, discardCfg, discardLogger)

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

		h := handlers.New(mockDB, mockCache, discardCfg, discardLogger)

		saver := h.NewRedirect()
		assert.NotNil(t, saver)

		ctrl.Finish()
	})
}

func TestSave_DBHitOnly(t *testing.T) {
	testCases := []struct {
		name           string
		dbBehavior     func(m *mocks.MockDatabase, req handlers.Request)
		request        handlers.Request
		expectedStatus int
	}{
		{
			name: "happy path",
			dbBehavior: func(m *mocks.MockDatabase, req handlers.Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, gomock.Any()).Return(nil)
			},
			request:        handlers.Request{URL: "https://www.google.com", Alias: "google"},
			expectedStatus: http.StatusCreated,
		},

		{
			name: "duplicate alias",
			dbBehavior: func(m *mocks.MockDatabase, req handlers.Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, gomock.Any()).Return(database.ErrURLExist)
			},
			request:        handlers.Request{URL: "https://exmaple.com", Alias: "already_exist"},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:           "empty url",
			request:        handlers.Request{Alias: "empty URL"},
			dbBehavior:     func(m *mocks.MockDatabase, req handlers.Request) {},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:    "empty alias",
			request: handlers.Request{URL: "http://url.witout.alias"},
			dbBehavior: func(m *mocks.MockDatabase, req handlers.Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},

		{
			name:    "internal database error",
			request: handlers.Request{URL: "http://valid/url", Alias: "real valid"},
			dbBehavior: func(m *mocks.MockDatabase, req handlers.Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, gomock.Any()).Return(ErrInternalDatabase)
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

			r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
			w := httptest.NewRecorder()

			h := handlers.New(dbMock, mockCache, discardCfg, discardLogger)

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
		expectedStatus int
	}{
		{
			name:  "happy path",
			alias: "alias",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(hasDel, nil)
			},
			expectedStatus: http.StatusOK,
		},

		{
			name:  "unknown alias",
			alias: "unknown",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(noDel, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name:  "internal database error",
			alias: "valid",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(noDel, ErrInternalDatabase)
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
				tt.dbBehavior(dbMock, tt.alias)
			}

			q := url.Values{}
			q.Add("alias", tt.alias)

			r := httptest.NewRequest(http.MethodDelete, path+"?"+q.Encode(), nil)
			w := httptest.NewRecorder()

			h := handlers.New(dbMock, mockCache, discardCfg, discardLogger)

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
		cacheBehavior func(c *cachemock.MockCache, alias string)
		wantStatus    int
		wantRedirect  bool
	}{
		{
			name:  "happy path - cache miss, db hit",
			alias: "valid",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Return("http://something.com", nil)
			},
			cacheBehavior: func(c *cachemock.MockCache, alias string) {
				c.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Set(gomock.Any(), alias, "http://something.com").Return(nil)
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
			cacheBehavior: func(c *cachemock.MockCache, alias string) {
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
			cacheBehavior: func(c *cachemock.MockCache, alias string) {
				c.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(0)
				c.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
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
			cacheBehavior: func(m *cachemock.MockCache, alias string) {
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
			cacheBehavior: func(m *cachemock.MockCache, alias string) {
				m.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				m.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantStatus:   http.StatusInternalServerError,
			wantRedirect: false,
		},
		{
			name:  "cache set error",
			alias: "valid",
			dbBehavior: func(m *mocks.MockDatabase, alias string) {
				m.EXPECT().GetURl(gomock.Any(), alias).Return("http://something.com", nil)
			},
			cacheBehavior: func(m *cachemock.MockCache, alias string) {
				m.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				m.EXPECT().Set(gomock.Any(), alias, "http://something.com").Return(errors.New("cache set error"))
			},
			wantStatus:   http.StatusFound,
			wantRedirect: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dbMock := mocks.NewMockDatabase(ctrl)
			cacheMock := cachemock.NewMockCache(ctrl)

			if tt.dbBehavior != nil {
				tt.dbBehavior(dbMock, tt.alias)
			}

			if tt.cacheBehavior != nil {
				tt.cacheBehavior(cacheMock, tt.alias)
			}

			q := url.Values{}
			if tt.alias != "" {
				q.Add("alias", tt.alias)
			}

			h := handlers.New(dbMock, cacheMock, discardCfg, discardLogger)

			r := httptest.NewRequest(http.MethodGet, path+"?"+q.Encode(), nil)
			w := httptest.NewRecorder()

			redirector := h.NewRedirect()
			redirector(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantRedirect, wasRedirect(w))
		})
	}
}

func wasRedirect(w *httptest.ResponseRecorder) bool {
	return w.Code == http.StatusFound
}

func TestHandlers_SaveRedirectDelete_HappyPath(t *testing.T) {
	t.Run("Save_Redirect_Delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		dbMock := mocks.NewMockDatabase(ctrl)
		cacheMock := cachemock.NewMockCache(ctrl)

		request := handlers.Request{URL: "http://www.google.com", Alias: "google"}
		h := handlers.New(dbMock, cacheMock, discardCfg, discardLogger)

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

			deleter(w, r)

			assert.Equal(t, w.Code, http.StatusOK)
		}

		ctrl.Finish()
	})
}

func TestHandlers_SaveRedirectDelete_BadRequests(t *testing.T) {
	testCases := []struct {
		name                   string
		request                handlers.Request // For save
		alias                  string           // For redirect/delete
		saveBehavior           func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request)
		redirectBehavior       func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string)
		deleteBehavior         func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string)
		expectedSaveStatus     int
		expectedRedirectStatus int
		expectedDeleteStatus   int
	}{
		{
			name:    "empty URL",
			request: handlers.Request{Alias: "empty_url"},
			alias:   "empty_url",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
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
			request: handlers.Request{URL: "invalid-url", Alias: "bad_format"},
			alias:   "bad_format",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
				// No expectations - should fail validation
			},
			expectedSaveStatus: http.StatusBadRequest,
		},
		{
			name:    "empty alias for redirect",
			request: handlers.Request{URL: "http://valid.com", Alias: "valid"},
			alias:   "",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
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
			request: handlers.Request{URL: "http://valid.com", Alias: "will_not_exist"},
			alias:   "will_not_exist",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
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
			request: handlers.Request{URL: "http://valid.com", Alias: "valid"},
			alias:   "",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
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
			request: handlers.Request{URL: "http://valid.com", Alias: "db_error"},
			alias:   "db_error",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(ErrInternalDatabase)
			},
			expectedSaveStatus: http.StatusInternalServerError,
		},
		{
			name:    "database error during redirect",
			request: handlers.Request{URL: "http://valid.com", Alias: "db_error_redirect"},
			alias:   "db_error_redirect",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(nil)
			},
			redirectBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				c.EXPECT().Get(gomock.Any(), alias).Return("", cache.ErrKeyNotExist)
				c.EXPECT().Expire(gomock.Any(), gomock.Any()).Times(0)
				m.EXPECT().GetURl(gomock.Any(), alias).Return("", ErrInternalDatabase)
			},
			expectedSaveStatus:     http.StatusCreated,
			expectedRedirectStatus: http.StatusInternalServerError,
		},
		{
			name:    "database error during delete",
			request: handlers.Request{URL: "http://valid.com", Alias: "db_error_delete"},
			alias:   "db_error_delete",
			saveBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, req handlers.Request) {
				m.EXPECT().SaveURL(gomock.Any(), req.URL, req.Alias).Return(nil)
			},
			deleteBehavior: func(m *mocks.MockDatabase, c *cachemock.MockCache, alias string) {
				m.EXPECT().DeleteURL(gomock.Any(), alias).Return(noDel, ErrInternalDatabase)
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

			h := handlers.New(dbMock, cacheMock, discardCfg, discardLogger)

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
