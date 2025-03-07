package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Metrics struct {
	mu           sync.Mutex
	Total        int
	SuccessCount int
	ErrorCount   int
	Latencies    []time.Duration
}

func main() {
	var (
		requests    = flag.Int("requests", 100, "Total requests to send")
		concurrency = flag.Int("concurrency", 10, "Number of concurrent workers")
		url         = flag.String("url", "http://localhost:8080/trigger", "Target URL")
		payload     = flag.String("payload", `{"data":"test"}`, "Request payload")
	)
	flag.Parse()

	metrics := &Metrics{Latencies: make([]time.Duration, 0, *requests)}
	start := time.Now()
	
	var wg sync.WaitGroup
	wg.Add(*requests)

	// Worker pool
	work := make(chan struct{}, *concurrency)
	for i := 0; i < *concurrency; i++ {
		go func() {
			client := &http.Client{Timeout: 5 * time.Second}
			for range work {
				startReq := time.Now()
				resp, err := client.Post(*url, "application/json", bytes.NewBufferString(*payload))
				
				metrics.mu.Lock()
				metrics.Total++
				if err != nil || (resp != nil && resp.StatusCode != http.StatusAccepted) {
					metrics.ErrorCount++
				} else {
					metrics.SuccessCount++
				}
				if resp != nil {
					resp.Body.Close()
				}
				metrics.Latencies = append(metrics.Latencies, time.Since(startReq))
				metrics.mu.Unlock()
				
				wg.Done()
			}
		}()
	}

	for i := 0; i < *requests; i++ {
		work <- struct{}{}
	}
	close(work)
	
	wg.Wait()
	printReport(metrics, time.Since(start), *concurrency)
}

func printReport(m *Metrics, duration time.Duration, concurrency int) {
	fmt.Printf(`
Load Test Report
================
Requests:      %d
Concurrency:   %d
Total Time:    %v
Req/sec:       %.2f

Success:       %d (%.2f%%)
Errors:        %d (%.2f%%)

Latency Stats:
  Average:     %v
  95th %%ile:   %v
  99th %%ile:   %v
`,
		m.Total,
		concurrency,
		duration.Round(time.Millisecond),
		float64(m.Total)/duration.Seconds(),
		m.SuccessCount, 100*float64(m.SuccessCount)/float64(m.Total),
		m.ErrorCount, 100*float64(m.ErrorCount)/float64(m.Total),
		averageLatency(m.Latencies),
		percentile(m.Latencies, 0.95),
		percentile(m.Latencies, 0.99),
	)
}

func percentile(latencies []time.Duration, p float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	index := int(float64(len(latencies)) * p)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	return latencies[index]
}

func averageLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	return total / time.Duration(len(latencies))
}