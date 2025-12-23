package review

import (
	"context"
	"runtime"
	"time"

	"github.com/JNZader/goreview/goreview/internal/metrics"
)

// InstrumentedEngine wraps Engine with metrics collection.
type InstrumentedEngine struct {
	engine    *Engine
	collector *metrics.Collector
}

// NewInstrumentedEngine creates an engine with metrics instrumentation.
func NewInstrumentedEngine(engine *Engine) *InstrumentedEngine {
	return &InstrumentedEngine{
		engine:    engine,
		collector: metrics.Global(),
	}
}

// NewInstrumentedEngineWithCollector creates an engine with a custom collector.
func NewInstrumentedEngineWithCollector(engine *Engine, collector *metrics.Collector) *InstrumentedEngine {
	return &InstrumentedEngine{
		engine:    engine,
		collector: collector,
	}
}

// Run executes the review with metrics collection.
func (ie *InstrumentedEngine) Run(ctx context.Context) (*Result, error) {
	// Increment reviews counter
	ie.collector.Counter(metrics.MetricReviewsTotal).Inc()

	// Timer for total duration
	timer := ie.collector.Timer(metrics.MetricReviewDuration).Start()
	defer timer.Stop()

	// Update memory metrics before
	ie.updateMemoryMetrics()

	// Execute review
	result, err := ie.engine.Run(ctx)

	// Record errors
	if err != nil {
		ie.collector.Counter(metrics.MetricErrors).Inc()
		return result, err
	}

	// Record result metrics
	if result != nil {
		ie.collector.Counter(metrics.MetricFilesProcessed).Add(int64(len(result.Files)))
		ie.collector.Counter(metrics.MetricIssuesFound).Add(int64(result.TotalIssues))

		// Count cached vs non-cached
		for _, f := range result.Files {
			if f.Cached {
				ie.collector.Counter(metrics.MetricCacheHits).Inc()
			} else {
				ie.collector.Counter(metrics.MetricCacheMisses).Inc()
			}
		}
	}

	// Update memory metrics after
	ie.updateMemoryMetrics()

	return result, nil
}

// updateMemoryMetrics updates memory and goroutine gauges.
func (ie *InstrumentedEngine) updateMemoryMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ie.collector.Gauge(metrics.MetricMemoryUsage).Set(float64(m.Alloc))
	ie.collector.Gauge(metrics.MetricGoroutines).Set(float64(runtime.NumGoroutine()))
}

// RecordProviderLatency records the latency of a provider request.
func (ie *InstrumentedEngine) RecordProviderLatency(duration time.Duration) {
	ie.collector.Histogram(metrics.MetricProviderLatency).Observe(duration.Seconds())
}

// RecordProviderRequest records a provider request.
func (ie *InstrumentedEngine) RecordProviderRequest() {
	ie.collector.Counter(metrics.MetricProviderRequests).Inc()
}

// RecordProviderError records a provider error.
func (ie *InstrumentedEngine) RecordProviderError() {
	ie.collector.Counter(metrics.MetricProviderErrors).Inc()
}

// RecordCacheHit records a cache hit.
func (ie *InstrumentedEngine) RecordCacheHit() {
	ie.collector.Counter(metrics.MetricCacheHits).Inc()
}

// RecordCacheMiss records a cache miss.
func (ie *InstrumentedEngine) RecordCacheMiss() {
	ie.collector.Counter(metrics.MetricCacheMisses).Inc()
}

// SetCacheSize sets the current cache size gauge.
func (ie *InstrumentedEngine) SetCacheSize(size int) {
	ie.collector.Gauge(metrics.MetricCacheSize).Set(float64(size))
}

// Metrics returns the metrics as JSON.
func (ie *InstrumentedEngine) Metrics() ([]byte, error) {
	return ie.collector.Export()
}

// MetricsPrometheus returns the metrics in Prometheus format.
func (ie *InstrumentedEngine) MetricsPrometheus() string {
	return ie.collector.ExportPrometheus()
}

// Stats returns a summary of review statistics.
func (ie *InstrumentedEngine) Stats() ReviewStats {
	return ReviewStats{
		TotalReviews:     ie.collector.Counter(metrics.MetricReviewsTotal).Value(),
		TotalFiles:       ie.collector.Counter(metrics.MetricFilesProcessed).Value(),
		TotalIssues:      ie.collector.Counter(metrics.MetricIssuesFound).Value(),
		TotalErrors:      ie.collector.Counter(metrics.MetricErrors).Value(),
		CacheHits:        ie.collector.Counter(metrics.MetricCacheHits).Value(),
		CacheMisses:      ie.collector.Counter(metrics.MetricCacheMisses).Value(),
		ProviderRequests: ie.collector.Counter(metrics.MetricProviderRequests).Value(),
		ProviderErrors:   ie.collector.Counter(metrics.MetricProviderErrors).Value(),
		MemoryBytes:      uint64(ie.collector.Gauge(metrics.MetricMemoryUsage).Value()),
		Goroutines:       int(ie.collector.Gauge(metrics.MetricGoroutines).Value()),
		Uptime:           ie.collector.Uptime(),
	}
}

// ReviewStats contains aggregate review statistics.
type ReviewStats struct {
	TotalReviews     int64         `json:"total_reviews"`
	TotalFiles       int64         `json:"total_files"`
	TotalIssues      int64         `json:"total_issues"`
	TotalErrors      int64         `json:"total_errors"`
	CacheHits        int64         `json:"cache_hits"`
	CacheMisses      int64         `json:"cache_misses"`
	ProviderRequests int64         `json:"provider_requests"`
	ProviderErrors   int64         `json:"provider_errors"`
	MemoryBytes      uint64        `json:"memory_bytes"`
	Goroutines       int           `json:"goroutines"`
	Uptime           time.Duration `json:"uptime"`
}

// CacheHitRate returns the cache hit rate as a percentage (0-100).
func (s ReviewStats) CacheHitRate() float64 {
	total := s.CacheHits + s.CacheMisses
	if total == 0 {
		return 0
	}
	return float64(s.CacheHits) / float64(total) * 100
}

// ProviderErrorRate returns the provider error rate as a percentage (0-100).
func (s ReviewStats) ProviderErrorRate() float64 {
	if s.ProviderRequests == 0 {
		return 0
	}
	return float64(s.ProviderErrors) / float64(s.ProviderRequests) * 100
}
