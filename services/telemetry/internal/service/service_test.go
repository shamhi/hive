package service

import (
	"context"
	"errors"
	"hive/pkg/logger"
	"testing"

	"hive/services/telemetry/internal/domain/drone"
)

var (
	lg, _ = logger.NewLogger("dev")
)

func TestTelemetryService_RegisterConnection_ReplacesOld(t *testing.T) {
	svc := NewTelemetryService(nil, nil, "", "", lg)

	c1 := svc.RegisterConnection("d1")
	c2 := svc.RegisterConnection("d1")

	if c1 == c2 {
		t.Fatalf("expected different connections")
	}

	select {
	case _, ok := <-c1.Commands:
		if ok {
			t.Fatalf("expected closed old channel")
		}
	default:
		t.Fatalf("expected closed old channel")
	}
}

func TestTelemetryService_UnregisterConnection_Closes(t *testing.T) {
	svc := NewTelemetryService(nil, nil, "", "", lg)

	c := svc.RegisterConnection("d1")
	svc.UnregisterConnection("d1")

	select {
	case _, ok := <-c.Commands:
		if ok {
			t.Fatalf("expected closed channel")
		}
	default:
		t.Fatalf("expected closed channel")
	}
}

func TestTelemetryService_EnqueueCommand_NotConnected(t *testing.T) {
	svc := NewTelemetryService(nil, nil, "", "", lg)

	err := svc.EnqueueCommand(context.Background(), &drone.ServerCommand{DroneID: "d1"})
	if errors.Is(err, ErrDroneNotConnected) {
		t.Fatalf("expected ErrDroneNotConnected")
	}
}

func TestTelemetryService_EnqueueCommand_Success(t *testing.T) {
	svc := NewTelemetryService(nil, nil, "", "", lg)

	conn := svc.RegisterConnection("d1")
	cmd := &drone.ServerCommand{DroneID: "d1", CommandID: "c1"}

	err := svc.EnqueueCommand(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := <-conn.Commands
	if got.CommandID != "c1" {
		t.Fatalf("unexpected command")
	}
}

func TestTelemetryService_EnqueueCommand_ContextCanceledWhenFull(t *testing.T) {
	svc := NewTelemetryService(nil, nil, "", "", lg)

	conn := svc.RegisterConnection("d1")
	for i := 0; i < cap(conn.Commands); i++ {
		conn.Commands <- &drone.ServerCommand{DroneID: "d1"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.EnqueueCommand(ctx, &drone.ServerCommand{DroneID: "d1"})
	if err == nil {
		t.Fatalf("expected error")
	}
}
