package grpc

import (
	"context"
	"hive/pkg/logger"
	"testing"

	pbCommon "hive/gen/common"
	pb "hive/gen/telemetry"
	"hive/services/telemetry/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	lg, _ = logger.NewLogger("dev")
)

func TestServer_SendCommand_InvalidArgument(t *testing.T) {
	svc := service.NewTelemetryService(nil, nil, "", "", lg)
	srv := NewServer(svc)

	_, err := srv.SendCommand(context.Background(), &pb.SendCommandRequest{
		DroneId: "",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestServer_SendCommand_NotConnected_FallbackSuccessFalse(t *testing.T) {
	svc := service.NewTelemetryService(nil, nil, "", "", lg)
	srv := NewServer(svc)

	resp, err := srv.SendCommand(context.Background(), &pb.SendCommandRequest{
		DroneId: "d1",
		Action:  pb.DroneAction_ACTION_FLY_TO,
		Target:  &pbCommon.Location{Lat: 55, Lon: 37},
		Type:    pb.TargetType_TARGET_POINT,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetSuccess() {
		t.Fatalf("expected success=false")
	}
}

func TestServer_SendCommand_Connected_SuccessTrue(t *testing.T) {
	svc := service.NewTelemetryService(nil, nil, "", "", lg)
	svc.RegisterConnection("d1")
	srv := NewServer(svc)

	resp, err := srv.SendCommand(context.Background(), &pb.SendCommandRequest{
		DroneId: "d1",
		Action:  pb.DroneAction_ACTION_FLY_TO,
		Target:  &pbCommon.Location{Lat: 55, Lon: 37},
		Type:    pb.TargetType_TARGET_POINT,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Fatalf("expected success=true")
	}
}
