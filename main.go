package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/httptrace"
	"os"
	"sync"
	"time"

	"github.com/ParkerData/parkbench/config"
	parker_pb "github.com/ParkerData/parkbench/pb/parker_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Define CLI option for config file path and protocol
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	useGRPC := flag.Bool("grpc", false, "Use gRPC protocol (default: HTTP)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create shared HTTP client if using HTTP
	var httpClient *http.Client
	if !*useGRPC {
		httpClient = &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 1000,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
			Timeout: 120 * time.Second,
		}
	}

	// Read the CSV file
	file, err := os.Open(cfg.CSVFilePath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	println("input csv rows:", len(records))

	// Channel to distribute IDs to workers
	idChan := make(chan string, 10000)
	go func() {
		for x := 0; x < cfg.RepeatTimes; x++ {
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
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if *useGRPC {
				if cfg.GRPCServerAddress == "" {
					log.Fatalf("gRPC server address not provided in config")
				}
				grpcQueryJob(cfg, idChan, latencyChan)
			} else {
				if cfg.HTTPServerAddress == "" {
					log.Fatalf("HTTP server address not provided in config")
				}
				httpQueryJob(httpClient, cfg.HTTPServerAddress, cfg.JWTString, idChan, latencyChan, cfg)
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

func tracedRequestContext() context.Context {
	return httptrace.WithClientTrace(context.Background(), &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			// fmt.Println("🧪 Connection info:")
			// fmt.Println("  Reused:  ", info.Reused)
			// fmt.Println("  WasIdle:", info.WasIdle)
			// fmt.Println("  IdleTime:", info.IdleTime)
		},
	})
}

func httpQueryJob(httpClient *http.Client, httpServerAddress string, jwtString string, idChan chan string, latencyChan chan time.Duration, cfg *config.Config) {
	count := 0
	for id := range idChan {
		start := time.Now()

		targetUrl := fmt.Sprintf("%s/find/%s/%s/%s", httpServerAddress, cfg.AccountName, cfg.TableName, id)
		req, err := http.NewRequestWithContext(tracedRequestContext(), http.MethodGet, targetUrl, nil)
		if err != nil {
			log.Fatalf("Failed to create HTTP request to %v: %v", targetUrl, err)
		}

		if jwtString != "" {
			req.Header["Authorization"] = []string{"Bearer " + jwtString}
		}
		req.Close = false

		count++
		// fmt.Printf("%d: Resolved URL: %s\n", count, targetUrl)
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

func grpcQueryJob(cfg *config.Config, idChan chan string, latencyChan chan time.Duration) {
	// Set up a secure gRPC client using TLS
	creds := credentials.NewClientTLSFromCert(nil, "") // nil means use system's trusted CAs

	// Set up a gRPC client
	conn, err := grpc.NewClient(cfg.GRPCServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := parker_pb.NewGatewayClient(conn)

	ctx := context.Background()
	if cfg.JWTString != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+cfg.JWTString)
	}

	for id := range idChan {
		start := time.Now()

		// Create a FindRequest
		request := &parker_pb.FindRequest{
			Account: cfg.AccountName,
			Table:   cfg.TableName,
			Key: &parker_pb.Key{
				Kind: &parker_pb.Key_StringValue{
					StringValue: id,
				},
			},
		}

		// Call the Find method
		_, err := client.Find(ctx, request)
		if err != nil {
			log.Fatalf("Failed to call Find: %v", err)
		}

		latency := time.Since(start)
		latencyChan <- latency
	}
}
