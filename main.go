package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/ParkerData/parkbench/pb/parker_pb"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var (
	httpClient = &http.Client{}
)

func main() {
	// Define CLI options
	// grpcServerAddress := flag.String("grpcAddress", "", "gRPC server address")
	httpServerAddress := flag.String("httpAddress", "localhost:8250", "http server address")
	csvFilePath := flag.String("csv", "ids.csv", "Path to the CSV file with a list of IDs")
	concurrency := flag.Int("concurrency", 20, "Number of concurrent requests")
	indexColumn := flag.String("idColumn", "id", "id column name")
	flag.Parse()

	httpClient.

	// Read the CSV file
	file, err := os.Open(*csvFilePath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	// Extract IDs from the first column
	var ids []string
	for i, record := range records {
		if len(record) > 0 && i > 0 {
			ids = append(ids, record[0])
		}
	}

	// Randomize the order of IDs
	for i := range ids {
		j := i + int(time.Now().UnixNano())%(len(ids)-i)
		ids[i], ids[j] = ids[j], ids[i]
	}

	println("Randomized of IDs: ", len(ids))

	// Channel to distribute IDs to workers
	idChan := make(chan string, 10000)
	go func() {
		for _, id := range ids {
			idChan <- id
		}
		close(idChan)
	}()

	// WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Channel to collect latencies
	latencyChan := make(chan time.Duration, len(ids))

	// Start workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			httpQueryJob(*httpServerAddress, *indexColumn, idChan, latencyChan)

			//if *grpcServerAddress != "" {
			//	grpcQueryJob(*grpcServerAddress, idChan, latencyChan)
			//} else if *httpServerAddress != "" {
			//	httpQueryJob(*httpServerAddress, *indexColumn, idChan, latencyChan)
			//} else {
			//	log.Fatalf("Either gRPC or HTTP server address must be provided")
			//}

		}()
	}

	// Monitor and print latency and requests per second
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		var totalRequests int
		var totalLatency time.Duration

		for {
			select {
			case latency := <-latencyChan:
				totalRequests++
				totalLatency += latency
			case <-ticker.C:
				if totalRequests > 0 {
					avgLatency := totalLatency / time.Duration(totalRequests)
					fmt.Printf("Requests per second: %d, Average latency: %v\n", totalRequests, avgLatency)
					totalRequests = 0
					totalLatency = 0
				}
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
	close(latencyChan)
}

func httpQueryJob(httpServerAddress string, idColumn string, idChan chan string, latencyChan chan time.Duration) {

	for id := range idChan {
		start := time.Now()

		// send a http request to http://<httpServerAddress>/query?<idColumn>=<id>
		// and get the response
		targetUrl := fmt.Sprintf("http://%s/query?%s=%s", httpServerAddress, idColumn, id)
		req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
		if err != nil {
			log.Fatalf("Failed to create HTTP request to %v: %v", targetUrl, err)
		}
		resp, err = httpClient.Do(req)
		if err != nil {
			log.Fatalf("Failed to send HTTP request to %v: %v", targetUrl, err)
		}

		// read the response body
		io.ReadAll(resp.Body)

		latency := time.Since(start)
		latencyChan <- latency
	}

}

func grpcQueryJob(grpcServerAddress string, idChan chan string, latencyChan chan time.Duration) {
	// Set up a gRPC client
	conn, err := grpc.Dial(grpcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := parker_pb.NewParkerClient(conn)

	for id := range idChan {
		start := time.Now()

		// Create a LookupRequest
		request := &parker_pb.LookupRequest{
			Key: &parker_pb.Key{
				Kind: &parker_pb.Key_StringValue{
					StringValue: id,
				},
			},
		}

		// Call the Lookup method
		_, err := client.Lookup(context.Background(), request)
		if err != nil {
			log.Fatalf("Failed to call Lookup: %v", err)
		}

		latency := time.Since(start)
		latencyChan <- latency
	}
}
