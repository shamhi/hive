.PHONY: proto

proto:
	rm -rf gen/*
	mkdir -p gen/dispatch gen/order gen/telemetry gen/tracking

	protoc -I proto proto/telemetry.proto --go_out=gen --go_opt=paths=source_relative --go-grpc_out=gen --go-grpc_opt=paths=source_relative
	protoc -I proto proto/tracking.proto --go_out=gen --go_opt=paths=source_relative --go-grpc_out=gen --go-grpc_opt=paths=source_relative
	protoc -I proto proto/order.proto --go_out=gen --go_opt=paths=source_relative --go-grpc_out=gen --go-grpc_opt=paths=source_relative
	protoc -I proto proto/dispatch.proto --go_out=gen --go_opt=paths=source_relative --go-grpc_out=gen --go-grpc_opt=paths=source_relative

	cd gen && go mod init hive/gen || true
	cd gen && go mod tidy
