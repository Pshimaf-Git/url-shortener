package server

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerOptions(t *testing.T) {
	t.Run("WithAddr", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		WithAddr(":8080")(s)
		assert.Equal(t, ":8080", s.serv.Addr)
	})

	t.Run("WithTLSConfig", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		cfg := &tls.Config{}
		WithTLSConfig(cfg)(s)
		assert.Equal(t, cfg, s.serv.TLSConfig)
	})

	t.Run("WithProtocol", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		protocols := &http.Protocols{}
		WithProtocol(protocols)(s)
		assert.Equal(t, protocols, s.serv.Protocols)
	})

	t.Run("WithDisableGeneralOptionsHandler", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		WithDisableGeneralOptionsHandler(true)(s)
		assert.True(t, s.serv.DisableGeneralOptionsHandler)
	})

	t.Run("WithErrLogger", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		logger := log.Default()
		WithErrLogger(logger)(s)
		assert.Equal(t, logger, s.serv.ErrorLog)
	})

	t.Run("WithReadTimeout", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		timeout := 10 * time.Second
		WithReadTimeout(timeout)(s)
		assert.Equal(t, timeout, s.serv.ReadTimeout)
	})

	t.Run("WithReadHeaderTimeout", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		timeout := 5 * time.Second
		WithReadHeaderTimeout(timeout)(s)
		assert.Equal(t, timeout, s.serv.ReadHeaderTimeout)
	})

	t.Run("WithWriteTimeout", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		timeout := 15 * time.Second
		WithWriteTimeout(timeout)(s)
		assert.Equal(t, timeout, s.serv.WriteTimeout)
	})

	t.Run("WithIdleTimeout", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		timeout := 30 * time.Second
		WithIdleTimeout(timeout)(s)
		assert.Equal(t, timeout, s.serv.IdleTimeout)
	})

	t.Run("WithMaxHeaderBytes", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		max := 4096
		WithMaxHeaderBytes(max)(s)
		assert.Equal(t, max, s.serv.MaxHeaderBytes)
	})

	t.Run("WithHandler", func(t *testing.T) {
		s := &Server{serv: &http.Server{}}
		handler := http.NewServeMux()
		WithHandler(handler)(s)
		assert.Equal(t, handler, s.serv.Handler)
	})
}

func TestBuildAddr(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		expected string
	}{
		{
			name:     "IPv4 address",
			host:     "127.0.0.1",
			port:     "8080",
			expected: "127.0.0.1:8080",
		},
		{
			name:     "IPv6 address",
			host:     "::1",
			port:     "8080",
			expected: "[::1]:8080",
		},
		{
			name:     "Hostname",
			host:     "localhost",
			port:     "8080",
			expected: "localhost:8080",
		},
		{
			name:     "Empty host",
			host:     "",
			port:     "8080",
			expected: ":8080",
		},
		{
			name:     "Empty port",
			host:     "localhost",
			port:     "",
			expected: "localhost:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildAddr(tt.host, tt.port)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultServer(t *testing.T) {
	server := DefaultServer()

	require.NotNil(t, server)
	require.NotNil(t, server.serv)

	assert.Nil(t, server.serv.Handler)
	assert.Equal(t, http.DefaultMaxHeaderBytes, server.serv.MaxHeaderBytes)
	assert.Equal(t, time.Minute, server.serv.ReadTimeout)
	assert.Equal(t, time.Minute, server.serv.ReadHeaderTimeout)
	assert.Equal(t, time.Minute, server.serv.WriteTimeout)
	assert.Equal(t, time.Minute, server.serv.IdleTimeout)
}

func TestNewWithOptions(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name     string
		options  []ServerOptions
		validate func(*testing.T, *Server)
	}{
		{
			name: "WithHandler",
			options: []ServerOptions{
				WithHandler(handler),
			},
			validate: func(t *testing.T, s *Server) {
				assert.NotNil(t, s.serv.Handler)
			},
		},
		{
			name: "WithAddress",
			options: []ServerOptions{
				WithAddr(":8080"),
			},
			validate: func(t *testing.T, s *Server) {
				assert.Equal(t, ":8080", s.serv.Addr)
			},
		},
		{
			name: "MultipleOptions",
			options: []ServerOptions{
				WithAddr(":9090"),
				WithHandler(handler),
			},
			validate: func(t *testing.T, s *Server) {
				assert.Equal(t, ":9090", s.serv.Addr)
				assert.NotNil(t, s.serv.Handler)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New(tt.options...)
			require.NotNil(t, server)
			tt.validate(t, server)
		})
	}
}

func TestServer_Run(t *testing.T) {
	t.Run("SuccessfulStart", func(t *testing.T) {
		server := New(
			WithAddr(":0"),
			WithHandler(http.NewServeMux()),
		)

		errChan := make(chan error, 1)
		go func() {
			errChan <- server.Run()
		}()

		time.Sleep(100 * time.Millisecond)
		select {
		case err := <-errChan:
			require.NoError(t, err)
		default:
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, server.Shutdown(ctx))
	})

	t.Run("InvalidAddress", func(t *testing.T) {
		server := New(
			WithAddr("invalid-address-format"),
			WithHandler(http.NewServeMux()),
		)

		err := server.Run()
		require.Error(t, err)
	})
}

func TestServer_Shutdown(t *testing.T) {
	t.Run("SuccessfulShutdown", func(t *testing.T) {
		server := New(
			WithAddr(":0"),
			WithHandler(http.NewServeMux()),
		)

		go func() {
			_ = server.Run()
		}()
		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, server.Shutdown(ctx))
	})

	t.Run("ShutdownNotRunning", func(t *testing.T) {
		server := New()
		err := server.Shutdown(context.Background())
		require.NoError(t, err)
	})

	t.Run("ShutdownContextCancellation", func(t *testing.T) {
		server := New(
			WithAddr(":0"),
			WithHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
			})),
		)

		go func() {
			_ = server.Run()
		}()
		time.Sleep(100 * time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := server.Shutdown(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < 100*time.Millisecond, "Shutdown должен завершиться быстро")
		require.NoError(t, err)
	})
}
