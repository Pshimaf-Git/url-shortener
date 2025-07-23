package server

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("without_options", func(t *testing.T) {
		server := New()
		require.NotNil(t, server)
		assert.Equal(t, DefaultServer(), server)
	})
}

func TestNewWithConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config *config.ServerConfig
	}{
		{
			name: "default_config",
			config: &config.ServerConfig{
				Host:         "localhost",
				Port:         "8080",
				ReadTimeout:  30 * time.Second,
				IdleTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
			},
		},

		{
			name:   "empty_config",
			config: &config.ServerConfig{},
		},

		{
			name: "custom_config",
			config: &config.ServerConfig{
				Host:         "",
				Port:         "9090",
				ReadTimeout:  10 * time.Second,
				IdleTimeout:  20 * time.Second,
				WriteTimeout: 15 * time.Second,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			server := NewWithConfig(tt.config, http.NewServeMux())
			require.NotNil(t, server)
			require.NotNil(t, server.serv)

			host, port, err := net.SplitHostPort(server.serv.Addr)
			require.NoError(t, err)

			assert.Equal(t, tt.config.Host, host)
			assert.Equal(t, tt.config.Port, port)
			assert.Equal(t, tt.config.ReadTimeout, server.serv.ReadTimeout)
			assert.Equal(t, tt.config.IdleTimeout, server.serv.IdleTimeout)
			assert.Equal(t, tt.config.WriteTimeout, server.serv.WriteTimeout)
			assert.Equal(t, tt.config.ReadTimeout, server.serv.ReadTimeout)
		})
	}
}

func TestServerOptions(t *testing.T) {
	t.Run("WithAddr", func(t *testing.T) {
		addr := ":3000"
		s := New(WithAddr(addr))
		assert.Equal(t, addr, s.serv.Addr)
	})

	t.Run("WithTLSConfig", func(t *testing.T) {
		cfg := &tls.Config{}
		s := New(WithTLSConfig(cfg))
		assert.Equal(t, cfg, s.serv.TLSConfig)
	})

	t.Run("WithProtocols", func(t *testing.T) {
		protocols := &http.Protocols{}
		s := New(WithProtocols(protocols))
		assert.Equal(t, protocols, s.serv.Protocols)
	})

	t.Run("WithDisableGeneralOptionsHandler", func(t *testing.T) {
		s := New(WithDisableGeneralOptionsHandler(true))
		assert.True(t, s.serv.DisableGeneralOptionsHandler)
	})

	t.Run("WithErrLogger", func(t *testing.T) {
		logger := log.Default()
		s := New(WithErrLogger(logger))
		assert.Same(t, logger, s.serv.ErrorLog)
	})

	t.Run("WithReadTimeout", func(t *testing.T) {
		timeout := 10 * time.Second
		s := New(WithReadTimeout(timeout))
		assert.Equal(t, timeout, s.serv.ReadTimeout)
	})

	t.Run("WithReadHeaderTimeout", func(t *testing.T) {
		timeout := 5 * time.Second
		s := New(WithReadHeaderTimeout(timeout))
		assert.Equal(t, timeout, s.serv.ReadHeaderTimeout)
	})

	t.Run("WithWriteTimeout", func(t *testing.T) {
		timeout := 15 * time.Second
		s := New(WithWriteTimeout(timeout))
		assert.Equal(t, timeout, s.serv.WriteTimeout)
	})

	t.Run("WithIdleTimeout", func(t *testing.T) {
		timeout := 30 * time.Second
		s := New(WithIdleTimeout(timeout))
		assert.Equal(t, timeout, s.serv.IdleTimeout)
	})

	t.Run("WithMaxHeaderBytes", func(t *testing.T) {
		s := New(WithMaxHeaderBytes(http.DefaultMaxHeaderBytes))
		assert.Equal(t, http.DefaultMaxHeaderBytes, s.serv.MaxHeaderBytes)
	})

	t.Run("WithHandler", func(t *testing.T) {
		handler := http.NewServeMux()
		s := New(WithHandler(handler))
		assert.Same(t, handler, s.serv.Handler)
	})

	t.Run("WithHostPort", func(t *testing.T) {
		host, port := "127.0.0.1", ":8000"
		s := New(WithHostPort(host, port))
		assert.Equal(t, net.JoinHostPort(host, port), s.serv.Addr)
	})

	t.Run("multiply_opts", func(t *testing.T) {
		var (
			host, port        = "127.0.0.1:", ":80"
			handler           = chi.NewRouter()
			readTimeout       = time.Second
			readHeaderTimeout = time.Duration(100000)
			writeTimeot       = time.Hour
			idleTimeout       = time.Millisecond * 10000
			maxHeaderBytes    = http.DefaultMaxHeaderBytes
			errlogger         = log.New(io.Discard, "pref", log.Lmicroseconds)
			protocols         = &http.Protocols{}
			tlsConfig         = &tls.Config{}
		)

		options := []ServerOption{
			WithHostPort(host, port),
			WithHandler(handler),
			WithMaxHeaderBytes(maxHeaderBytes),
			WithReadTimeout(readTimeout),
			WithReadHeaderTimeout(readHeaderTimeout),
			WithIdleTimeout(idleTimeout),
			WithWriteTimeout(writeTimeot),
			WithErrLogger(errlogger),
			WithDisableGeneralOptionsHandler(true),
			WithProtocols(protocols),
			WithTLSConfig(tlsConfig),
		}

		s := New(options...)

		assert.Equal(t, net.JoinHostPort(host, port), s.serv.Addr)
		assert.Equal(t, readTimeout, s.serv.ReadTimeout)
		assert.Equal(t, writeTimeot, s.serv.WriteTimeout)
		assert.Equal(t, idleTimeout, s.serv.IdleTimeout)
		assert.Equal(t, readHeaderTimeout, s.serv.ReadHeaderTimeout)
		assert.Equal(t, maxHeaderBytes, s.serv.MaxHeaderBytes)

		assert.Same(t, handler, s.serv.Handler)
		assert.Same(t, errlogger, s.serv.ErrorLog)
		assert.Same(t, tlsConfig, s.serv.TLSConfig)
		assert.Same(t, protocols, s.serv.Protocols)
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

	assert.NotNil(t, server.serv.Handler)
	assert.Equal(t, http.DefaultMaxHeaderBytes, server.serv.MaxHeaderBytes)
	assert.Equal(t, time.Minute, server.serv.ReadTimeout)
	assert.Equal(t, time.Minute, server.serv.ReadHeaderTimeout)
	assert.Equal(t, time.Minute, server.serv.WriteTimeout)
	assert.Equal(t, time.Minute, server.serv.IdleTimeout)
}

func TestServer_Run(t *testing.T) {
	t.Run("SuccessfulStart", func(t *testing.T) {
		server := New(
			WithAddr(":0"),
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
		)

		err := server.Run()
		require.Error(t, err)
	})
}

func TestServer_Shutdown(t *testing.T) {
	t.Run("SuccessfulShutdown", func(t *testing.T) {
		server := New(
			WithAddr(":0"),
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
