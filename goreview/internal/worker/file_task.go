package worker

import (
	"context"
	"fmt"
	"time"
)

// FileReviewer interface for reviewing files.
type FileReviewer interface {
	ReviewFile(ctx context.Context, path, content string) (*FileReviewResult, error)
}

// FileReviewResult is the result of reviewing a file.
type FileReviewResult struct {
	FilePath string
	Issues   []Issue
	Score    float64
	Duration time.Duration
}

// Issue represents an issue found during review.
type Issue struct {
	Line     int
	Column   int
	Severity string
	Message  string
	Rule     string
}

// FileTask represents a task to review a file.
type FileTask struct {
	id       string
	filePath string
	content  string
	reviewer FileReviewer
	result   *FileReviewResult
}

// NewFileTask creates a new file review task.
func NewFileTask(path, content string, reviewer FileReviewer) *FileTask {
	return &FileTask{
		id:       fmt.Sprintf("file:%s", path),
		filePath: path,
		content:  content,
		reviewer: reviewer,
	}
}

// ID returns the task identifier.
func (t *FileTask) ID() string {
	return t.id
}

// Execute executes the file review task.
func (t *FileTask) Execute(ctx context.Context) error {
	result, err := t.reviewer.ReviewFile(ctx, t.filePath, t.content)
	if err != nil {
		return err
	}
	t.result = result
	return nil
}

// Result returns the file review result.
func (t *FileTask) Result() *FileReviewResult {
	return t.result
}

// FilePath returns the file path being reviewed.
func (t *FileTask) FilePath() string {
	return t.filePath
}

// BatchTask represents a batch of tasks to be executed.
type BatchTask struct {
	id    string
	tasks []Task
}

// NewBatchTask creates a new batch task.
func NewBatchTask(id string, tasks []Task) *BatchTask {
	return &BatchTask{
		id:    id,
		tasks: tasks,
	}
}

// ID returns the batch task identifier.
func (b *BatchTask) ID() string {
	return b.id
}

// Execute executes all tasks in the batch sequentially.
func (b *BatchTask) Execute(ctx context.Context) error {
	for _, task := range b.tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := task.Execute(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// FuncTask wraps a function as a task.
type FuncTask struct {
	id string
	fn func(ctx context.Context) error
}

// NewFuncTask creates a task from a function.
func NewFuncTask(id string, fn func(ctx context.Context) error) *FuncTask {
	return &FuncTask{
		id: id,
		fn: fn,
	}
}

// ID returns the task identifier.
func (f *FuncTask) ID() string {
	return f.id
}

// Execute executes the function.
func (f *FuncTask) Execute(ctx context.Context) error {
	return f.fn(ctx)
}
