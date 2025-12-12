FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY gen ./gen
COPY services/telemetry ./services/telemetry

WORKDIR /app/services/telemetry

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/bin/server \
    ./cmd/server/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/bin/emulator \
    ./cmd/emulator/main.go

FROM alpine:latest AS runtime

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/emulator /app/emulator
COPY --from=builder /app/services/telemetry/config /app/config

RUN chown -R appuser:appgroup /app
USER appuser

FROM runtime AS server

ENTRYPOINT ["/app/server"]

FROM runtime AS emulator

ENTRYPOINT ["/app/emulator"]