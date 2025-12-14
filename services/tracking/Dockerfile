FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY gen ./gen
COPY services/tracking ./services/tracking

WORKDIR /app/services/tracking

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/grpc-server \
    ./cmd/grpc-server/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/kafka-worker \
    ./cmd/kafka-worker/main.go

FROM alpine:latest as runtime
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -S -D appuser

WORKDIR /app

FROM runtime AS grpc-server
USER appuser
COPY --from=builder /app/bin/grpc-server /app/grpc-server
ENTRYPOINT ["/app/grpc-server"]

FROM runtime AS kafka-worker
USER appuser
COPY --from=builder /app/bin/kafka-worker /app/kafka-worker
ENTRYPOINT ["/app/kafka-worker"]
