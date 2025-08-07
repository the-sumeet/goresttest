package main

import (
	"fmt"
	"sync"
	"time"
)

type BenchmarkExecutor struct {
	client *HTTPClient
}

func NewBenchmarkExecutor(baseURL string) *BenchmarkExecutor {
	return &BenchmarkExecutor{
		client: NewHTTPClient(baseURL),
	}
}

func (be *BenchmarkExecutor) ExecuteBenchmarks(benchmarks []Benchmark, variables map[string]string) ([]*BenchmarkResult, error) {
	var results []*BenchmarkResult
	
	for _, benchmark := range benchmarks {
		result, err := be.executeBenchmark(benchmark, variables)
		if err != nil {
			return nil, fmt.Errorf("benchmark %s failed: %w", benchmark.Name, err)
		}
		results = append(results, result)
	}
	
	return results, nil
}

func (be *BenchmarkExecutor) executeBenchmark(benchmark Benchmark, variables map[string]string) (*BenchmarkResult, error) {
	if benchmark.Concurrent <= 0 {
		benchmark.Concurrent = 1
	}
	
	var totalRequests int
	var duration time.Duration
	
	if benchmark.Requests > 0 {
		totalRequests = benchmark.Requests
		duration = 0
	} else if benchmark.Duration > 0 {
		totalRequests = 0
		duration = benchmark.Duration
	} else {
		return nil, fmt.Errorf("either requests or duration must be specified")
	}
	
	start := time.Now()
	
	var result *BenchmarkResult
	var err error
	
	if totalRequests > 0 {
		result, err = be.executeFixedRequests(benchmark, variables, totalRequests)
	} else {
		result, err = be.executeTimedRequests(benchmark, variables, duration)
	}
	
	if err != nil {
		return nil, err
	}
	
	result.Name = benchmark.Name
	result.TotalTime = time.Since(start)
	
	if result.TotalRequests > 0 {
		result.RequestsPerSec = float64(result.TotalRequests) / result.TotalTime.Seconds()
		result.AvgResponseTime = time.Duration(int64(result.TotalTime) / int64(result.TotalRequests))
	}
	
	return result, nil
}

func (be *BenchmarkExecutor) executeFixedRequests(benchmark Benchmark, variables map[string]string, totalRequests int) (*BenchmarkResult, error) {
	requestsPerWorker := totalRequests / benchmark.Concurrent
	extraRequests := totalRequests % benchmark.Concurrent
	
	var wg sync.WaitGroup
	resultChan := make(chan RequestResult, totalRequests)
	
	for i := 0; i < benchmark.Concurrent; i++ {
		wg.Add(1)
		requests := requestsPerWorker
		if i < extraRequests {
			requests++
		}
		
		go func(numRequests int) {
			defer wg.Done()
			be.workerFixedRequests(benchmark, variables, numRequests, resultChan)
		}(requests)
	}
	
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	return be.collectResults(resultChan), nil
}

func (be *BenchmarkExecutor) executeTimedRequests(benchmark Benchmark, variables map[string]string, duration time.Duration) (*BenchmarkResult, error) {
	var wg sync.WaitGroup
	resultChan := make(chan RequestResult, 10000)
	stopChan := make(chan struct{})
	
	for i := 0; i < benchmark.Concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			be.workerTimedRequests(benchmark, variables, stopChan, resultChan)
		}()
	}
	
	time.AfterFunc(duration, func() {
		close(stopChan)
	})
	
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	return be.collectResults(resultChan), nil
}

func (be *BenchmarkExecutor) workerFixedRequests(benchmark Benchmark, variables map[string]string, requests int, resultChan chan<- RequestResult) {
	test := Test{
		Method:  benchmark.Method,
		URL:     benchmark.URL,
		Headers: benchmark.Headers,
		Body:    benchmark.Body,
	}
	
	for i := 0; i < requests; i++ {
		start := time.Now()
		result, err := be.client.ExecuteRequest(test, variables)
		duration := time.Since(start)
		
		requestResult := RequestResult{
			Success:      err == nil && result.Success,
			Duration:     duration,
			StatusCode:   result.StatusCode,
			ResponseSize: len(result.Response),
		}
		
		resultChan <- requestResult
	}
}

func (be *BenchmarkExecutor) workerTimedRequests(benchmark Benchmark, variables map[string]string, stopChan <-chan struct{}, resultChan chan<- RequestResult) {
	test := Test{
		Method:  benchmark.Method,
		URL:     benchmark.URL,
		Headers: benchmark.Headers,
		Body:    benchmark.Body,
	}
	
	for {
		select {
		case <-stopChan:
			return
		default:
			start := time.Now()
			result, err := be.client.ExecuteRequest(test, variables)
			duration := time.Since(start)
			
			requestResult := RequestResult{
				Success:      err == nil && result.Success,
				Duration:     duration,
				StatusCode:   result.StatusCode,
				ResponseSize: len(result.Response),
			}
			
			resultChan <- requestResult
		}
	}
}

func (be *BenchmarkExecutor) collectResults(resultChan <-chan RequestResult) *BenchmarkResult {
	var results []RequestResult
	
	for result := range resultChan {
		results = append(results, result)
	}
	
	if len(results) == 0 {
		return &BenchmarkResult{}
	}
	
	benchResult := &BenchmarkResult{
		TotalRequests: len(results),
	}
	
	var totalDuration time.Duration
	minDuration := results[0].Duration
	maxDuration := results[0].Duration
	
	for _, result := range results {
		totalDuration += result.Duration
		
		if result.Success {
			benchResult.SuccessfulReqs++
		} else {
			benchResult.FailedReqs++
		}
		
		if result.Duration < minDuration {
			minDuration = result.Duration
		}
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
	}
	
	benchResult.MinResponseTime = minDuration
	benchResult.MaxResponseTime = maxDuration
	
	return benchResult
}

type RequestResult struct {
	Success      bool
	Duration     time.Duration
	StatusCode   int
	ResponseSize int
}