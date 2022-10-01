package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nikolai5slo/ttlock2mqtt/server/handlers"
)

type Server struct {
	address  string
	engine   *gin.Engine
	handlers *handlers.Handlers
}

type Conf func(*Server) error

func New(cfg ...Conf) (*Server, error) {
	srv := &Server{
		address: "0.0.0.0:8080", // Default address
	}

	for _, c := range cfg {
		if err := c(srv); err != nil {
			return srv, fmt.Errorf("failed to configure server: %w", err)
		}
	}

	return srv, nil
}

func WithHandlers(h *handlers.Handlers) Conf {
	return func(s *Server) error {
		s.handlers = h
		return nil
	}
}

func WithAddress(address string) Conf {
	return func(s *Server) error {
		s.address = address
		return nil
	}
}

func (s *Server) Run() error {
	r := gin.Default()

	s.engine = r

	// Load tempaltes
	r.LoadHTMLGlob("templates/*")

	s.handlers.Register(s.engine)

	err := r.Run(s.address)

	if err != nil {
		return err
	}

	return nil
}
