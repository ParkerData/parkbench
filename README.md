# parkbench
Benchmark Parker point lookup service

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


$ parkbench -httpAddress localhost:8250 -csv ids.csv -idColumn id -concurrency 20
...

```