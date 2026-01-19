.PHONY: gen

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