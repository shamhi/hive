package grpc

import (
	"context"
	"testing"

	pb "hive/gen/dispatch"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAssignDrone_EmptyOrderID(t *testing.T) {
	srv := NewServer(nil, nil)

	resp, err := srv.AssignDrone(context.Background(), &pb.AssignDroneRequest{})
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected code %v, got %v (err=%v)", codes.InvalidArgument, status.Code(err), err)
	}
}

func TestAssignDrone_NilDeliveryLocation(t *testing.T) {
	srv := NewServer(nil, nil)

	resp, err := srv.AssignDrone(context.Background(), &pb.AssignDroneRequest{
		OrderId: "order-1",
	})
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected code %v, got %v (err=%v)", codes.InvalidArgument, status.Code(err), err)
	}
}

func TestGetAssignment_EmptyDroneID(t *testing.T) {
	srv := NewServer(nil, nil)

	resp, err := srv.GetAssignment(context.Background(), &pb.GetAssignmentRequest{})
	if resp != nil {
		t.Fatalf("expected nil response, got %+v", resp)
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected code %v, got %v (err=%v)", codes.InvalidArgument, status.Code(err), err)
	}
}
