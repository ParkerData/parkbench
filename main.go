package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/ParkerData/parkbench/pb/parker_pb"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var ()

func main() {
	// Define CLI options
	grpcServerAddress := flag.String("grpcAddress", "", "gRPC server address")
	httpServerAddress := flag.String("httpAddress", "localhost:8250", "http server address")
	csvFilePath := flag.String("csv", "ids.csv", "Path to the CSV file with a list of IDs")
	concurrency := flag.Int("concurrency", 20, "Number of concurrent requests")
	indexColumn := flag.String("idColumn", "id", "id column name")
	repeatTimes := flag.Int("repeat", 1, "Number of times to repeat the test")
	jwtString := flag.String("jwt", "", "JWT token")

	flag.Parse()

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

	records = records[1:]
	println("input csv rows:", len(records))

	// Channel to distribute IDs to workers
	idChan := make(chan string, 10000)
	go func() {
		for x := 0; x < *repeatTimes; x++ {

			// Randomize the order of IDs
			total := len(records)
			for i := range records {
				j := rand.IntN(total)
				records[i], records[j] = records[j], records[i]
			}

			for _, record := range records {
				idChan <- record[0]
			}
		}
		close(idChan)
	}()

	// WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Channel to collect latencies
	latencyChan := make(chan time.Duration, 10000)

	// Start workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if *grpcServerAddress != "" {
				grpcQueryJob(*grpcServerAddress, *jwtString, idChan, latencyChan)
			} else if *httpServerAddress != "" {
				httpQueryJob(*httpServerAddress, *jwtString, *indexColumn, idChan, latencyChan)
			} else {
				log.Fatalf("Either gRPC or HTTP server address must be provided")
			}

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

func httpQueryJob(httpServerAddress string, jwtString string, idColumn string, idChan chan string, latencyChan chan time.Duration) {

	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        4,
			MaxIdleConnsPerHost: 4,
		},
	}

	for id := range idChan {
		start := time.Now()

		// send a http request to http://<httpServerAddress>/query?<idColumn>=<id>
		// and get the response
		targetUrl := fmt.Sprintf("http://%s/query?%s=%s", httpServerAddress, idColumn, id)
		req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
		if jwtString != "" {
			req.Header["Authorization"] = []string{"Bearer " + jwtString}
		}
		if err != nil {
			log.Fatalf("Failed to create HTTP request to %v: %v", targetUrl, err)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatalf("Failed to send HTTP request to %v: %v", targetUrl, err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Failed to get a successful response from %v: %v", targetUrl, resp.Status)
		}

		// read the response body
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		latency := time.Since(start)
		latencyChan <- latency
	}

}

func grpcQueryJob(grpcServerAddress string, jwtString string, idChan chan string, latencyChan chan time.Duration) {
	// Set up a gRPC client
	conn, err := grpc.Dial(grpcServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := parker_pb.NewParkerClient(conn)

	ctx := context.Background()
	if jwtString != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+jwtString)
	}

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
		_, err := client.Lookup(ctx, request)
		if err != nil {
			log.Fatalf("Failed to call Lookup: %v", err)
		}

		latency := time.Since(start)
		latencyChan <- latency
	}
}
