package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OllamaHealth represents the health status of Ollama
type OllamaHealth struct {
	IsRunning     bool     `json:"is_running"`
	Version       string   `json:"version,omitempty"`
	Models        []string `json:"models,omitempty"`
	Endpoint      string   `json:"endpoint"`
	LastChecked   time.Time `json:"last_checked"`
	ErrorMessage  string   `json:"error_message,omitempty"`
}

// OllamaModel represents an available Ollama model
type OllamaModel struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModifiedAt   time.Time `json:"modified_at"`
	Digest       string    `json:"digest"`
	Details      ModelDetails `json:"details"`
}

// ModelDetails contains detailed information about a model
type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// OllamaTagsResponse represents the response from /api/tags
type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

// OllamaVersionResponse represents the response from /api/version
type OllamaVersionResponse struct {
	Version string `json:"version"`
}

// CheckOllamaHealth performs a comprehensive health check of Ollama
func (c *OllamaClient) CheckOllamaHealth(ctx context.Context, endpoint string) (*OllamaHealth, error) {
	health := &OllamaHealth{
		Endpoint:    endpoint,
		LastChecked: time.Now(),
		IsRunning:   false,
	}

	// Test basic connectivity
	if err := c.testConnectivity(ctx, endpoint); err != nil {
		health.ErrorMessage = fmt.Sprintf("Connection failed: %v", err)
		return health, nil // Return health status, not error
	}

	// Get version information
	version, err := c.getOllamaVersion(ctx, endpoint)
	if err != nil {
		health.ErrorMessage = fmt.Sprintf("Failed to get version: %v", err)
		return health, nil
	}
	health.Version = version

	// Get available models
	models, err := c.getOllamaModels(ctx, endpoint)
	if err != nil {
		health.ErrorMessage = fmt.Sprintf("Failed to get models: %v", err)
		return health, nil
	}

	// Extract model names
	modelNames := make([]string, len(models))
	for i, model := range models {
		modelNames[i] = model.Name
	}
	health.Models = modelNames
	health.IsRunning = true

	return health, nil
}

// testConnectivity tests basic connectivity to Ollama
func (c *OllamaClient) testConnectivity(ctx context.Context, endpoint string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// getOllamaVersion gets the Ollama version
func (c *OllamaClient) getOllamaVersion(ctx context.Context, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/api/version", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var versionResp OllamaVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&versionResp); err != nil {
		return "", fmt.Errorf("failed to decode version response: %w", err)
	}

	return versionResp.Version, nil
}

// getOllamaModels gets the list of available models
func (c *OllamaClient) getOllamaModels(ctx context.Context, endpoint string) ([]OllamaModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tagsResp OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode tags response: %w", err)
	}

	return tagsResp.Models, nil
}

// PullModel pulls a model from Ollama registry
func (c *OllamaClient) PullModel(ctx context.Context, endpoint, modelName string) error {
	pullReq := map[string]interface{}{
		"name": modelName,
	}

	jsonData, err := json.Marshal(pullReq)
	if err != nil {
		return fmt.Errorf("failed to marshal pull request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint+"/api/pull", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pull failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// GetRecommendedModels returns a list of recommended models for synthetic data generation
func GetRecommendedModels() []RecommendedModel {
	return []RecommendedModel{
		{
			Name:        "llama3.1:8b",
			Description: "Meta's Llama 3.1 8B - Excellent for structured data generation",
			Size:        "4.7GB",
			Recommended: true,
			UseCase:     "General purpose synthetic data generation",
		},
		{
			Name:        "llama3.1:70b",
			Description: "Meta's Llama 3.1 70B - Highest quality but requires more resources",
			Size:        "40GB",
			Recommended: false,
			UseCase:     "High-quality complex data generation (requires powerful hardware)",
		},
		{
			Name:        "mistral:7b",
			Description: "Mistral 7B - Fast and efficient for structured data",
			Size:        "4.1GB",
			Recommended: true,
			UseCase:     "Fast synthetic data generation with good quality",
		},
		{
			Name:        "codellama:13b",
			Description: "Code Llama 13B - Good for technical/structured data",
			Size:        "7.3GB",
			Recommended: false,
			UseCase:     "Technical data, API responses, structured formats",
		},
		{
			Name:        "phi3:mini",
			Description: "Microsoft Phi-3 Mini - Lightweight and fast",
			Size:        "2.3GB",
			Recommended: true,
			UseCase:     "Quick testing and lightweight data generation",
		},
	}
}

// RecommendedModel represents a recommended model for FauxFoundry
type RecommendedModel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Size        string `json:"size"`
	Recommended bool   `json:"recommended"`
	UseCase     string `json:"use_case"`
}

// GetOllamaInstallInstructions returns platform-specific installation instructions
func GetOllamaInstallInstructions() map[string]string {
	return map[string]string{
		"macos": `# Install Ollama on macOS
curl -fsSL https://ollama.ai/install.sh | sh

# Or using Homebrew
brew install ollama

# Start Ollama service
ollama serve

# Pull a recommended model
ollama pull llama3.1:8b`,

		"linux": `# Install Ollama on Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama service
ollama serve

# Pull a recommended model
ollama pull llama3.1:8b`,

		"windows": `# Install Ollama on Windows
# Download from: https://ollama.ai/download/windows

# After installation, open PowerShell/CMD and run:
ollama serve

# Pull a recommended model
ollama pull llama3.1:8b`,

		"docker": `# Run Ollama with Docker
docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama

# Pull a model
docker exec -it ollama ollama pull llama3.1:8b`,
	}
}
