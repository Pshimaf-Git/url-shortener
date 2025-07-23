package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

func WithAddr(addr string) ServerOption {
	return func(s *Server) {
		s.serv.Addr = addr
	}
}

func WithHostPort(host string, port string) ServerOption {
	return WithAddr(BuildAddr(host, port))
}

func WithTLSConfig(tlsConfig *tls.Config) ServerOption {
	return func(s *Server) {
		s.serv.TLSConfig = tlsConfig
	}
}

func WithProtocols(protocols *http.Protocols) ServerOption {
	return func(s *Server) {
		s.serv.Protocols = protocols
	}
}

func WithDisableGeneralOptionsHandler(b bool) ServerOption {
	return func(s *Server) {
		s.serv.DisableGeneralOptionsHandler = b
	}
}

func WithErrLogger(logger *log.Logger) ServerOption {
	return func(s *Server) {
		s.serv.ErrorLog = logger
	}
}

func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.serv.ReadTimeout = timeout
	}
}

func WithReadHeaderTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.serv.ReadHeaderTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.serv.WriteTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.serv.IdleTimeout = timeout
	}
}

func WithMaxHeaderBytes(max int) ServerOption {
	return func(s *Server) {
		s.serv.MaxHeaderBytes = max
	}
}

func WithHandler(handler http.Handler) ServerOption {
	return func(s *Server) {
		s.serv.Handler = handler
	}
}
