package profiler

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew_CPUProfile(t *testing.T) {
	dir := t.TempDir()
	cpuFile := filepath.Join(dir, "cpu.prof")

	p, err := New(Config{
		CPUProfile: cpuFile,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Do some work
	sum := 0
	for i := 0; i < 100000; i++ {
		sum += i
	}
	_ = sum

	if err := p.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(cpuFile); os.IsNotExist(err) {
		t.Error("CPU profile file was not created")
	}
}

func TestNew_MemProfile(t *testing.T) {
	dir := t.TempDir()
	memFile := filepath.Join(dir, "mem.prof")

	p, err := New(Config{
		MemProfile: memFile,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Allocate some memory
	data := make([]byte, 1024*1024)
	_ = data

	if err := p.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Check file was created
	if _, err := os.Stat(memFile); os.IsNotExist(err) {
		t.Error("Memory profile file was not created")
	}
}

func TestNew_InvalidCPUPath(t *testing.T) {
	_, err := New(Config{
		CPUProfile: "/nonexistent/path/cpu.prof",
	})
	if err == nil {
		t.Error("Expected error for invalid CPU profile path")
	}
}

func TestStats(t *testing.T) {
	stats := Stats()

	if stats.Alloc == 0 {
		t.Error("Alloc should be > 0")
	}
	if stats.Sys == 0 {
		t.Error("Sys should be > 0")
	}
	if stats.HeapAlloc == 0 {
		t.Error("HeapAlloc should be > 0")
	}
}

func TestMemStats_String(t *testing.T) {
	stats := MemStats{
		Alloc:     1024 * 1024,
		HeapAlloc: 512 * 1024,
		Sys:       10 * 1024 * 1024,
		NumGC:     5,
	}

	str := stats.String()
	if str == "" {
		t.Error("String() should not return empty")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}

	for _, tc := range tests {
		result := formatBytes(tc.bytes)
		if result != tc.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", tc.bytes, result, tc.expected)
		}
	}
}

func TestProfiler_Duration(t *testing.T) {
	p, err := New(Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	duration := p.Duration()
	if duration < 10*time.Millisecond {
		t.Errorf("Duration() = %v, expected >= 10ms", duration)
	}

	p.Stop()
}
