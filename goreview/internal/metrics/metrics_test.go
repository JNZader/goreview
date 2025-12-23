package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	c := NewCollector()

	counter := c.Counter("test_counter")
	counter.Inc()
	counter.Inc()
	counter.Add(5)

	if counter.Value() != 7 {
		t.Errorf("expected 7, got %d", counter.Value())
	}
}

func TestCounter_Concurrent(t *testing.T) {
	counter := &Counter{}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				counter.Inc()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if counter.Value() != 1000 {
		t.Errorf("expected 1000, got %d", counter.Value())
	}
}

func TestGauge(t *testing.T) {
	c := NewCollector()

	gauge := c.Gauge("test_gauge")
	gauge.Set(42.5)

	if gauge.Value() != 42.5 {
		t.Errorf("expected 42.5, got %f", gauge.Value())
	}

	gauge.Set(10.0)
	if gauge.Value() != 10.0 {
		t.Errorf("expected 10.0, got %f", gauge.Value())
	}

	gauge.Inc()
	if gauge.Value() != 11.0 {
		t.Errorf("expected 11.0, got %f", gauge.Value())
	}

	gauge.Dec()
	if gauge.Value() != 10.0 {
		t.Errorf("expected 10.0, got %f", gauge.Value())
	}

	gauge.Add(5.5)
	if gauge.Value() != 15.5 {
		t.Errorf("expected 15.5, got %f", gauge.Value())
	}
}

func TestHistogram(t *testing.T) {
	hist := NewHistogram(100)

	// Add values 1-100
	for i := 1; i <= 100; i++ {
		hist.Observe(float64(i))
	}

	stats := hist.Stats()

	if stats.Count != 100 {
		t.Errorf("expected 100 count, got %d", stats.Count)
	}
	if stats.Min != 1 {
		t.Errorf("expected min 1, got %f", stats.Min)
	}
	if stats.Max != 100 {
		t.Errorf("expected max 100, got %f", stats.Max)
	}
	if stats.Avg != 50.5 {
		t.Errorf("expected avg 50.5, got %f", stats.Avg)
	}
}

func TestHistogram_Rotation(t *testing.T) {
	hist := NewHistogram(10)

	// Add 15 values to trigger rotation
	for i := 1; i <= 15; i++ {
		hist.Observe(float64(i))
	}

	stats := hist.Stats()

	if stats.Count != 10 {
		t.Errorf("expected 10 count after rotation, got %d", stats.Count)
	}
	// Oldest values (1-5) should be discarded
	if stats.Min != 6 {
		t.Errorf("expected min 6 after rotation, got %f", stats.Min)
	}
}

func TestHistogram_Percentile(t *testing.T) {
	hist := NewHistogram(100)

	for i := 1; i <= 100; i++ {
		hist.Observe(float64(i))
	}

	p50 := hist.Percentile(50)
	if p50 < 49 || p50 > 51 {
		t.Errorf("expected p50 around 50, got %f", p50)
	}

	p90 := hist.Percentile(90)
	if p90 < 89 || p90 > 91 {
		t.Errorf("expected p90 around 90, got %f", p90)
	}
}

func TestHistogram_Empty(t *testing.T) {
	hist := NewHistogram(100)

	stats := hist.Stats()
	if stats.Count != 0 {
		t.Errorf("expected 0 count, got %d", stats.Count)
	}

	p50 := hist.Percentile(50)
	if p50 != 0 {
		t.Errorf("expected 0 for empty histogram, got %f", p50)
	}
}

func TestTimer(t *testing.T) {
	c := NewCollector()
	timer := c.Timer("test_timer")

	// Simulate operation
	ctx := timer.Start()
	time.Sleep(10 * time.Millisecond)
	duration := ctx.Stop()

	if duration < 10*time.Millisecond {
		t.Errorf("expected at least 10ms, got %v", duration)
	}

	stats := timer.histogram.Stats()
	if stats.Count != 1 {
		t.Errorf("expected 1 measurement, got %d", stats.Count)
	}
}

func TestCollector_SameMetric(t *testing.T) {
	c := NewCollector()

	counter1 := c.Counter("test")
	counter1.Inc()

	counter2 := c.Counter("test")
	counter2.Inc()

	if counter1.Value() != 2 {
		t.Errorf("expected same counter instance, got %d", counter1.Value())
	}
}

func TestExportJSON(t *testing.T) {
	c := NewCollector()

	c.Counter("files_processed").Add(10)
	c.Gauge("memory_mb").Set(256.5)
	c.Histogram("response_time").Observe(0.5)

	data, err := c.Export()
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("export returned empty data")
	}

	// Verify JSON structure
	str := string(data)
	if !strings.Contains(str, "files_processed") {
		t.Error("missing counter in JSON export")
	}
	if !strings.Contains(str, "memory_mb") {
		t.Error("missing gauge in JSON export")
	}
}

func TestExportPrometheus(t *testing.T) {
	c := NewCollector()

	c.Counter("http_requests_total").Add(100)
	c.Gauge("goroutines").Set(50)
	c.Histogram("request_duration").Observe(0.1)
	c.Histogram("request_duration").Observe(0.2)

	output := c.ExportPrometheus()

	if output == "" {
		t.Error("prometheus export returned empty")
	}

	if !strings.Contains(output, "http_requests_total") {
		t.Error("missing counter in output")
	}
	if !strings.Contains(output, "goroutines") {
		t.Error("missing gauge in output")
	}
	if !strings.Contains(output, "request_duration_count") {
		t.Error("missing histogram in output")
	}
}

func TestReset(t *testing.T) {
	c := NewCollector()

	c.Counter("test").Inc()
	c.Gauge("test").Set(42)

	c.Reset()

	// After reset, new metrics should be created
	if c.Counter("test").Value() != 0 {
		t.Error("counter should be 0 after reset")
	}
}

func TestUptime(t *testing.T) {
	c := NewCollector()
	time.Sleep(10 * time.Millisecond)

	uptime := c.Uptime()
	if uptime < 10*time.Millisecond {
		t.Errorf("expected uptime >= 10ms, got %v", uptime)
	}
}

func TestGlobalMetrics(t *testing.T) {
	// These should not panic
	IncCounter("global_test_counter")
	AddCounter("global_test_counter", 5)
	SetGauge("global_test_gauge", 42)
	IncGauge("global_test_gauge")
	DecGauge("global_test_gauge")
	ObserveHistogram("global_test_hist", 0.5)
	ctx := StartTimer("global_test_timer")
	ctx.Stop()

	// Verify global collector works
	if Global().Counter("global_test_counter").Value() != 6 {
		t.Error("global counter not working")
	}
}

// Benchmarks

func BenchmarkCounter_Inc(b *testing.B) {
	counter := &Counter{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}

func BenchmarkHistogram_Observe(b *testing.B) {
	hist := NewHistogram(10000)
	b.RunParallel(func(pb *testing.PB) {
		i := 0.0
		for pb.Next() {
			hist.Observe(i)
			i++
		}
	})
}

func BenchmarkTimer_StartStop(b *testing.B) {
	c := NewCollector()
	timer := c.Timer("bench_timer")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := timer.Start()
		ctx.Stop()
	}
}

func BenchmarkCollector_Export(b *testing.B) {
	c := NewCollector()

	// Add some metrics
	for i := 0; i < 10; i++ {
		c.Counter("counter_" + string(rune('a'+i))).Add(int64(i * 10))
		c.Gauge("gauge_" + string(rune('a'+i))).Set(float64(i))
		for j := 0; j < 100; j++ {
			c.Histogram("hist_" + string(rune('a'+i))).Observe(float64(j))
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Export()
	}
}
