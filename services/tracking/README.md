# Tracking

## Configuration

example for local development

GRPC Server

```bash
# GRPC Server
export GRPCHOST=127.0.0.1
export GRPCPORT=50051

# Redis
export REDIS_HOST=127.0.0.1
export REDIS_PORT=6379
export PREDIS_PASSOWRD=
export REDIS_DB=0

# Environment
export ENV=dev
```

Kafka Worker
```bash
# Kafka
export KAFKA_BROKERS=localhost:9092
export KAFKA_GROUP_ID=tracking-group
export KAFKA_TOPIC=telemetry.raw

# Redis
export REDIS_HOST=127.0.0.1
export REDIS_PORT=6379
export PREDIS_PASSOWRD=
export REDIS_DB=0

# Environment
export ENV=dev
```

## Start

GRPC Server

```bash
go run ./services/tracking/cmd/grpc-server/main.go
```

Kafka Worker

```bash
go run ./services/tracking/cmd/kafka-worker/main.go
```
