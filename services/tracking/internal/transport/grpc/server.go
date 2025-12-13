package grpc

import (
	"fmt"
	"hive/pkg/logger"
	"hive/services/tracking/internal/config"
	"hive/services/tracking/internal/repository"
	"hive/services/tracking/internal/service"
	"net"
	"strconv"

	trackingGen "hive/gen/tracking"

	"google.golang.org/grpc"
)

type Server struct {
	srv *grpc.Server
	lis net.Listener
}

func New(cfg *config.GRPCConfig, lg *logger.Logger) (*Server, error) {
	addr := cfg.Host + ":" + strconv.Itoa(cfg.Port)
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

func (s *Server) RegisterService(repo repository.TrackingRepository) {
	trackingService := service.New(repo)
	trackingGen.RegisterTrackingServiceServer(s.srv, trackingService)
}

func (s *Server) Start() error {
	return s.srv.Serve(s.lis)
}

func (s *Server) Stop() {
	s.srv.GracefulStop()
}
