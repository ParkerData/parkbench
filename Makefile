generate-go:
	protoc --go_out=. --go-grpc_out=.  --proto_path=./protos ./protos/parker.proto ./protos/common.proto
