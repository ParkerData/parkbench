import argparse
import csv
import json
import random
import time
from concurrent.futures import ThreadPoolExecutor
from typing import Dict, List, Optional
import requests
import grpc
import pandas as pd
import numpy as np
from tqdm import tqdm
from pb.gateway_pb2 import FindRequest, Key
from pb.gateway_pb2_grpc import GatewayStub

class BenchmarkConfig:
    def __init__(self, config_path: str):
        with open(config_path) as f:
            config = json.load(f)
        
        self.csv_file_path = config.get('csv')
        self.account_name = config.get('account')
        self.table_name = config.get('table')
        self.http_server_address = config.get('httpAddress')
        self.grpc_server_address = config.get('grpcAddress')
        self.jwt_string = config.get('jwt')
        self.concurrency = config.get('concurrency', 10)
        self.repeat_times = config.get('repeat', 1)

class BenchmarkResults:
    def __init__(self):
        self.latencies = []
        self.requests_per_second = []
        self.start_time = time.time()
        self.total_requests = 0

    def add_latency(self, latency: float):
        self.latencies.append(latency)
        self.total_requests += 1

    def get_summary(self) -> Dict:
        return {
            'total_requests': self.total_requests,
            'avg_latency_ms': np.mean(self.latencies) * 1000,
            'p50_latency_ms': np.percentile(self.latencies, 50) * 1000,
            'p95_latency_ms': np.percentile(self.latencies, 95) * 1000,
            'p99_latency_ms': np.percentile(self.latencies, 99) * 1000,
            'requests_per_second': self.total_requests / (time.time() - self.start_time)
        }

def http_query_job(config: BenchmarkConfig, ids: List[str], results: BenchmarkResults):
    session = requests.Session()
    headers = {}
    if config.jwt_string:
        headers['Authorization'] = f'Bearer {config.jwt_string}'

    for id in ids:
        start = time.time()
        url = f"{config.http_server_address}/find/{config.account_name}/{config.table_name}/{id}"
        try:
            response = session.get(url, headers=headers, timeout=30)
            response.raise_for_status()
            latency = time.time() - start
            results.add_latency(latency)
        except Exception as e:
            print(f"Error: {e}")
            raise

def grpc_query_job(config: BenchmarkConfig, ids: List[str], results: BenchmarkResults):
    # Set up a secure gRPC client using TLS
    channel = grpc.secure_channel(
        config.grpc_server_address,
        grpc.ssl_channel_credentials()
    )
    
    # Create the gRPC client
    client = GatewayStub(channel)
    
    # Set up metadata with JWT if provided
    metadata = []
    if config.jwt_string:
        metadata.append(('authorization', f'Bearer {config.jwt_string}'))
    
    for id in ids:
        start = time.time()
        try:
            # Create the FindRequest
            request = FindRequest(
                account=config.account_name,
                table=config.table_name,
                key=Key(string_value=id)
            )
            
            # Call the Find method
            response = client.Find(request, metadata=metadata)
            
            # Record latency
            latency = time.time() - start
            results.add_latency(latency)
        except Exception as e:
            print(f"Error: {e}")
            raise

def run_benchmark(config: BenchmarkConfig, use_grpc: bool = False):
    try:
        # Read IDs from CSV
        df = pd.read_csv(config.csv_file_path)
        ids = df.iloc[:, 0].tolist()
        
        # Create results object
        results = BenchmarkResults()
        
        # Calculate batch size for each worker
        total_ids = len(ids) * config.repeat_times
        batch_size = total_ids // config.concurrency
        
        # Create batches of IDs
        all_ids = []
        for _ in range(config.repeat_times):
            random.shuffle(ids)
            all_ids.extend(ids)
        
        id_batches = [all_ids[i:i + batch_size] for i in range(0, len(all_ids), batch_size)]
        
        # Start periodic logging
        stop_logging = False
        def log_stats():
            last_time = time.time()
            last_requests = 0
            while not stop_logging:
                time.sleep(1)
                current_time = time.time()
                current_requests = results.total_requests
                elapsed = current_time - last_time
                if elapsed > 0:
                    rps = (current_requests - last_requests) / elapsed
                    if current_requests > 0:
                        avg_latency = np.mean(results.latencies[-1000:]) * 1000  # Last 1000 requests
                        print(f"Requests per second: {rps:.0f}, Average latency: {avg_latency:.2f}ms")
                    last_time = current_time
                    last_requests = current_requests
        
        import threading
        logging_thread = threading.Thread(target=log_stats)
        logging_thread.start()
        
        # Run benchmark
        with ThreadPoolExecutor(max_workers=config.concurrency) as executor:
            futures = []
            for batch in id_batches:
                if use_grpc:
                    futures.append(executor.submit(grpc_query_job, config, batch, results))
                else:
                    futures.append(executor.submit(http_query_job, config, batch, results))
            
            # Wait for all futures to complete
            for future in tqdm(futures, desc="Running benchmark"):
                future.result()
        
        # Stop logging thread
        stop_logging = True
        logging_thread.join()
        
        # Print final results
        summary = results.get_summary()
        print("\nBenchmark Results:")
        print(f"Total Requests: {summary['total_requests']}")
        print(f"Average Latency: {summary['avg_latency_ms']:.2f}ms")
        print(f"P50 Latency: {summary['p50_latency_ms']:.2f}ms")
        print(f"P95 Latency: {summary['p95_latency_ms']:.2f}ms")
        print(f"P99 Latency: {summary['p99_latency_ms']:.2f}ms")
        print(f"Requests per Second: {summary['requests_per_second']:.2f}")
    except Exception as e:
        print(f"Error: {e}")
        raise

def main():
    parser = argparse.ArgumentParser(description='Run a benchmark against the Parker service')
    parser.add_argument('--config', default='config.json', help='Path to the configuration file')
    parser.add_argument('--grpc', action='store_true', help='Use gRPC protocol (default: HTTP)')
    args = parser.parse_args()
    
    config = BenchmarkConfig(args.config)
    run_benchmark(config, args.grpc)

if __name__ == '__main__':
    main() 