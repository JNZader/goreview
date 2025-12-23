package review

import (
	"testing"
	"time"

	"github.com/JNZader/goreview/goreview/internal/metrics"
)

func TestInstrumentedEngine_New(t *testing.T) {
	engine := &Engine{}
	ie := NewInstrumentedEngine(engine)

	if ie.engine != engine {
		t.Error("engine not set correctly")
	}
	if ie.collector == nil {
		t.Error("collector should not be nil")
	}
}

func TestInstrumentedEngine_WithCollector(t *testing.T) {
	engine := &Engine{}
	collector := metrics.NewCollector()
	ie := NewInstrumentedEngineWithCollector(engine, collector)

	if ie.collector != collector {
		t.Error("custom collector not set correctly")
	}
}

func TestInstrumentedEngine_RecordMetrics(t *testing.T) {
	engine := &Engine{}
	collector := metrics.NewCollector()
	ie := NewInstrumentedEngineWithCollector(engine, collector)

	// Test provider latency
	ie.RecordProviderLatency(100 * time.Millisecond)

	// Test provider request/error
	ie.RecordProviderRequest()
	ie.RecordProviderRequest()
	ie.RecordProviderError()

	// Test cache
	ie.RecordCacheHit()
	ie.RecordCacheHit()
	ie.RecordCacheMiss()

	// Test cache size
	ie.SetCacheSize(42)

	// Verify metrics
	stats := ie.Stats()

	if stats.ProviderRequests != 2 {
		t.Errorf("expected 2 provider requests, got %d", stats.ProviderRequests)
	}
	if stats.ProviderErrors != 1 {
		t.Errorf("expected 1 provider error, got %d", stats.ProviderErrors)
	}
	if stats.CacheHits != 2 {
		t.Errorf("expected 2 cache hits, got %d", stats.CacheHits)
	}
	if stats.CacheMisses != 1 {
		t.Errorf("expected 1 cache miss, got %d", stats.CacheMisses)
	}
}

func TestInstrumentedEngine_Metrics(t *testing.T) {
	engine := &Engine{}
	collector := metrics.NewCollector()
	ie := NewInstrumentedEngineWithCollector(engine, collector)

	ie.RecordProviderRequest()

	data, err := ie.Metrics()
	if err != nil {
		t.Fatalf("Metrics() failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Metrics() returned empty data")
	}
}

func TestInstrumentedEngine_MetricsPrometheus(t *testing.T) {
	engine := &Engine{}
	collector := metrics.NewCollector()
	ie := NewInstrumentedEngineWithCollector(engine, collector)

	ie.RecordProviderRequest()

	output := ie.MetricsPrometheus()
	if output == "" {
		t.Error("MetricsPrometheus() returned empty")
	}
}

func TestEngineStats_CacheHitRate(t *testing.T) {
	tests := []struct {
		hits   int64
		misses int64
		want   float64
	}{
		{0, 0, 0},
		{10, 0, 100},
		{0, 10, 0},
		{5, 5, 50},
		{3, 7, 30},
	}

	for _, tc := range tests {
		stats := EngineStats{
			CacheHits:   tc.hits,
			CacheMisses: tc.misses,
		}
		got := stats.CacheHitRate()
		if got != tc.want {
			t.Errorf("CacheHitRate(%d, %d) = %f, want %f", tc.hits, tc.misses, got, tc.want)
		}
	}
}

func TestEngineStats_ProviderErrorRate(t *testing.T) {
	tests := []struct {
		requests int64
		errors   int64
		want     float64
	}{
		{0, 0, 0},
		{10, 0, 0},
		{10, 1, 10},
		{10, 5, 50},
		{100, 100, 100},
	}

	for _, tc := range tests {
		stats := EngineStats{
			ProviderRequests: tc.requests,
			ProviderErrors:   tc.errors,
		}
		got := stats.ProviderErrorRate()
		if got != tc.want {
			t.Errorf("ProviderErrorRate(%d, %d) = %f, want %f", tc.requests, tc.errors, got, tc.want)
		}
	}
}

func TestInstrumentedEngine_UpdateMemoryMetrics(t *testing.T) {
	engine := &Engine{}
	collector := metrics.NewCollector()
	ie := NewInstrumentedEngineWithCollector(engine, collector)

	ie.updateMemoryMetrics()

	stats := ie.Stats()

	if stats.MemoryBytes == 0 {
		t.Error("MemoryBytes should be > 0")
	}
	if stats.Goroutines == 0 {
		t.Error("Goroutines should be > 0")
	}
}

func TestInstrumentedEngine_Stats_Uptime(t *testing.T) {
	engine := &Engine{}
	collector := metrics.NewCollector()
	ie := NewInstrumentedEngineWithCollector(engine, collector)

	time.Sleep(10 * time.Millisecond)

	stats := ie.Stats()
	if stats.Uptime < 10*time.Millisecond {
		t.Errorf("Uptime should be >= 10ms, got %v", stats.Uptime)
	}
}
