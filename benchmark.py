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
        response = session.get(url, headers=headers)
        response.raise_for_status()
        latency = time.time() - start
        results.add_latency(latency)

def grpc_query_job(config: BenchmarkConfig, ids: List[str], results: BenchmarkResults):
    # TODO: Implement gRPC client
    pass

def run_benchmark(config: BenchmarkConfig, use_grpc: bool = False):
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
    
    # Print results
    summary = results.get_summary()
    print("\nBenchmark Results:")
    print(f"Total Requests: {summary['total_requests']}")
    print(f"Average Latency: {summary['avg_latency_ms']:.2f}ms")
    print(f"P50 Latency: {summary['p50_latency_ms']:.2f}ms")
    print(f"P95 Latency: {summary['p95_latency_ms']:.2f}ms")
    print(f"P99 Latency: {summary['p99_latency_ms']:.2f}ms")
    print(f"Requests per Second: {summary['requests_per_second']:.2f}")

def main():
    parser = argparse.ArgumentParser(description='Run a benchmark against the Parker service')
    parser.add_argument('--config', default='config.json', help='Path to the configuration file')
    parser.add_argument('--grpc', action='store_true', help='Use gRPC protocol (default: HTTP)')
    args = parser.parse_args()
    
    config = BenchmarkConfig(args.config)
    run_benchmark(config, args.grpc)

if __name__ == '__main__':
    main() 