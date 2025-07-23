package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/Pshimaf-Git/url-shortener/api/internal/http-server/middleware/ratelimiter"
	"github.com/go-chi/httprate"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	limitHandlerCalled := false
	limitHandler := func(w http.ResponseWriter, r *http.Request) {
		limitHandlerCalled = true
		w.WriteHeader(http.StatusTooManyRequests)
	}

	limiter := ratelimiter.New(1, time.Second, limitHandler)
	assert.NotNil(t, limiter)

	// Should allow first request, block second
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, limitHandlerCalled)

	// Second request should trigger rate limit
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.True(t, limitHandlerCalled)
}

func TestNewWithConfig(t *testing.T) {
	cfg := &config.ServerConfig{
		RequesLimit:  1,
		WindowLength: time.Second,
	}
	limitHandlerCalled := false
	limitHandler := func(w http.ResponseWriter, r *http.Request) {
		limitHandlerCalled = true
		w.WriteHeader(http.StatusTooManyRequests)
	}

	limiter := ratelimiter.NewWithConfig(cfg, limitHandler)
	assert.NotNil(t, limiter)

	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, limitHandlerCalled)

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.True(t, limitHandlerCalled)
}

func TestNew_Options(t *testing.T) {
	limitHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}
	opt := httprate.WithLimitHandler(limitHandler)
	limiter := ratelimiter.New(2, time.Second, limitHandler, opt)
	assert.NotNil(t, limiter)
}
