.PHONY: generate-go setup benchmark-http benchmark-grpc benchmark-go-http benchmark-go-grpc clean

# Generate Go code from protobuf definitions
generate-go:
	protoc --go_out=. --go-grpc_out=. --proto_path=./protos ./protos/gateway.proto

# Install Python dependencies in virtual environment
setup:
	python3 -m venv .venv
	. .venv/bin/activate && pip install -r requirements.txt

# Run Python HTTP benchmark
benchmark-http:
	. .venv/bin/activate && python benchmark.py --config config.json

# Run Python gRPC benchmark
benchmark-grpc:
	. .venv/bin/activate && python benchmark.py --config config.json --grpc

# Run Go HTTP benchmark
benchmark-go-http:
	go run main.go -config config.json

# Run Go gRPC benchmark
benchmark-go-grpc:
	go run main.go -config config.json --grpc

# Clean generated files
clean:
	rm -rf pb/*.go
	rm -rf .venv

