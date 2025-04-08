# Parker Benchmark Tool

A benchmarking tool for testing the performance of Parker services, available in both Python and Go implementations. The tool supports both HTTP and gRPC protocols and provides detailed performance metrics.

## Prerequisites

- Go 1.21 or later (for Go benchmarks)
- Python 3.8+ (for Python benchmarks)
- Make

## Installation

1. Clone the repository
2. Install dependencies:
```bash
# Install Python dependencies
make setup

# Generate Go protobuf files
make generate-go
```

## Configuration

Create a `config.json` file with the following structure:

```json
{
    "csv": "test_data.csv",
    "account": "your_account",
    "table": "your_table",
    "httpAddress": "http://localhost:4284",
    "grpcAddress": "localhost:4283",
    "jwt": "your_jwt_token",
    "concurrency": 10,
    "repeat": 1
}
```

### Configuration Options

- `csv`: Path to the CSV file containing IDs to query
- `account`: Your Parker account name
- `table`: The table to query
- `httpAddress`: HTTP server address (for HTTP protocol)
- `grpcAddress`: gRPC server address (for gRPC protocol)
- `jwt`: JWT token for authentication (optional)
- `concurrency`: Number of concurrent workers
- `repeat`: Number of times to repeat the benchmark

## Usage

The benchmark tool can be run using Make targets:

### Python Implementation

```bash
# Run HTTP benchmark using Python
make benchmark-http

# Run gRPC benchmark using Python
make benchmark-grpc
```

### Go Implementation

```bash
# Run HTTP benchmark using Go
make benchmark-go-http

# Run gRPC benchmark using Go
make benchmark-go-grpc
```

## Output

The tool will display:
- Total number of requests processed
- Average latency
- P50, P95, and P99 latency percentiles
- Requests per second

## Example Output

```
Running benchmark: 100%|██████████| 10/10 [00:05<00:00, 1.98it/s]

Benchmark Results:
Total Requests: 1000
Average Latency: 50.23ms
P50 Latency: 45.67ms
P95 Latency: 75.89ms
P99 Latency: 89.12ms
Requests per Second: 198.45
```

## Development

### Building

```bash
# Generate Go protobuf files
make generate-go

# Clean generated files
make clean
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
