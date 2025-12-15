FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY gen ./gen
COPY services/store ./services/store

WORKDIR /app/services/store

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/bin/server \
    ./cmd/server/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/bin/migrate \
    ./cmd/migrate/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /app/bin/generate \
    ./cmd/generate/main.go

FROM alpine:latest AS runtime

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/migrate /app/migrate
COPY --from=builder /app/bin/generate /app/generate
COPY --from=builder /app/services/store/config /app/config
COPY --from=builder /app/services/store/migrations /app/migrations

RUN chown -R appuser:appgroup /app
USER appuser

FROM runtime AS server

ENTRYPOINT ["/app/server"]

FROM runtime AS migrate

ENTRYPOINT ["/app/migrate"]

FROM runtime AS generate

ENTRYPOINT ["/app/generate"]