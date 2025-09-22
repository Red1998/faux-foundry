package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/copyleftdev/faux-foundry/pkg/types"
)

// Client represents an LLM client interface
type Client interface {
	Generate(ctx context.Context, spec *types.Specification, count int) ([]types.Record, error)
	TestConnection(ctx context.Context, endpoint string) error
	ListModels(ctx context.Context, endpoint string) ([]string, error)
}

// OllamaClient implements the Client interface for Ollama
type OllamaClient struct {
	httpClient *http.Client
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Error     string `json:"error,omitempty"`
}

// Generate generates synthetic data using the LLM with timeout handling
func (c *OllamaClient) Generate(ctx context.Context, spec *types.Specification, count int) ([]types.Record, error) {
	// Use timeout handler for robust generation
	handler := NewTimeoutHandler(c, DefaultRetryConfig())
	return handler.GenerateWithRetry(ctx, spec, count)
}

// GenerateWithConfig generates data with custom retry configuration
func (c *OllamaClient) GenerateWithConfig(ctx context.Context, spec *types.Specification, count int, config *RetryConfig) ([]types.Record, error) {
	// Use timeout handler with custom config
	handler := NewTimeoutHandler(c, config)
	return handler.GenerateWithRetry(ctx, spec, count)
}

// GenerateBasic generates data without timeout handling (for internal use)
func (c *OllamaClient) GenerateBasic(ctx context.Context, spec *types.Specification, count int) ([]types.Record, error) {
	// Build prompt from specification
	prompt := c.buildPrompt(spec, count)
	
	// Make request to Ollama
	req := OllamaRequest{
		Model:  spec.Model.Name,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": spec.Model.Temperature,
		},
	}
	
	response, err := c.makeRequest(ctx, spec.Model.Endpoint+"/api/generate", req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate data: %w", err)
	}
	
	// Debug: Show that we got a real LLM response
	fmt.Printf("🤖 Raw LLM Response (%d chars): %s...\n", len(response.Response), response.Response[:min(100, len(response.Response))])
	
	// Parse response into records
	records, err := c.parseResponse(response.Response, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return records, nil
}

// generateDemoData generates realistic demo data without requiring Ollama
func (c *OllamaClient) generateDemoData(spec *types.Specification, count int) []types.Record {
	records := make([]types.Record, 0, count)
	
	// Sample realistic data for medical/healthcare domain
	firstNames := []string{"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Emily", "James", "Ashley"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}
	
	for i := 0; i < count; i++ {
		record := make(types.Record)
		
		for _, field := range spec.Dataset.Fields {
			switch field.Type {
			case "string":
				if field.Pattern != "" {
					record[field.Name] = c.generatePatternString(field.Pattern, i)
				} else if field.Name == "first_name" {
					record[field.Name] = firstNames[i%len(firstNames)]
				} else if field.Name == "last_name" {
					record[field.Name] = lastNames[i%len(lastNames)]
				} else {
					record[field.Name] = fmt.Sprintf("sample_%s_%d", field.Name, i)
				}
			case "email":
				record[field.Name] = fmt.Sprintf("%s.%s@example.com", 
					firstNames[i%len(firstNames)], lastNames[i%len(lastNames)])
			case "integer":
				if len(field.Range) == 2 {
					record[field.Name] = field.Range[0] + (i % (field.Range[1] - field.Range[0]))
				} else {
					record[field.Name] = i + 1
				}
			case "float":
				if len(field.Range) == 2 {
					baseValue := float64(field.Range[0])
					rangeSize := float64(field.Range[1] - field.Range[0])
					record[field.Name] = baseValue + (float64(i%100)/100.0)*rangeSize
				} else {
					record[field.Name] = float64(i) * 10.5
				}
			case "boolean":
				record[field.Name] = i%2 == 0
			case "enum":
				if len(field.Values) > 0 {
					record[field.Name] = field.Values[i%len(field.Values)]
				}
			case "date":
				// Generate dates in the past 30 years
				baseDate := time.Now().AddDate(-30, 0, 0)
				daysToAdd := i * 365 // Spread across years
				record[field.Name] = baseDate.AddDate(0, 0, daysToAdd).Format("2006-01-02")
			case "datetime":
				record[field.Name] = time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339)
			case "array":
				record[field.Name] = []string{fmt.Sprintf("item_%d_1", i), fmt.Sprintf("item_%d_2", i)}
			default:
				record[field.Name] = fmt.Sprintf("sample_%s_%d", field.Type, i)
			}
		}
		
		records = append(records, record)
	}
	
	return records
}

// generatePatternString generates a string that matches a regex pattern
func (c *OllamaClient) generatePatternString(pattern string, seed int) string {
	// Simple pattern matching for common medical patterns
	switch pattern {
	case "^PAT[0-9]{8}$":
		return fmt.Sprintf("PAT%08d", 10000000+seed)
	case "^[A-Z]{3}[0-9]{9}$":
		return fmt.Sprintf("ABC%09d", 100000000+seed)
	case "^[0-9]{3}-[0-9]{2}-[0-9]{4}$":
		return fmt.Sprintf("%03d-%02d-%04d", 100+seed%900, seed%100, 1000+seed%9000)
	case "^[0-9]{5}(-[0-9]{4})?$":
		if seed%2 == 0 {
			return fmt.Sprintf("%05d", 10000+seed%90000)
		}
		return fmt.Sprintf("%05d-%04d", 10000+seed%90000, 1000+seed%9000)
	case "^\\([0-9]{3}\\) [0-9]{3}-[0-9]{4}$":
		return fmt.Sprintf("(%03d) %03d-%04d", 200+seed%800, 100+seed%900, 1000+seed%9000)
	case "^GRP[0-9]{6}[A-Z]{2}$":
		letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		return fmt.Sprintf("GRP%06d%c%c", 100000+seed%900000, 
			letters[seed%26], letters[(seed+1)%26])
	case "^POL[0-9]{10}$":
		return fmt.Sprintf("POL%010d", 1000000000+seed%9000000000)
	case "^EDI[0-9]{12}$":
		return fmt.Sprintf("EDI%012d", seed)
	case "^[0-9]{5}$":
		return fmt.Sprintf("%05d", 10000+seed%90000)
	case "^[0-9]{10}$":
		return fmt.Sprintf("%010d", 1000000000+seed%9000000000)
	case "^[0-9]{9}$":
		return fmt.Sprintf("%09d", 100000000+seed%900000000)
	case "^[0-9]{2}-[0-9]{7}$":
		return fmt.Sprintf("%02d-%07d", 10+seed%90, 1000000+seed%9000000)
	case "^CH[0-9]{6}$":
		return fmt.Sprintf("CH%06d", 100000+seed%900000)
	case "^BTH[0-9]{8}$":
		return fmt.Sprintf("BTH%08d", 10000000+seed%90000000)
	default:
		return fmt.Sprintf("pattern_match_%d", seed)
	}
}

// TestConnection tests the connection to an Ollama endpoint
func (c *OllamaClient) TestConnection(ctx context.Context, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return nil
}

// ListModels lists available models from Ollama
func (c *OllamaClient) ListModels(ctx context.Context, endpoint string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}
	
	return models, nil
}

// buildPrompt builds a prompt from the specification
func (c *OllamaClient) buildPrompt(spec *types.Specification, count int) string {
	prompt := fmt.Sprintf(`Generate %d unique JSON records for %s.

Each record should be a valid JSON object with the following fields:
`, count, spec.Dataset.Domain)

	for _, field := range spec.Dataset.Fields {
		prompt += fmt.Sprintf("- %s (%s)", field.Name, field.Type)
		if field.Required {
			prompt += " [required]"
		}
		if field.Description != "" {
			prompt += fmt.Sprintf(": %s", field.Description)
		}
		if field.Pattern != "" {
			prompt += fmt.Sprintf(" (pattern: %s)", field.Pattern)
		}
		if len(field.Range) == 2 {
			prompt += fmt.Sprintf(" (range: %d-%d)", field.Range[0], field.Range[1])
		}
		if len(field.Values) > 0 {
			prompt += fmt.Sprintf(" (values: %v)", field.Values)
		}
		prompt += "\n"
	}

	prompt += `
Requirements:
- Each record must be unique
- Output only valid JSON objects, one per line
- Follow the field constraints exactly
- Make the data realistic and diverse
- Do not include any explanatory text

Generate the records now:`

	return prompt
}

// makeRequest makes an HTTP request to Ollama
func (c *OllamaClient) makeRequest(ctx context.Context, url string, req OllamaRequest) (*OllamaResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
	
	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if response.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", response.Error)
	}
	
	return &response, nil
}

// parseResponse parses the LLM response into records
func (c *OllamaClient) parseResponse(response string, spec *types.Specification) ([]types.Record, error) {
	records := make([]types.Record, 0)
	
	// Remove markdown formatting
	response = strings.ReplaceAll(response, "```json", "")
	response = strings.ReplaceAll(response, "```", "")
	
	// Split into potential JSON objects by looking for }{ patterns and newlines
	response = strings.ReplaceAll(response, "}\n{", "}\n\n{")
	response = strings.ReplaceAll(response, "}{", "}\n\n{")
	
	// Split by double newlines to separate JSON objects
	parts := strings.Split(response, "\n\n")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		
		// Skip empty parts or explanatory text
		if part == "" || 
		   strings.HasPrefix(part, "Here") || strings.HasPrefix(part, "I'll") ||
		   strings.HasPrefix(part, "The") || strings.HasPrefix(part, "Based") ||
		   strings.HasPrefix(part, "Note") || strings.HasPrefix(part, "This") {
			continue
		}
		
		// Try to parse as JSON object
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			var record types.Record
			if err := json.Unmarshal([]byte(part), &record); err != nil {
				// Try to fix common JSON issues
				fixedPart := c.fixCommonJSONIssues(part)
				if err := json.Unmarshal([]byte(fixedPart), &record); err != nil {
					continue // Skip this record if we can't parse it
				}
			}
			
			// Validate that the record has the expected fields
			if c.validateRecord(record, spec) {
				records = append(records, record)
			}
		}
	}
	
	// If we couldn't parse any records from the LLM response, return an error
	if len(records) == 0 {
		fmt.Printf("❌ Failed to parse any records from LLM response\n")
		fmt.Printf("LLM Response (first 500 chars): %s\n", response[:min(500, len(response))])
		fmt.Printf("Response length: %d characters\n", len(response))
		return nil, fmt.Errorf("could not parse any valid JSON records from LLM response")
	}
	
	fmt.Printf("✅ Successfully parsed %d records from LLM response\n", len(records))
	return records, nil
}

// fixCommonJSONIssues attempts to fix common JSON formatting issues
func (c *OllamaClient) fixCommonJSONIssues(jsonStr string) string {
	// Remove trailing commas before closing braces
	jsonStr = strings.ReplaceAll(jsonStr, ",\n}", "\n}")
	jsonStr = strings.ReplaceAll(jsonStr, ", }", " }")
	
	// Fix null values that might be unquoted
	jsonStr = strings.ReplaceAll(jsonStr, ": null", ": null")
	
	return jsonStr
}

// validateRecord checks if a record contains the required fields from the spec
func (c *OllamaClient) validateRecord(record types.Record, spec *types.Specification) bool {
	requiredFields := 0
	presentFields := 0
	
	for _, field := range spec.Dataset.Fields {
		if field.Required {
			requiredFields++
			if _, exists := record[field.Name]; exists {
				presentFields++
			}
		}
	}
	
	// Require at least 80% of required fields to be present
	if requiredFields > 0 {
		return float64(presentFields)/float64(requiredFields) >= 0.8
	}
	
	// If no required fields, just check that we have some fields
	return len(record) > 0
}

// IsVerbose checks if verbose mode is enabled (placeholder - would be injected)
func IsVerbose() bool {
	// TODO: This should be injected from CLI context
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
