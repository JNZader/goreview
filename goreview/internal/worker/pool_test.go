package worker

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// mockTask for testing
type mockTask struct {
	id       string
	duration time.Duration
	err      error
}

func (t *mockTask) ID() string { return t.id }
func (t *mockTask) Execute(ctx context.Context) error {
	select {
	case <-time.After(t.duration):
		return t.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestPool_BasicExecution(t *testing.T) {
	pool := NewPool(Config{Workers: 2, QueueSize: 10})
	pool.Start()
	defer pool.Stop()

	// Submit tasks
	for i := 0; i < 5; i++ {
		task := &mockTask{
			id:       fmt.Sprintf("task-%d", i),
			duration: 10 * time.Millisecond,
		}
		if err := pool.Submit(task); err != nil {
			t.Fatalf("submit failed: %v", err)
		}
	}

	// Collect results
	results := 0
	timeout := time.After(1 * time.Second)
	for results < 5 {
		select {
		case r := <-pool.Results():
			if r.Error != nil {
				t.Errorf("unexpected error: %v", r.Error)
			}
			results++
		case <-timeout:
			t.Fatal("timeout waiting for results")
		}
	}

	stats := pool.Stats()
	if stats.Processed != 5 {
		t.Errorf("expected 5 processed, got %d", stats.Processed)
	}
}

func TestPool_ErrorHandling(t *testing.T) {
	pool := NewPool(Config{Workers: 2})
	pool.Start()
	defer pool.Stop()

	expectedErr := errors.New("task failed")
	task := &mockTask{
		id:       "failing-task",
		duration: 10 * time.Millisecond,
		err:      expectedErr,
	}

	pool.Submit(task)

	result := <-pool.Results()
	if result.Error != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, result.Error)
	}

	stats := pool.Stats()
	if stats.Errors != 1 {
		t.Errorf("expected 1 error, got %d", stats.Errors)
	}
}

func TestPool_Cancellation(t *testing.T) {
	pool := NewPool(Config{Workers: 2})
	pool.Start()

	// Submit long task
	task := &mockTask{
		id:       "long-task",
		duration: 10 * time.Second,
	}
	pool.Submit(task)

	// Cancel immediately
	pool.Stop()

	// Pool should stop without blocking
}

func TestPool_ConcurrentSubmit(t *testing.T) {
	pool := NewPool(Config{Workers: 4, QueueSize: 100})
	pool.Start()
	defer pool.Stop()

	var submitted atomic.Int64

	// Submit from multiple goroutines
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				task := &mockTask{
					id:       fmt.Sprintf("task-%d-%d", n, j),
					duration: time.Millisecond,
				}
				if err := pool.Submit(task); err == nil {
					submitted.Add(1)
				}
			}
		}(i)
	}

	// Wait for results
	time.Sleep(500 * time.Millisecond)

	stats := pool.Stats()
	if stats.Processed < 50 {
		t.Errorf("expected at least 50 processed, got %d", stats.Processed)
	}
}

func TestPool_NotStarted(t *testing.T) {
	pool := NewPool(Config{Workers: 2})
	// Don't start the pool

	task := &mockTask{id: "test"}
	err := pool.Submit(task)
	if err == nil {
		t.Error("expected error when submitting to unstarted pool")
	}
}

func TestPool_DoubleStart(t *testing.T) {
	pool := NewPool(Config{Workers: 2})
	pool.Start()
	pool.Start() // Should not panic or create duplicate workers
	pool.Stop()
}

func TestPool_DefaultConfig(t *testing.T) {
	pool := NewPool(Config{})

	if pool.workers != runtime.GOMAXPROCS(0) {
		t.Errorf("expected %d workers, got %d", runtime.GOMAXPROCS(0), pool.workers)
	}
}

func TestStats_String(t *testing.T) {
	stats := Stats{
		Workers:   4,
		Processed: 100,
		Errors:    5,
		Pending:   10,
	}

	str := stats.String()
	if str == "" {
		t.Error("Stats.String() should not return empty")
	}
}

func TestFuncTask(t *testing.T) {
	executed := false
	task := NewFuncTask("func-task", func(ctx context.Context) error {
		executed = true
		return nil
	})

	if task.ID() != "func-task" {
		t.Errorf("unexpected ID: %s", task.ID())
	}

	err := task.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("function was not executed")
	}
}

func TestBatchTask(t *testing.T) {
	var executed []string
	tasks := []Task{
		NewFuncTask("task-1", func(ctx context.Context) error {
			executed = append(executed, "task-1")
			return nil
		}),
		NewFuncTask("task-2", func(ctx context.Context) error {
			executed = append(executed, "task-2")
			return nil
		}),
	}

	batch := NewBatchTask("batch", tasks)
	if batch.ID() != "batch" {
		t.Errorf("unexpected ID: %s", batch.ID())
	}

	err := batch.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(executed) != 2 {
		t.Errorf("expected 2 tasks executed, got %d", len(executed))
	}
}

func TestBatchTask_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var executed int
	tasks := []Task{
		NewFuncTask("task-1", func(ctx context.Context) error {
			executed++
			return nil
		}),
		NewFuncTask("task-2", func(ctx context.Context) error {
			executed++
			return nil
		}),
	}

	batch := NewBatchTask("batch", tasks)
	err := batch.Execute(ctx)

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Benchmark
func BenchmarkPool_Throughput(b *testing.B) {
	pool := NewPool(Config{Workers: runtime.GOMAXPROCS(0), QueueSize: 1000})
	pool.Start()
	defer pool.Stop()

	// Consumer of results
	go func() {
		for range pool.Results() {
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		task := &mockTask{
			id:       fmt.Sprintf("task-%d", i),
			duration: 0, // Instant
		}
		pool.Submit(task)
	}
}

func BenchmarkPool_Concurrent(b *testing.B) {
	pool := NewPool(Config{Workers: runtime.GOMAXPROCS(0), QueueSize: 1000})
	pool.Start()
	defer pool.Stop()

	// Consumer
	go func() {
		for range pool.Results() {
		}
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			task := &mockTask{
				id:       fmt.Sprintf("task-%d", i),
				duration: 0,
			}
			pool.Submit(task)
			i++
		}
	})
}
