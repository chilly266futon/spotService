.PHONY: build run gen test help

gen:
	@mkdir -p gen/pb
	@find proto -name "*.proto" -exec \
		protoc \
		  -I proto \
		  --go_out=gen/pb \
		  --go-grpc_out=gen/pb \
		  --go_opt=paths=source_relative \
		  --go-grpc_opt=paths=source_relative \
		  {} \;

build:
	go build -o bin/spot_service cmd/main.go

run: build
	./bin/spot_service -config configs/config.yaml

test:
	go test ./... -v

deps:
	go mod download
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

