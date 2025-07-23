package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pshimaf-Git/url-shortener/api/internal/cache/cachemock"
	"github.com/Pshimaf-Git/url-shortener/api/internal/database/mocks"
	. "github.com/Pshimaf-Git/url-shortener/api/internal/http-server/handlers"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/middleware/ratelimiter"
	"github.com/Pshimaf-Git/url-shortener/api/internal/lib/api/resp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLimit(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := New(
			mocks.NewMockDatabase(ctrl), cachemock.NewMockCache(ctrl),
			discardCfg, discardLogger,
		)

		limiter := h.NewLimit()

		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()

		limiter(w, r)

		expectedBody := resp.Error(ErrTooManyRequests)
		var body resp.Response
		err := json.Unmarshal(w.Body.Bytes(), &body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Equal(t, expectedBody, body)
	})

	t.Run("as_middleware", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		h := New(
			mocks.NewMockDatabase(ctrl), cachemock.NewMockCache(ctrl),
			discardCfg, discardLogger,
		)

		limiter := h.NewLimit()

		mw := ratelimiter.New(0, 1, limiter)

		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()

		mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})).ServeHTTP(w, r)

		expectedBody := resp.Error(ErrTooManyRequests)
		var body resp.Response
		err := json.Unmarshal(w.Body.Bytes(), &body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Equal(t, expectedBody, body)
	})
}
