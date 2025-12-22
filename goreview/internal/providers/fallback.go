package providers

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// FallbackProvider wraps multiple providers with automatic failover.
type FallbackProvider struct {
	providers []Provider
	primary   int // Index of current primary provider
	mu        sync.RWMutex
}

// NewFallbackProvider creates a provider chain with automatic failover.
// Providers are tried in order: first one that works becomes primary.
func NewFallbackProvider(providers ...Provider) (*FallbackProvider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider required")
	}

	return &FallbackProvider{
		providers: providers,
		primary:   0,
	}, nil
}

func (f *FallbackProvider) Name() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return fmt.Sprintf("fallback(%s)", f.providers[f.primary].Name())
}

func (f *FallbackProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	f.mu.RLock()
	startIdx := f.primary
	f.mu.RUnlock()

	var lastErr error
	for i := 0; i < len(f.providers); i++ {
		idx := (startIdx + i) % len(f.providers)
		provider := f.providers[idx]

		resp, err := provider.Review(ctx, req)
		if err == nil {
			// Update primary if we fell back to a different provider
			if idx != startIdx {
				f.mu.Lock()
				f.primary = idx
				f.mu.Unlock()
				log.Printf("[fallback] Switched to provider: %s", provider.Name())
			}
			return resp, nil
		}

		lastErr = err
		log.Printf("[fallback] Provider %s failed: %v, trying next...", provider.Name(), err)
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

func (f *FallbackProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	f.mu.RLock()
	startIdx := f.primary
	f.mu.RUnlock()

	var lastErr error
	for i := 0; i < len(f.providers); i++ {
		idx := (startIdx + i) % len(f.providers)
		provider := f.providers[idx]

		msg, err := provider.GenerateCommitMessage(ctx, diff)
		if err == nil {
			return msg, nil
		}
		lastErr = err
		log.Printf("[fallback] Provider %s failed for commit msg: %v", provider.Name(), err)
	}

	return "", fmt.Errorf("all providers failed: %w", lastErr)
}

func (f *FallbackProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	f.mu.RLock()
	startIdx := f.primary
	f.mu.RUnlock()

	var lastErr error
	for i := 0; i < len(f.providers); i++ {
		idx := (startIdx + i) % len(f.providers)
		provider := f.providers[idx]

		doc, err := provider.GenerateDocumentation(ctx, diff, docContext)
		if err == nil {
			return doc, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("all providers failed: %w", lastErr)
}

func (f *FallbackProvider) HealthCheck(ctx context.Context) error {
	// Check all providers and report which ones are healthy
	var healthy []string
	var unhealthy []string

	for _, p := range f.providers {
		if err := p.HealthCheck(ctx); err != nil {
			unhealthy = append(unhealthy, p.Name())
		} else {
			healthy = append(healthy, p.Name())
		}
	}

	if len(healthy) == 0 {
		return fmt.Errorf("no healthy providers available, unhealthy: %v", unhealthy)
	}

	log.Printf("[fallback] Healthy providers: %v, Unhealthy: %v", healthy, unhealthy)
	return nil
}

func (f *FallbackProvider) Close() error {
	var errs []error
	for _, p := range f.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}
	return nil
}

// GetProviderStatus returns the status of all providers in the chain.
func (f *FallbackProvider) GetProviderStatus(ctx context.Context) map[string]bool {
	status := make(map[string]bool)
	for _, p := range f.providers {
		status[p.Name()] = p.HealthCheck(ctx) == nil
	}
	return status
}
