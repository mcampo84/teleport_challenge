.PHONY: certs protos start_server start_client test

certs:
	zsh scripts/generate_certs.sh

protos:
	find . -name "*.proto" -exec protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. {} \;
	go mod tidy

start_server:
	go run cmd/server/main.go

start_client:
	go run cmd/client/main.go

test:
	go test ./...
