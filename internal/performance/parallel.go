package performance

import (
	"fmt"
	"sync"

	"github.com/jomadu/arm/internal/installer"
	"github.com/jomadu/arm/internal/registry"
	"github.com/schollz/progressbar/v3"
)

// DownloadJob represents a single ruleset download task
type DownloadJob struct {
	Name            string
	VersionSpec     string
	RegistryName    string
	CleanName       string
	RegistryManager *registry.Manager
}

// DownloadResult represents the result of a download job
type DownloadResult struct {
	Job   DownloadJob
	Error error
}

// ParallelDownloader manages concurrent downloads with progress tracking
type ParallelDownloader struct {
	registryManager *registry.Manager
}

// NewParallelDownloader creates a new parallel downloader
func NewParallelDownloader(registryManager *registry.Manager) *ParallelDownloader {
	return &ParallelDownloader{
		registryManager: registryManager,
	}
}

// DownloadAll downloads multiple rulesets in parallel with progress tracking
func (pd *ParallelDownloader) DownloadAll(jobs []DownloadJob) []DownloadResult {
	if len(jobs) == 0 {
		return nil
	}

	// Group jobs by registry to respect concurrency limits
	registryJobs := make(map[string][]DownloadJob)
	for _, job := range jobs {
		registryJobs[job.RegistryName] = append(registryJobs[job.RegistryName], job)
	}

	// Create progress bar
	bar := progressbar.Default(int64(len(jobs)), "Installing rulesets")

	// Channel to collect results
	resultsChan := make(chan DownloadResult, len(jobs))

	// Process each registry's jobs with its own concurrency limit
	var wg sync.WaitGroup
	for registryName, regJobs := range registryJobs {
		wg.Add(1)
		go func(regName string, regJobs []DownloadJob) {
			defer wg.Done()
			pd.processRegistryJobs(regName, regJobs, resultsChan, bar)
		}(registryName, regJobs)
	}

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []DownloadResult
	for result := range resultsChan {
		results = append(results, result)
	}

	_ = bar.Finish()
	return results
}

// processRegistryJobs processes jobs for a single registry with concurrency limits
func (pd *ParallelDownloader) processRegistryJobs(registryName string, jobs []DownloadJob, resultsChan chan<- DownloadResult, bar *progressbar.ProgressBar) {
	concurrency := pd.registryManager.GetConcurrency(registryName)
	semaphore := make(chan struct{}, concurrency)

	var wg sync.WaitGroup
	for _, job := range jobs {
		wg.Add(1)
		go func(j DownloadJob) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Perform download
			result := pd.downloadSingle(j)
			resultsChan <- result
			_ = bar.Add(1)
		}(job)
	}

	wg.Wait()
}

// downloadSingle downloads a single ruleset
func (pd *ParallelDownloader) downloadSingle(job DownloadJob) DownloadResult {
	installer := installer.NewWithManager(job.RegistryManager, job.RegistryName, job.CleanName)
	err := installer.Install(job.CleanName, job.VersionSpec)
	return DownloadResult{
		Job:   job,
		Error: err,
	}
}

// PrintResults prints the results of parallel downloads
func PrintResults(results []DownloadResult) error {
	var failures []DownloadResult
	successCount := 0

	for _, result := range results {
		if result.Error != nil {
			failures = append(failures, result)
		} else {
			successCount++
		}
	}

	// Print summary
	if len(failures) == 0 {
		fmt.Printf("Successfully installed %d rulesets\n", successCount)
		return nil
	}

	// Print failures
	fmt.Printf("Installed %d/%d rulesets successfully\n", successCount, len(results))
	fmt.Println("Failed installations:")
	for _, failure := range failures {
		fmt.Printf("  - %s@%s: %v\n", failure.Job.Name, failure.Job.VersionSpec, failure.Error)
	}

	return fmt.Errorf("%d installations failed", len(failures))
}
