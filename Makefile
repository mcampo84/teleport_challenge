protos:
	find . -name "*.proto" -exec protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. {} \;
	go mod tidy
