package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

func WithAddr(addr string) ServerOptions {
	return func(s *Server) {
		s.serv.Addr = addr
	}
}

func WithTLSConfig(tlsConfig *tls.Config) ServerOptions {
	return func(s *Server) {
		s.serv.TLSConfig = tlsConfig
	}
}

func WithProtocol(protocols *http.Protocols) ServerOptions {
	return func(s *Server) {
		s.serv.Protocols = protocols
	}
}

func WithDisableGeneralOptionsHandler(b bool) ServerOptions {
	return func(s *Server) {
		s.serv.DisableGeneralOptionsHandler = b
	}
}

func WithErrLogger(logger *log.Logger) ServerOptions {
	return func(s *Server) {
		s.serv.ErrorLog = logger
	}
}

func WithReadTimeout(timeout time.Duration) ServerOptions {
	return func(s *Server) {
		s.serv.ReadTimeout = timeout
	}
}

func WithReadHeaderTimeout(timeout time.Duration) ServerOptions {
	return func(s *Server) {
		s.serv.ReadHeaderTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) ServerOptions {
	return func(s *Server) {
		s.serv.WriteTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) ServerOptions {
	return func(s *Server) {
		s.serv.IdleTimeout = timeout
	}
}

func WithMaxHeaderBytes(max int) ServerOptions {
	return func(s *Server) {
		s.serv.MaxHeaderBytes = max
	}
}

func WithHandler(handler http.Handler) ServerOptions {
	return func(s *Server) {
		s.serv.Handler = handler
	}
}
