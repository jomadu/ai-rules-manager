package install

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// InstallOrchestrator coordinates parallel installation operations
type InstallOrchestrator struct {
	installer    *Installer
	rateLimiters map[string]*TokenBucket
	mu           sync.RWMutex
}

// MultiInstallRequest represents multiple ruleset installation requests
type MultiInstallRequest struct {
	Requests []InstallRequest
	Progress ProgressCallback
}

// MultiInstallResult represents the results of multiple installations
type MultiInstallResult struct {
	Successful []InstallResult
	Failed     []InstallError
	Total      int
}

// InstallError represents a failed installation
type InstallError struct {
	Registry string
	Ruleset  string
	Error    error
}

// ProgressCallback is called to report installation progress
type ProgressCallback func(current, total int, operation string)

// TokenBucket implements token bucket rate limiting
type TokenBucket struct {
	tokens     int
	capacity   int
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewInstallOrchestrator creates a new parallel install orchestrator
func NewInstallOrchestrator(installer *Installer) *InstallOrchestrator {
	return &InstallOrchestrator{
		installer:    installer,
		rateLimiters: make(map[string]*TokenBucket),
	}
}

// InstallMultiple installs multiple rulesets in parallel with rate limiting
func (o *InstallOrchestrator) InstallMultiple(ctx context.Context, req *MultiInstallRequest) (*MultiInstallResult, error) {
	if len(req.Requests) == 0 {
		return &MultiInstallResult{Total: 0}, nil
	}

	// Group requests by registry for concurrency control
	registryGroups := o.groupByRegistry(req.Requests)

	// Initialize rate limiters for each registry
	o.initRateLimiters(registryGroups)

	// Create result channels
	resultChan := make(chan InstallResult, len(req.Requests))
	errorChan := make(chan InstallError, len(req.Requests))

	var wg sync.WaitGroup
	completed := 0
	total := len(req.Requests)
	var progressMu sync.Mutex

	// Process each registry group in parallel
	for registry, requests := range registryGroups {
		wg.Add(1)
		go func(reg string, reqs []InstallRequest) {
			defer wg.Done()
			o.processRegistryGroup(reg, reqs, resultChan, errorChan, req.Progress, &completed, total, &progressMu)
		}(registry, requests)
	}

	// Wait for all operations to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	result := &MultiInstallResult{Total: total}
	for res := range resultChan {
		result.Successful = append(result.Successful, res)
	}
	for err := range errorChan {
		result.Failed = append(result.Failed, err)
	}

	return result, nil
}

// groupByRegistry groups install requests by registry
func (o *InstallOrchestrator) groupByRegistry(requests []InstallRequest) map[string][]InstallRequest {
	groups := make(map[string][]InstallRequest)
	for i := range requests {
		req := &requests[i]
		groups[req.Registry] = append(groups[req.Registry], *req)
	}
	return groups
}

// initRateLimiters initializes rate limiters for registries
func (o *InstallOrchestrator) initRateLimiters(registryGroups map[string][]InstallRequest) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for registry := range registryGroups {
		if _, exists := o.rateLimiters[registry]; !exists {
			o.rateLimiters[registry] = o.createRateLimiter(registry)
		}
	}
}

// createRateLimiter creates a rate limiter for a registry
func (o *InstallOrchestrator) createRateLimiter(registry string) *TokenBucket {
	// Get rate limit from config
	rateLimit := o.getRateLimit(registry)
	capacity, refillRate := o.parseRateLimit(rateLimit)

	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// getRateLimit gets rate limit for a registry
func (o *InstallOrchestrator) getRateLimit(registry string) string {
	// Check registry-specific config
	if registryConfig := o.installer.config.RegistryConfigs[registry]; registryConfig != nil {
		if rateLimit := registryConfig["rateLimit"]; rateLimit != "" {
			return rateLimit
		}
	}

	// Check type defaults
	if registryConfig := o.installer.config.RegistryConfigs[registry]; registryConfig != nil {
		if registryType := registryConfig["type"]; registryType != "" {
			if typeDefaults := o.installer.config.TypeDefaults[registryType]; typeDefaults != nil {
				if rateLimit := typeDefaults["rateLimit"]; rateLimit != "" {
					return rateLimit
				}
			}
		}
	}

	// Default rate limit
	return "10/minute"
}

// parseRateLimit parses rate limit string (e.g., "10/minute", "100/hour")
func (o *InstallOrchestrator) parseRateLimit(rateLimit string) (int, time.Duration) {
	parts := strings.Split(rateLimit, "/")
	if len(parts) != 2 {
		return 10, time.Minute // Default
	}

	capacity, err := strconv.Atoi(parts[0])
	if err != nil {
		capacity = 10
	}

	var duration time.Duration
	switch parts[1] {
	case "second":
		duration = time.Second
	case "minute":
		duration = time.Minute
	case "hour":
		duration = time.Hour
	default:
		duration = time.Minute
	}

	refillRate := duration / time.Duration(capacity)
	return capacity, refillRate
}

// processRegistryGroup processes requests for a single registry with concurrency control
func (o *InstallOrchestrator) processRegistryGroup(registry string, requests []InstallRequest,
	resultChan chan<- InstallResult, errorChan chan<- InstallError, progress ProgressCallback, completed *int, total int, progressMu *sync.Mutex) {

	// Get concurrency limit for this registry
	concurrency := o.getConcurrency(registry)

	// Create semaphore for concurrency control
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := range requests {
		wg.Add(1)
		go func(request *InstallRequest) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Wait for rate limit
			o.waitForRateLimit(registry)

			// Report progress (thread-safe)
			progressMu.Lock()
			*completed++
			currentCompleted := *completed
			progressMu.Unlock()
			if progress != nil {
				progress(currentCompleted, total, fmt.Sprintf("Installing %s/%s", request.Registry, request.Ruleset))
			}

			// Perform installation
			result, err := o.installer.Install(request)
			if err != nil {
				errorChan <- InstallError{
					Registry: request.Registry,
					Ruleset:  request.Ruleset,
					Error:    err,
				}
			} else {
				resultChan <- *result
			}
		}(&requests[i])
	}

	wg.Wait()
}

// getConcurrency gets concurrency limit for a registry
func (o *InstallOrchestrator) getConcurrency(registry string) int {
	// Check registry-specific config
	if registryConfig := o.installer.config.RegistryConfigs[registry]; registryConfig != nil {
		if concurrencyStr := registryConfig["concurrency"]; concurrencyStr != "" {
			if concurrency, err := strconv.Atoi(concurrencyStr); err == nil {
				return concurrency
			}
		}
	}

	// Check type defaults
	if registryConfig := o.installer.config.RegistryConfigs[registry]; registryConfig != nil {
		if registryType := registryConfig["type"]; registryType != "" {
			if typeDefaults := o.installer.config.TypeDefaults[registryType]; typeDefaults != nil {
				if concurrencyStr := typeDefaults["concurrency"]; concurrencyStr != "" {
					if concurrency, err := strconv.Atoi(concurrencyStr); err == nil {
						return concurrency
					}
				}
			}
		}
	}

	// Default concurrency
	return 1
}

// waitForRateLimit waits for rate limit token
func (o *InstallOrchestrator) waitForRateLimit(registry string) {
	o.mu.RLock()
	bucket := o.rateLimiters[registry]
	o.mu.RUnlock()

	if bucket == nil {
		return
	}

	for !bucket.TakeToken() {
		time.Sleep(100 * time.Millisecond)
	}
}

// TakeToken attempts to take a token from the bucket
func (tb *TokenBucket) TakeToken() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed / tb.refillRate)

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Take token if available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}
