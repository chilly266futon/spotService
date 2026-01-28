.PHONY: all build gen test help

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

all: deps test build

build:
	go build -o bin/server cmd/main.go

run: build
	./bin/server -config configs/config.yaml

test:
	go test ./... -v

deps:
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

