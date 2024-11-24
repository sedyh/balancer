package web

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	server *http.Server
	group  *errgroup.Group
}

func NewServer(h http.Handler, addr string, limit int, timeout time.Duration) (*Server, error) {
	s := &Server{
		server: &http.Server{
			Addr:         addr,
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
			Handler:      NewLimiter(h, limit),
		},
		group: &errgroup.Group{},
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s.group.Go(func() error {
		if err := s.server.Serve(l); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	return s, nil
}

func (s *Server) Close(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return s.group.Wait()
}
