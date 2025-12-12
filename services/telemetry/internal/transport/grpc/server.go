package grpc

import (
	"context"
	"errors"
	"fmt"
	pb "hive/gen/telemetry"
	"hive/services/telemetry/internal/domain/drone"
	"hive/services/telemetry/internal/domain/mapping"
	"hive/services/telemetry/internal/service"
	"io"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedTelemetryServiceServer
	svc *service.TelemetryService
}

func NewServer(svc *service.TelemetryService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Link(stream pb.TelemetryService_LinkServer) error {
	firstReq, err := stream.Recv()
	if err != nil {
		return err
	}
	if firstReq.GetDroneId() == "" {
		return status.Error(codes.InvalidArgument, "drone id is required")
	}

	droneID := firstReq.GetDroneId()
	conn := s.svc.RegisterConnection(droneID)
	defer s.svc.UnregisterConnection(droneID)

	if err := s.handleRecv(stream.Context(), firstReq); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	errCh := make(chan error, 1)

	go func() {
		for cmd := range conn.Commands {
			resp := &pb.ServerCommand{
				CommandId: cmd.CommandID,
				Action:    mapping.DroneActionToProto(cmd.Action),
				Target:    mapping.LocationToProto(cmd.Target),
				Type:      mapping.DroneTargetTypeToProto(cmd.Type),
			}
			if err := stream.Send(resp); err != nil {
				errCh <- err
				return
			}
		}
	}()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := s.handleRecv(stream.Context(), req); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
}

func (s *Server) SendCommand(ctx context.Context, req *pb.SendCommandRequest) (*pb.SendCommandResponse, error) {
	if req.GetDroneId() == "" {
		return nil, status.Error(codes.InvalidArgument, "drone id is required")
	}

	action, ok := mapping.DroneActionFromProto(req.GetAction())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid drone action")
	}
	ttype, ok := mapping.DroneTargetTypeFromProto(req.GetType())
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid target type")
	}

	cmd := drone.ServerCommand{
		CommandID: uuid.NewString(),
		DroneID:   req.GetDroneId(),
		Action:    action,
		Target:    mapping.LocationFromProto(req.GetTarget()),
		Type:      ttype,
	}
	if err := s.svc.EnqueueCommand(ctx, &cmd); err != nil {
		if errors.Is(err, service.ErrDroneNotConnected) {
			return &pb.SendCommandResponse{Success: false}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendCommandResponse{Success: true}, nil
}

func (s *Server) handleRecv(ctx context.Context, req *pb.DroneTelemetry) error {
	st, ok := mapping.DroneStatusFromProto(req.GetStatus())
	if !ok {
		return fmt.Errorf("invalid drone status: %v", req.GetStatus())
	}
	ev, ok := mapping.DroneEventFromProto(req.GetEvent())
	if !ok {
		return fmt.Errorf("invalid drone event: %v", req.GetEvent())
	}
	loc := mapping.LocationFromProto(req.GetDroneLocation())
	if loc == nil {
		return fmt.Errorf("invalid drone location")
	}

	tm := drone.Telemetry{
		DroneID:             req.GetDroneId(),
		DroneLocation:       *loc,
		Battery:             req.GetBattery(),
		SpeedMps:            req.GetSpeedMps(),
		ConsumptionPerMeter: req.GetConsumptionPerMeter(),
		Status:              st,
		Timestamp:           req.GetTimestamp(),
		Event:               ev,
	}
	return s.svc.HandleTelemetry(ctx, tm)
}
