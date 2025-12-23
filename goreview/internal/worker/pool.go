// Package worker provides a worker pool for parallel task processing.
package worker

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

// Task represents a task to be executed by a worker.
type Task interface {
	Execute(ctx context.Context) error
	ID() string
}

// Result contains the result of a task execution.
type Result struct {
	TaskID string
	Error  error
}

// Pool manages a pool of workers for parallel processing.
type Pool struct {
	workers   int
	tasks     chan Task
	results   chan Result
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	started   atomic.Bool
	processed atomic.Int64
	errors    atomic.Int64
}

// Config configures the worker pool.
type Config struct {
	Workers   int // Number of workers (default: GOMAXPROCS)
	QueueSize int // Size of task queue (default: workers * 2)
}

// NewPool creates a new worker pool.
func NewPool(cfg Config) *Pool {
	if cfg.Workers <= 0 {
		cfg.Workers = runtime.GOMAXPROCS(0)
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = cfg.Workers * 2
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workers: cfg.Workers,
		tasks:   make(chan Task, cfg.QueueSize),
		results: make(chan Result, cfg.QueueSize),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the worker pool.
func (p *Pool) Start() {
	if p.started.Swap(true) {
		return // Already started
	}

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker is the goroutine that processes tasks.
func (p *Pool) worker(_ int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return

		case task, ok := <-p.tasks:
			if !ok {
				return
			}

			// Execute task
			err := task.Execute(p.ctx)

			// Record result
			p.processed.Add(1)
			if err != nil {
				p.errors.Add(1)
			}

			// Send result
			select {
			case p.results <- Result{TaskID: task.ID(), Error: err}:
			case <-p.ctx.Done():
				return
			}
		}
	}
}

// Submit submits a task to the pool.
func (p *Pool) Submit(task Task) error {
	if !p.started.Load() {
		return fmt.Errorf("pool not started")
	}

	select {
	case p.tasks <- task:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// SubmitWait submits a task and waits for its result.
func (p *Pool) SubmitWait(ctx context.Context, task Task) error {
	if err := p.Submit(task); err != nil {
		return err
	}

	for {
		select {
		case result := <-p.results:
			if result.TaskID == task.ID() {
				return result.Error
			}
			// Not our result, put it back (this is a simplification)
			// In real usage, you'd want a better coordination mechanism
		case <-ctx.Done():
			return ctx.Err()
		case <-p.ctx.Done():
			return p.ctx.Err()
		}
	}
}

// Results returns the results channel.
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Stop stops the pool gracefully.
func (p *Pool) Stop() {
	p.cancel()
	close(p.tasks)
	p.wg.Wait()
	close(p.results)
}

// StopWait stops the pool and waits for all tasks to complete.
func (p *Pool) StopWait() {
	close(p.tasks)
	p.wg.Wait()
	p.cancel()
	close(p.results)
}

// Stats returns pool statistics.
func (p *Pool) Stats() Stats {
	return Stats{
		Workers:   p.workers,
		Processed: p.processed.Load(),
		Errors:    p.errors.Load(),
		Pending:   len(p.tasks),
	}
}

// Stats contains pool statistics.
type Stats struct {
	Workers   int
	Processed int64
	Errors    int64
	Pending   int
}

// String returns a string representation of the stats.
func (s Stats) String() string {
	return fmt.Sprintf("workers=%d processed=%d errors=%d pending=%d",
		s.Workers, s.Processed, s.Errors, s.Pending)
}
