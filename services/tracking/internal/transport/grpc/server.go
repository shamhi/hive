package grpc

import (
	"fmt"
	"hive/pkg/logger"
	"hive/services/tracking/internal/config"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	srv *grpc.Server
	lis net.Listener
}

func New(cfg *config.GRPCServerConfig, lg *logger.Logger) (*Server, error) {
	addr := fmt.Sprintf(":%d", cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s := grpc.NewServer()

	return &Server{
		srv: s,
		lis: lis,
	}, nil
}

func (s *Server) Start() error {
	return s.srv.Serve(s.lis)
}

func (s *Server) Stop() {
	s.srv.GracefulStop()
}
