// Package profiler provides profiling utilities for CPU and memory analysis
package profiler

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // G108: pprof is intentionally exposed when explicitly enabled via --pprof-addr flag
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Profiler handles profile collection
type Profiler struct {
	cpuFile    *os.File
	memFile    string
	httpServer *http.Server
	startTime  time.Time
}

// Config configures the profiler
type Config struct {
	CPUProfile  string // File for CPU profile
	MemProfile  string // File for memory profile
	HTTPAddr    string // Address for pprof HTTP (e.g., ":6060")
	EnableTrace bool   // Enable tracing
}

// New creates a new profiler
func New(cfg Config) (*Profiler, error) {
	p := &Profiler{
		memFile:   cfg.MemProfile,
		startTime: time.Now(),
	}

	// Start CPU profiling if file specified
	if cfg.CPUProfile != "" {
		f, err := os.Create(cfg.CPUProfile)
		if err != nil {
			return nil, fmt.Errorf("failed to create CPU profile: %w", err)
		}
		p.cpuFile = f

		if err := pprof.StartCPUProfile(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to start CPU profile: %w", err)
		}
	}

	// Start HTTP server for pprof
	if cfg.HTTPAddr != "" {
		p.httpServer = &http.Server{
			Addr:         cfg.HTTPAddr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		}

		go func() {
			// pprof is already registered via import
			if err := p.httpServer.ListenAndServe(); err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "pprof server error: %v\n", err)
			}
		}()
	}

	return p, nil
}

// Stop stops profiling and saves results
func (p *Profiler) Stop() error {
	var errs []error

	// Stop CPU profiling
	if p.cpuFile != nil {
		pprof.StopCPUProfile()
		if err := p.cpuFile.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close CPU profile: %w", err))
		}
	}

	// Write memory profile
	if p.memFile != "" {
		// Force GC for accurate stats
		runtime.GC()

		f, err := os.Create(p.memFile)
		if err != nil {
			errs = append(errs, fmt.Errorf("create memory profile: %w", err))
		} else {
			defer f.Close()
			if err := pprof.WriteHeapProfile(f); err != nil {
				errs = append(errs, fmt.Errorf("write memory profile: %w", err))
			}
		}
	}

	// Stop HTTP server
	if p.httpServer != nil {
		if err := p.httpServer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close pprof server: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("profiler stop errors: %v", errs)
	}
	return nil
}

// Duration returns the time since profiler started
func (p *Profiler) Duration() time.Duration {
	return time.Since(p.startTime)
}

// Stats returns current memory statistics
func Stats() MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
		HeapAlloc:  m.HeapAlloc,
		HeapSys:    m.HeapSys,
		HeapIdle:   m.HeapIdle,
		HeapInuse:  m.HeapInuse,
	}
}

// MemStats contains memory statistics
type MemStats struct {
	Alloc      uint64 // Currently allocated bytes
	TotalAlloc uint64 // Total bytes allocated (cumulative)
	Sys        uint64 // Memory obtained from OS
	NumGC      uint32 // Number of GC runs
	HeapAlloc  uint64 // Heap bytes allocated
	HeapSys    uint64 // Heap obtained from OS
	HeapIdle   uint64 // Heap idle bytes
	HeapInuse  uint64 // Heap in-use bytes
}

// String formats the statistics
func (m MemStats) String() string {
	return fmt.Sprintf(
		"Alloc: %s, HeapAlloc: %s, Sys: %s, NumGC: %d",
		formatBytes(m.Alloc),
		formatBytes(m.HeapAlloc),
		formatBytes(m.Sys),
		m.NumGC,
	)
}

// formatBytes converts bytes to human-readable format
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
