package server

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Server struct {
	serv *http.Server
}

type ServerOptions func(*Server)

func BuildAddr(host, port string) string { return net.JoinHostPort(host, port) }

func DefaultServer() (serv *Server) {
	return &Server{
		serv: &http.Server{
			Handler:           nil,
			MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			ReadTimeout:       time.Minute,
			ReadHeaderTimeout: time.Minute,
			WriteTimeout:      time.Minute,
			IdleTimeout:       time.Minute,
		},
	}
}

func New(opts ...ServerOptions) *Server {
	server := DefaultServer()

	for _, fn := range opts {
		fn(server)
	}

	return server
}

func (s *Server) Run() error {
	return s.serv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.serv.Shutdown(ctx)
}
