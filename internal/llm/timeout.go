package llm

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/copyleftdev/faux-foundry/pkg/types"
)

// TimeoutStrategy defines different approaches to handle timeouts
type TimeoutStrategy int

const (
	StrategyRetry TimeoutStrategy = iota
	StrategyReduce
	StrategySimplify
	StrategyFallback
)

// RetryConfig defines retry behavior for timeout handling
type RetryConfig struct {
	MaxRetries       int           // Maximum number of retries
	BaseTimeout      time.Duration // Initial timeout duration
	MaxTimeout       time.Duration // Maximum timeout duration
	BackoffMultiplier float64      // Exponential backoff multiplier
	ReduceFactorOnTimeout float64  // Factor to reduce batch size on timeout
	MinBatchSize     int           // Minimum batch size before giving up
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:            3,
		BaseTimeout:           30 * time.Second,
		MaxTimeout:            5 * time.Minute,
		BackoffMultiplier:     2.0,
		ReduceFactorOnTimeout: 0.5,
		MinBatchSize:          1,
	}
}

// TimeoutHandler manages timeout detection and recovery strategies
type TimeoutHandler struct {
	config *RetryConfig
	client *OllamaClient
}

// NewTimeoutHandler creates a new timeout handler
func NewTimeoutHandler(client *OllamaClient, config *RetryConfig) *TimeoutHandler {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &TimeoutHandler{
		config: config,
		client: client,
	}
}

// GenerateWithRetry generates data with intelligent timeout handling
func (th *TimeoutHandler) GenerateWithRetry(ctx context.Context, spec *types.Specification, count int) ([]types.Record, error) {
	var allRecords []types.Record
	remaining := count
	currentBatchSize := spec.Model.BatchSize
	attempt := 0
	currentTimeout := th.config.BaseTimeout

	fmt.Printf("🔄 Starting generation with timeout handling (target: %d records)\n", count)

	for remaining > 0 && attempt < th.config.MaxRetries {
		batchSize := min(currentBatchSize, remaining)
		attempt++

		fmt.Printf("📦 Attempt %d: Generating %d records (timeout: %s)\n", 
			attempt, batchSize, currentTimeout)

		// Create context with current timeout
		batchCtx, cancel := context.WithTimeout(ctx, currentTimeout)
		
		// Try to generate the batch
		records, err := th.generateBatch(batchCtx, spec, batchSize)
		cancel()

		if err != nil {
			strategy := th.determineStrategy(err, attempt, currentBatchSize)
			fmt.Printf("⚠️  Batch failed: %v\n", err)
			fmt.Printf("🔧 Applying strategy: %s\n", th.strategyName(strategy))

			switch strategy {
			case StrategyRetry:
				// Increase timeout and retry same batch size
				currentTimeout = th.increaseTimeout(currentTimeout)
				fmt.Printf("⏱️  Increased timeout to %s, retrying...\n", currentTimeout)
				continue

			case StrategyReduce:
				// Reduce batch size and retry
				currentBatchSize = max(1, int(float64(currentBatchSize)*th.config.ReduceFactorOnTimeout))
				currentTimeout = th.config.BaseTimeout // Reset timeout
				fmt.Printf("📉 Reduced batch size to %d, resetting timeout\n", currentBatchSize)
				continue

			case StrategySimplify:
				// Simplify the specification and retry
				simplifiedSpec := th.simplifySpec(spec)
				records, err = th.generateBatch(batchCtx, simplifiedSpec, batchSize)
				if err != nil {
					fmt.Printf("❌ Simplified generation also failed: %v\n", err)
					continue
				}

			case StrategyFallback:
				// Use fallback generation method
				fmt.Printf("🆘 Using fallback generation method\n")
				records = th.generateFallbackData(spec, batchSize)
			}
		}

		// Successfully generated records
		if len(records) > 0 {
			allRecords = append(allRecords, records...)
			remaining -= len(records)
			fmt.Printf("✅ Generated %d records (%d remaining)\n", len(records), remaining)
			
			// Reset for next batch
			attempt = 0
			currentTimeout = th.config.BaseTimeout
		}
	}

	if len(allRecords) == 0 {
		return nil, fmt.Errorf("failed to generate any records after %d attempts", th.config.MaxRetries)
	}

	if len(allRecords) < count {
		fmt.Printf("⚠️  Generated %d/%d records (partial success)\n", len(allRecords), count)
	}

	return allRecords, nil
}

// generateBatch generates a single batch with the given context
func (th *TimeoutHandler) generateBatch(ctx context.Context, spec *types.Specification, count int) ([]types.Record, error) {
	// Use the basic generate method to avoid recursive timeout handling
	return th.client.GenerateBasic(ctx, spec, count)
}

// determineStrategy decides which strategy to use based on the error and context
func (th *TimeoutHandler) determineStrategy(err error, attempt int, currentBatchSize int) TimeoutStrategy {
	errStr := err.Error()

	// Check if it's a timeout error
	if isTimeoutError(err) {
		if attempt == 1 {
			return StrategyRetry // First timeout: just increase timeout
		} else if currentBatchSize > th.config.MinBatchSize {
			return StrategyReduce // Subsequent timeouts: reduce batch size
		} else {
			return StrategySimplify // Small batch still timing out: simplify spec
		}
	}

	// Check if it's a parsing error
	if isParsingError(errStr) {
		if attempt <= 2 {
			return StrategyRetry // Retry parsing errors a couple times
		} else {
			return StrategySimplify // Persistent parsing issues: simplify
		}
	}

	// Check if it's a connection error
	if isConnectionError(errStr) {
		return StrategyRetry // Always retry connection errors
	}

	// Default to retry for unknown errors
	return StrategyRetry
}

// isTimeoutError checks if the error is related to timeouts
func isTimeoutError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "timeout") || 
		   contains(errStr, "deadline exceeded") ||
		   contains(errStr, "context canceled")
}

// isParsingError checks if the error is related to parsing LLM responses
func isParsingError(errStr string) bool {
	return contains(errStr, "parse") ||
		   contains(errStr, "json") ||
		   contains(errStr, "unmarshal") ||
		   contains(errStr, "invalid character")
}

// isConnectionError checks if the error is related to network connectivity
func isConnectionError(errStr string) bool {
	return contains(errStr, "connection") ||
		   contains(errStr, "network") ||
		   contains(errStr, "refused") ||
		   contains(errStr, "unreachable")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		   findInString(s, substr)))
}

// findInString is a simple substring search
func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// increaseTimeout increases the timeout using exponential backoff
func (th *TimeoutHandler) increaseTimeout(current time.Duration) time.Duration {
	next := time.Duration(float64(current) * th.config.BackoffMultiplier)
	if next > th.config.MaxTimeout {
		return th.config.MaxTimeout
	}
	return next
}

// simplifySpec creates a simplified version of the specification for fallback
func (th *TimeoutHandler) simplifySpec(spec *types.Specification) *types.Specification {
	simplified := *spec // Copy the spec
	
	// Reduce temperature for more predictable output
	simplified.Model.Temperature = math.Min(simplified.Model.Temperature, 0.3)
	
	// Keep only required fields to reduce complexity
	var requiredFields []types.Field
	for _, field := range spec.Dataset.Fields {
		if field.Required {
			requiredFields = append(requiredFields, field)
		}
	}
	
	// If no required fields, keep first 3 fields
	if len(requiredFields) == 0 {
		maxFields := min(3, len(spec.Dataset.Fields))
		requiredFields = spec.Dataset.Fields[:maxFields]
	}
	
	simplified.Dataset.Fields = requiredFields
	simplified.Dataset.Domain = "Simplified " + spec.Dataset.Domain
	
	return &simplified
}

// generateFallbackData generates basic fallback data when all else fails
func (th *TimeoutHandler) generateFallbackData(spec *types.Specification, count int) []types.Record {
	fmt.Printf("🔧 Generating fallback data (basic patterns)\n")
	
	records := make([]types.Record, count)
	for i := 0; i < count; i++ {
		record := make(types.Record)
		
		for _, field := range spec.Dataset.Fields {
			switch field.Type {
			case "string":
				if field.Pattern != "" {
					record[field.Name] = th.client.generatePatternString(field.Pattern, i)
				} else {
					record[field.Name] = fmt.Sprintf("fallback_%s_%d", field.Name, i)
				}
			case "email":
				record[field.Name] = fmt.Sprintf("user%d@example.com", i)
			case "integer":
				if len(field.Range) == 2 {
					record[field.Name] = field.Range[0] + (i % (field.Range[1] - field.Range[0]))
				} else {
					record[field.Name] = i + 1
				}
			case "float":
				if len(field.Range) == 2 {
					record[field.Name] = float64(field.Range[0]) + float64(i%100)/100.0*float64(field.Range[1]-field.Range[0])
				} else {
					record[field.Name] = float64(i) + 0.5
				}
			case "boolean":
				record[field.Name] = i%2 == 0
			case "enum":
				if len(field.Values) > 0 {
					record[field.Name] = field.Values[i%len(field.Values)]
				}
			case "date":
				record[field.Name] = time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			case "datetime":
				record[field.Name] = time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339)
			default:
				record[field.Name] = fmt.Sprintf("fallback_%s_%d", field.Type, i)
			}
		}
		
		records[i] = record
	}
	
	return records
}

// strategyName returns a human-readable name for the strategy
func (th *TimeoutHandler) strategyName(strategy TimeoutStrategy) string {
	switch strategy {
	case StrategyRetry:
		return "Retry with increased timeout"
	case StrategyReduce:
		return "Reduce batch size"
	case StrategySimplify:
		return "Simplify specification"
	case StrategyFallback:
		return "Use fallback generation"
	default:
		return "Unknown strategy"
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
