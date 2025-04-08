.PHONY: generate-go generate-python setup benchmark-http benchmark-grpc benchmark-go-http benchmark-go-grpc clean

# Generate Go code from protobuf definitions
generate-go:
	protoc --go_out=. --go-grpc_out=. --proto_path=./protos ./protos/gateway.proto

# Install Python dependencies in virtual environment
setup:
	python3 -m venv .venv
	. .venv/bin/activate && pip install -r requirements.txt
	. .venv/bin/activate && pip install grpcio-tools
	PYTHONPATH=. . .venv/bin/activate && python3 -m grpc_tools.protoc -I./protos --python_out=pb --grpc_python_out=pb ./protos/gateway.proto

# Generate Python code from protobuf definitions (requires setup to be run first)
generate-python:
	PYTHONPATH=. . .venv/bin/activate && python3 -m grpc_tools.protoc -I./protos --python_out=pb --grpc_python_out=pb ./protos/gateway.proto

# Run Python HTTP benchmark
benchmark-http:
	PYTHONPATH=. . .venv/bin/activate && python benchmark.py --config config.json

# Run Python gRPC benchmark
benchmark-grpc:
	PYTHONPATH=. . .venv/bin/activate && python benchmark.py --config config.json --grpc

# Run Go HTTP benchmark
benchmark-go-http:
	go run main.go -config config.json

# Run Go gRPC benchmark
benchmark-go-grpc:
	go run main.go -config config.json --grpc

# Clean generated files
clean:
	rm -rf pb/gateway_pb2*.py
	rm -rf pb/__pycache__
	rm -rf pb/*.go
	rm -rf .venv

