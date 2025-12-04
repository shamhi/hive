.PHONY: proto

proto:
	rm -rf gen/*
	mkdir -p gen/telemetry gen/tracking gen/order gen/dispatch gen/store

	protoc -I proto proto/telemetry.proto --go_out=gen/telemetry --go_opt=paths=source_relative --go-grpc_out=gen/telemetry --go-grpc_opt=paths=source_relative
	protoc -I proto proto/tracking.proto --go_out=gen/tracking --go_opt=paths=source_relative --go-grpc_out=gen/tracking --go-grpc_opt=paths=source_relative
	protoc -I proto proto/order.proto --go_out=gen/order --go_opt=paths=source_relative --go-grpc_out=gen/order --go-grpc_opt=paths=source_relative
	protoc -I proto proto/dispatch.proto --go_out=gen/dispatch --go_opt=paths=source_relative --go-grpc_out=gen/dispatch --go-grpc_opt=paths=source_relative
	protoc -I proto proto/store.proto --go_out=gen/store --go_opt=paths=source_relative --go-grpc_out=gen/store --go-grpc_opt=paths=source_relative

lint:
	golangci-lint run -v
