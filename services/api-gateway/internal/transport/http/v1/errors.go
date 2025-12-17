package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func writeGRPCError(c echo.Context, err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return jsonError(c, http.StatusInternalServerError, fallbackMsg)
	}

	msg := st.Message()
	if msg == "" {
		msg = fallbackMsg
	}

	switch st.Code() {
	case codes.NotFound:
		return jsonError(c, http.StatusNotFound, msg)
	case codes.InvalidArgument:
		return jsonError(c, http.StatusBadRequest, msg)
	case codes.AlreadyExists:
		return jsonError(c, http.StatusConflict, msg)
	case codes.FailedPrecondition:
		return jsonError(c, http.StatusPreconditionFailed, msg)
	case codes.Unauthenticated:
		return jsonError(c, http.StatusUnauthorized, msg)
	case codes.PermissionDenied:
		return jsonError(c, http.StatusForbidden, msg)
	case codes.DeadlineExceeded:
		return jsonError(c, http.StatusGatewayTimeout, "upstream timeout")
	case codes.Unavailable, codes.ResourceExhausted:
		return jsonError(c, http.StatusServiceUnavailable, "upstream unavailable")
	default:
		return jsonError(c, http.StatusInternalServerError, fallbackMsg)
	}
}
