package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Pshimaf-Git/url-shortener/api/internal/config"
)

type Server struct {
	serv *http.Server
}

type ServerOption func(*Server)

func BuildAddr(host, port string) string {
	return net.JoinHostPort(host, port)
}

func DefaultServer() (serv *Server) {
	return &Server{
		serv: &http.Server{
			Handler:           http.NewServeMux(),
			MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			ReadTimeout:       time.Minute,
			ReadHeaderTimeout: time.Minute,
			WriteTimeout:      time.Minute,
			IdleTimeout:       time.Minute,
		},
	}
}

func New(options ...ServerOption) *Server {
	server := DefaultServer()

	for _, option := range options {
		if option != nil {
			option(server)
		}
	}

	return server
}

func NewWithConfig(cfg *config.ServerConfig, handler http.Handler, options ...ServerOption) *Server {
	return New(
		WithHostPort(cfg.Host, cfg.Port),
		WithHandler(handler),
		WithReadTimeout(cfg.ReadTimeout),
		WithWriteTimeout(cfg.WriteTimeout),
		WithIdleTimeout(cfg.IdleTimeout),
	)
}

func (s *Server) Run() error {
	return s.serv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.serv.Shutdown(ctx)
}
