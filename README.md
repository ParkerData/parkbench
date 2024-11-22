# parkbench
Benchmark Parker point lookup service

Parker is an ultra-low-latency and high-concurrency point query service for Parquet files.
You can test the point query API on your own parquet files with this docker image https://hub.docker.com/r/parkerdb/parker-preview

This repo is for a tool to benchmark the point query performance.

## Install

Download the latest release from the [releases page](https://github.com/ParkerData/parkbench/releases) and decompress it.

Or install it using `go install` if you have Go installed:
```shell
go install github.com/ParkerData/parkbench@latest
```

## Run

```shell
$ parkbench -h
Usage of parkbench:
  -concurrency int
        Number of concurrent requests (default 20)
  -csv string
        Path to the CSV file with a list of IDs (default "ids.csv")
  -httpAddress string
        http server address (default "localhost:8250")
  -idColumn string
        id column name (default "id")
  -repeat int
        Number of times to repeat the test (default 1)


$ parkbench -httpAddress localhost:8250 -csv ids.csv -idColumn id -concurrency 20
...

```
