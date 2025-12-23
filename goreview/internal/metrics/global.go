package metrics

import "sync"

var (
	globalCollector *Collector
	once            sync.Once
)

// Global returns the global metrics collector.
func Global() *Collector {
	once.Do(func() {
		globalCollector = NewCollector()
	})
	return globalCollector
}

// Convenience functions for quick access

// IncCounter increments a global counter by 1.
func IncCounter(name string) {
	Global().Counter(name).Inc()
}

// AddCounter adds n to a global counter.
func AddCounter(name string, n int64) {
	Global().Counter(name).Add(n)
}

// SetGauge sets a global gauge value.
func SetGauge(name string, v float64) {
	Global().Gauge(name).Set(v)
}

// IncGauge increments a global gauge by 1.
func IncGauge(name string) {
	Global().Gauge(name).Inc()
}

// DecGauge decrements a global gauge by 1.
func DecGauge(name string) {
	Global().Gauge(name).Dec()
}

// ObserveHistogram observes a value in a global histogram.
func ObserveHistogram(name string, v float64) {
	Global().Histogram(name).Observe(v)
}

// StartTimer starts a global timer.
func StartTimer(name string) *TimerContext {
	return Global().Timer(name).Start()
}

// Metric names for goreview
const (
	// Review metrics
	MetricReviewsTotal   = "goreview_reviews_total"
	MetricReviewDuration = "goreview_review_duration"
	MetricFilesProcessed = "goreview_files_processed_total"
	MetricFilesSkipped   = "goreview_files_skipped_total"
	MetricIssuesFound    = "goreview_issues_found_total"

	// Provider metrics
	MetricProviderRequests = "goreview_provider_requests_total"
	MetricProviderErrors   = "goreview_provider_errors_total"
	MetricProviderLatency  = "goreview_provider_latency"

	// Cache metrics
	MetricCacheHits   = "goreview_cache_hits_total"
	MetricCacheMisses = "goreview_cache_misses_total"
	MetricCacheSize   = "goreview_cache_size"

	// System metrics
	MetricMemoryUsage = "goreview_memory_bytes"
	MetricGoroutines  = "goreview_goroutines"
	MetricErrors      = "goreview_errors_total"
)
