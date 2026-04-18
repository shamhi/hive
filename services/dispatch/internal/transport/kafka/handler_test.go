package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestHandle_InvalidJSON_ReturnsWrappedError(t *testing.T) {
	h := NewHandler(nil)

	err := h.Handle(context.Background(), []byte("{"))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal telemetry") {
		t.Fatalf("expected error to contain %q, got %q", "unmarshal telemetry", err.Error())
	}

	if _, ok := errors.AsType[*json.SyntaxError](err); !ok {
		t.Fatalf("expected wrapped *json.SyntaxError, got %T", err)
	}
}

func TestHandle_EmptyJSON_ReturnsWrappedError(t *testing.T) {
	h := NewHandler(nil)

	err := h.Handle(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal telemetry") {
		t.Fatalf("expected error to contain %q, got %q", "unmarshal telemetry", err.Error())
	}
}
