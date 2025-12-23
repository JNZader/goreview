// Package metrics provides performance metrics collection and export.
package metrics

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Collector collects and manages metrics.
type Collector struct {
	mu         sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	timers     map[string]*Timer
	startTime  time.Time
}

// NewCollector creates a new metrics collector.
func NewCollector() *Collector {
	return &Collector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		timers:     make(map[string]*Timer),
		startTime:  time.Now(),
	}
}

// Counter is a monotonically increasing counter.
type Counter struct {
	value int64
	mu    sync.Mutex
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

// Add adds n to the counter.
func (c *Counter) Add(n int64) {
	c.mu.Lock()
	c.value += n
	c.mu.Unlock()
}

// Value returns the current counter value.
func (c *Counter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Gauge represents a value that can go up or down.
type Gauge struct {
	value float64
	mu    sync.Mutex
}

// Set sets the gauge value.
func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	g.value = v
	g.mu.Unlock()
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	g.mu.Lock()
	g.value++
	g.mu.Unlock()
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	g.mu.Lock()
	g.value--
	g.mu.Unlock()
}

// Add adds a value to the gauge.
func (g *Gauge) Add(v float64) {
	g.mu.Lock()
	g.value += v
	g.mu.Unlock()
}

// Value returns the current gauge value.
func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Histogram collects distribution of values.
type Histogram struct {
	values []float64
	mu     sync.Mutex
	max    int
}

// NewHistogram creates a new histogram with max values capacity.
func NewHistogram(maxValues int) *Histogram {
	return &Histogram{
		values: make([]float64, 0, maxValues),
		max:    maxValues,
	}
}

// Observe records a value in the histogram.
func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.values) >= h.max {
		// Rotate: discard oldest
		h.values = h.values[1:]
	}
	h.values = append(h.values, v)
}

// Percentile returns the p-th percentile (0-100).
func (h *Histogram) Percentile(p float64) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.values) == 0 {
		return 0
	}

	sorted := make([]float64, len(h.values))
	copy(sorted, h.values)
	sort.Float64s(sorted)

	idx := int(float64(len(sorted)-1) * p / 100)
	return sorted[idx]
}

// Stats returns histogram statistics.
func (h *Histogram) Stats() HistogramStats {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.values) == 0 {
		return HistogramStats{}
	}

	sorted := make([]float64, len(h.values))
	copy(sorted, h.values)
	sort.Float64s(sorted)

	var sum float64
	for _, v := range sorted {
		sum += v
	}

	n := len(sorted)
	return HistogramStats{
		Count: n,
		Min:   sorted[0],
		Max:   sorted[n-1],
		Avg:   sum / float64(n),
		P50:   sorted[n*50/100],
		P90:   sorted[n*90/100],
		P99:   sorted[(n*99)/100],
	}
}

// HistogramStats contains histogram statistics.
type HistogramStats struct {
	Count int     `json:"count"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Avg   float64 `json:"avg"`
	P50   float64 `json:"p50"`
	P90   float64 `json:"p90"`
	P99   float64 `json:"p99"`
}

// Timer measures durations.
type Timer struct {
	histogram *Histogram
}

// Start starts a new timer context.
func (t *Timer) Start() *TimerContext {
	return &TimerContext{
		timer: t,
		start: time.Now(),
	}
}

// TimerContext represents an active timer.
type TimerContext struct {
	timer *Timer
	start time.Time
}

// Stop stops the timer and records the duration.
func (tc *TimerContext) Stop() time.Duration {
	d := time.Since(tc.start)
	tc.timer.histogram.Observe(d.Seconds())
	return d
}

// StopWithLabels stops the timer with additional metadata (for future use).
func (tc *TimerContext) StopWithLabels(_ map[string]string) time.Duration {
	return tc.Stop()
}

// Counter returns or creates a counter.
func (c *Collector) Counter(name string) *Counter {
	c.mu.Lock()
	defer c.mu.Unlock()

	if counter, ok := c.counters[name]; ok {
		return counter
	}

	counter := &Counter{}
	c.counters[name] = counter
	return counter
}

// Gauge returns or creates a gauge.
func (c *Collector) Gauge(name string) *Gauge {
	c.mu.Lock()
	defer c.mu.Unlock()

	if gauge, ok := c.gauges[name]; ok {
		return gauge
	}

	gauge := &Gauge{}
	c.gauges[name] = gauge
	return gauge
}

// Histogram returns or creates a histogram.
func (c *Collector) Histogram(name string) *Histogram {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hist, ok := c.histograms[name]; ok {
		return hist
	}

	hist := NewHistogram(1000)
	c.histograms[name] = hist
	return hist
}

// Timer returns or creates a timer.
func (c *Collector) Timer(name string) *Timer {
	c.mu.Lock()
	defer c.mu.Unlock()

	if timer, ok := c.timers[name]; ok {
		return timer
	}

	timer := &Timer{histogram: NewHistogram(1000)}
	c.timers[name] = timer
	return timer
}

// Uptime returns the duration since the collector was created.
func (c *Collector) Uptime() time.Duration {
	return time.Since(c.startTime)
}

// Export exports metrics to JSON.
func (c *Collector) Export() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	export := struct {
		Uptime     string                    `json:"uptime"`
		Counters   map[string]int64          `json:"counters"`
		Gauges     map[string]float64        `json:"gauges"`
		Histograms map[string]HistogramStats `json:"histograms"`
		Timers     map[string]HistogramStats `json:"timers"`
	}{
		Uptime:     time.Since(c.startTime).String(),
		Counters:   make(map[string]int64),
		Gauges:     make(map[string]float64),
		Histograms: make(map[string]HistogramStats),
		Timers:     make(map[string]HistogramStats),
	}

	for name, counter := range c.counters {
		export.Counters[name] = counter.Value()
	}

	for name, gauge := range c.gauges {
		export.Gauges[name] = gauge.Value()
	}

	for name, hist := range c.histograms {
		export.Histograms[name] = hist.Stats()
	}

	for name, timer := range c.timers {
		export.Timers[name] = timer.histogram.Stats()
	}

	return json.MarshalIndent(export, "", "  ")
}

// ExportPrometheus exports metrics in Prometheus format.
func (c *Collector) ExportPrometheus() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sb strings.Builder

	// Counters
	for name, counter := range c.counters {
		sb.WriteString(fmt.Sprintf("# TYPE %s counter\n", name))
		sb.WriteString(fmt.Sprintf("%s %d\n", name, counter.Value()))
	}

	// Gauges
	for name, gauge := range c.gauges {
		sb.WriteString(fmt.Sprintf("# TYPE %s gauge\n", name))
		sb.WriteString(fmt.Sprintf("%s %f\n", name, gauge.Value()))
	}

	// Histograms (as summary)
	for name, hist := range c.histograms {
		stats := hist.Stats()
		sb.WriteString(fmt.Sprintf("# TYPE %s summary\n", name))
		sb.WriteString(fmt.Sprintf("%s_count %d\n", name, stats.Count))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.5\"} %f\n", name, stats.P50))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.9\"} %f\n", name, stats.P90))
		sb.WriteString(fmt.Sprintf("%s{quantile=\"0.99\"} %f\n", name, stats.P99))
	}

	// Timers (as summary with _seconds suffix)
	for name, timer := range c.timers {
		stats := timer.histogram.Stats()
		sb.WriteString(fmt.Sprintf("# TYPE %s_seconds summary\n", name))
		sb.WriteString(fmt.Sprintf("%s_seconds_count %d\n", name, stats.Count))
		sb.WriteString(fmt.Sprintf("%s_seconds{quantile=\"0.5\"} %f\n", name, stats.P50))
		sb.WriteString(fmt.Sprintf("%s_seconds{quantile=\"0.9\"} %f\n", name, stats.P90))
		sb.WriteString(fmt.Sprintf("%s_seconds{quantile=\"0.99\"} %f\n", name, stats.P99))
	}

	return sb.String()
}

// Reset resets all metrics.
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counters = make(map[string]*Counter)
	c.gauges = make(map[string]*Gauge)
	c.histograms = make(map[string]*Histogram)
	c.timers = make(map[string]*Timer)
	c.startTime = time.Now()
}
