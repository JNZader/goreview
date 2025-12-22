package commands

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressReporter handles CLI progress output.
type ProgressReporter struct {
	total     int
	current   int
	mu        sync.Mutex
	startTime time.Time
	spinner   int
	done      chan struct{}
}

// NewProgressReporter creates a new progress reporter.
func NewProgressReporter(total int) *ProgressReporter {
	return &ProgressReporter{
		total:     total,
		startTime: time.Now(),
		done:      make(chan struct{}),
	}
}

// Start begins the progress display.
func (p *ProgressReporter) Start() {
	go func() {
		spinnerChars := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-p.done:
				return
			case <-ticker.C:
				p.mu.Lock()
				spinner := spinnerChars[p.spinner%len(spinnerChars)]
				p.spinner++
				progress := float64(p.current) / float64(p.total) * 100
				bar := p.renderBar(progress)
				fmt.Fprintf(os.Stderr, "\r%c Reviewing files %s %.0f%% (%d/%d)",
					spinner, bar, progress, p.current, p.total)
				p.mu.Unlock()
			}
		}
	}()
}

// Increment advances the progress by one.
func (p *ProgressReporter) Increment(filename string) {
	p.mu.Lock()
	p.current++
	p.mu.Unlock()
}

// Finish completes the progress display.
func (p *ProgressReporter) Finish() {
	close(p.done)
	duration := time.Since(p.startTime)
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "Reviewed %d files in %s\n", p.total, duration.Round(time.Millisecond))
}

func (p *ProgressReporter) renderBar(progress float64) string {
	width := 20
	filled := int(progress / 100 * float64(width))
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

// PrintSummary prints a summary of the review results.
func PrintSummary(totalIssues int, files int, duration time.Duration) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "  Review Complete\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "  Files reviewed: %d\n", files)
	fmt.Fprintf(os.Stderr, "  Issues found:   %d\n", totalIssues)
	fmt.Fprintf(os.Stderr, "  Duration:       %s\n", duration.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}
