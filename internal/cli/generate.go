package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/copyleftdev/faux-foundry/internal/dedup"
	"github.com/copyleftdev/faux-foundry/internal/llm"
	"github.com/copyleftdev/faux-foundry/internal/output"
	"github.com/copyleftdev/faux-foundry/internal/spec"
	"github.com/copyleftdev/faux-foundry/pkg/types"
)

var (
	specFile    string
	outputFile  string
	count       int
	timeout     string
	seed        int64
	dryRun      bool
	interactive bool
	maxRetries  int
	minBatchSize int
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate synthetic data from specification",
	Long: `Generate synthetic data from a YAML specification file. The command will parse the 
specification, connect to the configured LLM backend, and generate exactly N unique records 
as specified in the configuration.

Examples:
  # Basic generation
  fauxfoundry generate --spec customer.yaml

  # Override record count and specify output
  fauxfoundry generate --spec customer.yaml --count 5000 --output data.jsonl.gz

  # Dry run to validate specification
  fauxfoundry generate --spec customer.yaml --dry-run

  # Interactive mode with TUI
  fauxfoundry generate --interactive`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVarP(&specFile, "spec", "s", "", "path to YAML specification file (required)")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path (stdout if not specified)")
	generateCmd.Flags().IntVarP(&count, "count", "n", 0, "override record count from specification")
	generateCmd.Flags().StringVarP(&timeout, "timeout", "t", "2h", "maximum execution time")
	generateCmd.Flags().Int64Var(&seed, "seed", 0, "random seed for reproducibility")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "validate specification without generating data")
	generateCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "launch interactive TUI mode")
	generateCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "maximum number of retry attempts on timeout")
	generateCmd.Flags().IntVar(&minBatchSize, "min-batch-size", 1, "minimum batch size before giving up")

	// Mark required flags
	generateCmd.MarkFlagRequired("spec")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if interactive {
		return runInteractiveGenerate()
	}

	// Validate inputs
	if specFile == "" {
		return fmt.Errorf("specification file is required")
	}

	if !fileExists(specFile) {
		return fmt.Errorf("specification file not found: %s", specFile)
	}

	// Parse timeout
	timeoutDuration, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("invalid timeout format: %s", timeout)
	}

	// Load and validate specification
	specification, err := spec.LoadFromFile(specFile)
	if err != nil {
		return fmt.Errorf("failed to load specification: %w", err)
	}

	// Override count if specified
	if count > 0 {
		specification.Dataset.Count = count
	}

	// Validate specification
	if err := spec.Validate(specification); err != nil {
		return fmt.Errorf("specification validation failed: %w", err)
	}

	if dryRun {
		fmt.Printf("✓ Specification is valid\n")
		fmt.Printf("  Domain: %s\n", specification.Dataset.Domain)
		fmt.Printf("  Fields: %d\n", len(specification.Dataset.Fields))
		fmt.Printf("  Target records: %d\n", specification.Dataset.Count)
		fmt.Printf("  Model: %s\n", specification.Model.Name)
		fmt.Printf("  Endpoint: %s\n", specification.Model.Endpoint)
		return nil
	}

	// Determine output path
	output := outputFile
	if output == "" {
		output = "stdout"
	} else {
		// Ensure output directory exists
		if dir := filepath.Dir(output); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}
		}
	}

	// Create generation job
	job := &types.GenerationJob{
		ID:         generateJobID(),
		Spec:       *specification,
		Status:     types.JobStatusPending,
		StartTime:  time.Now(),
		OutputPath: output,
		Progress: types.Progress{
			Target: specification.Dataset.Count,
		},
	}

	if !IsQuiet() {
		fmt.Printf("Starting data generation...\n")
		fmt.Printf("  Specification: %s\n", specFile)
		fmt.Printf("  Output: %s\n", output)
		fmt.Printf("  Target records: %d\n", job.Progress.Target)
		fmt.Printf("  Timeout: %s\n", timeoutDuration)
		if seed != 0 {
			fmt.Printf("  Seed: %d\n", seed)
		}
		fmt.Println()
	}

	// Start actual generation
	return runGeneration(job, timeoutDuration)
}

func runInteractiveGenerate() error {
	// TODO: Launch TUI mode
	return fmt.Errorf("interactive mode not yet implemented - use 'fauxfoundry tui' instead")
}

func runGeneration(job *types.GenerationJob, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Initialize components with custom retry config
	retryConfig := &llm.RetryConfig{
		MaxRetries:            maxRetries,
		BaseTimeout:           30 * time.Second,
		MaxTimeout:            5 * time.Minute,
		BackoffMultiplier:     2.0,
		ReduceFactorOnTimeout: 0.5,
		MinBatchSize:          minBatchSize,
	}
	
	llmClient := llm.NewOllamaClient()
	deduplicator := dedup.NewBatchDeduplicator(job.Spec.Model.BatchSize)
	
	// Create output writer
	writer, err := output.NewStreamingWriter(job.OutputPath, 100) // Buffer 100 records
	if err != nil {
		return fmt.Errorf("failed to create output writer: %w", err)
	}
	defer writer.Close()

	// Check Ollama health before generation
	health, err := llmClient.CheckOllamaHealth(ctx, job.Spec.Model.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to check Ollama health: %w", err)
	}

	if !health.IsRunning {
		if !IsQuiet() {
			fmt.Printf("❌ Ollama is not running at %s\n", job.Spec.Model.Endpoint)
			fmt.Printf("💡 Run 'fauxfoundry doctor' for setup instructions\n")
		}
		return fmt.Errorf("Ollama is not running: %s", health.ErrorMessage)
	}

	// Check if the specified model is available
	modelAvailable := false
	for _, model := range health.Models {
		if model == job.Spec.Model.Name {
			modelAvailable = true
			break
		}
	}

	if !modelAvailable {
		if !IsQuiet() {
			fmt.Printf("❌ Model '%s' is not available\n", job.Spec.Model.Name)
			fmt.Printf("📋 Available models: %v\n", health.Models)
			fmt.Printf("💡 Install the model: ollama pull %s\n", job.Spec.Model.Name)
		}
		return fmt.Errorf("model '%s' is not available", job.Spec.Model.Name)
	}

	if !IsQuiet() {
		fmt.Printf("🚀 Generation started (Job ID: %s)\n", job.ID)
		fmt.Printf("🤖 Connected to %s at %s\n", job.Spec.Model.Name, job.Spec.Model.Endpoint)
		fmt.Println()
	}

	// Generation loop
	job.Status = types.JobStatusRunning
	startTime := time.Now()
	generated := 0
	batchCount := 0

	for generated < job.Progress.Target {
		select {
		case <-ctx.Done():
			return fmt.Errorf("generation timed out after %s", timeout)
		default:
		}

		// Calculate batch size for this iteration
		remaining := job.Progress.Target - generated
		batchSize := job.Spec.Model.BatchSize
		if remaining < batchSize {
			batchSize = remaining
		}

		batchCount++
		if !IsQuiet() {
			fmt.Printf("📦 Generating batch %d (%d records)...\n", batchCount, batchSize)
		}

		// Generate batch with custom retry config
		records, err := llmClient.GenerateWithConfig(ctx, &job.Spec, batchSize, retryConfig)
		if err != nil {
			return fmt.Errorf("failed to generate batch %d: %w", batchCount, err)
		}

		// Deduplicate records
		uniqueRecords := deduplicator.ProcessBatch(records)
		
		// Write unique records
		for _, record := range uniqueRecords {
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write record: %w", err)
			}
			generated++
			
			// Update progress
			job.Progress.Generated = generated
			elapsed := time.Since(startTime)
			job.Progress.ElapsedTime = elapsed.String()
			
			if generated > 0 {
				rate := float64(generated) / elapsed.Seconds()
				job.Progress.Rate = rate
				
				if rate > 0 {
					remaining := job.Progress.Target - generated
					eta := time.Duration(float64(remaining)/rate) * time.Second
					job.Progress.EstimatedETA = eta.String()
				}
			}
		}

		// Show progress
		if !IsQuiet() {
			stats := deduplicator.GetStats()
			progress := float64(generated) / float64(job.Progress.Target) * 100
			fmt.Printf("📈 Progress: %.1f%% (%d/%d records) | %s\n", 
				progress, generated, job.Progress.Target, stats.String())
		}

		// Break if we've reached the target
		if generated >= job.Progress.Target {
			break
		}
	}

	// Finalize
	job.Status = types.JobStatusCompleted
	endTime := time.Now()
	job.EndTime = &endTime
	
	// Final flush
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	if !IsQuiet() {
		totalTime := endTime.Sub(startTime)
		finalStats := deduplicator.GetStats()
		
		fmt.Printf("\n✅ Generation completed successfully!\n")
		fmt.Printf("📊 Final Statistics:\n")
		fmt.Printf("   • Records generated: %d\n", generated)
		fmt.Printf("   • %s\n", finalStats.String())
		fmt.Printf("   • Total time: %s\n", totalTime.String())
		fmt.Printf("   • Average rate: %.2f records/second\n", float64(generated)/totalTime.Seconds())
		fmt.Printf("📁 Output written to: %s\n", writer.GetPath())
	}

	return nil
}

func simulateGeneration(job *types.GenerationJob, timeout time.Duration) error {
	if !IsQuiet() {
		fmt.Printf("🚀 Generation started (Job ID: %s)\n", job.ID)
		fmt.Printf("⏱️  This is a simulation - actual implementation coming soon!\n")
		fmt.Printf("📊 Would generate %d records to: %s\n", job.Progress.Target, job.OutputPath)
		fmt.Printf("🤖 Would use model: %s at %s\n", job.Spec.Model.Name, job.Spec.Model.Endpoint)
		fmt.Println()
		
		// Simulate some progress
		for i := 0; i < 5; i++ {
			time.Sleep(200 * time.Millisecond)
			progress := (i + 1) * 20
			fmt.Printf("📈 Progress: %d%% (%d/%d records)\n", progress, progress*job.Progress.Target/100, job.Progress.Target)
		}
		
		fmt.Printf("\n✅ Generation completed successfully!\n")
		fmt.Printf("📁 Output written to: %s\n", job.OutputPath)
	}
	
	return nil
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().Unix())
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
