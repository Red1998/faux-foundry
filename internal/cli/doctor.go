package cli

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/copyleftdev/faux-foundry/internal/llm"
)

// doctorCmd represents the doctor command for system health checks
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and Ollama setup",
	Long: `Run comprehensive health checks to ensure FauxFoundry is properly configured.
This command checks:
  - Ollama installation and connectivity
  - Available models
  - System requirements
  - Configuration issues

Examples:
  # Run full health check
  fauxfoundry doctor

  # Check specific endpoint
  fauxfoundry doctor --endpoint http://localhost:11434`,
	RunE: runDoctor,
}

var (
	doctorEndpoint string
)

func init() {
	doctorCmd.Flags().StringVar(&doctorEndpoint, "endpoint", "http://localhost:11434", "Ollama endpoint to check")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	if !IsQuiet() {
		fmt.Printf("🏥 FauxFoundry Health Check\n")
		fmt.Printf("═══════════════════════════\n\n")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check system information
	if !IsQuiet() {
		fmt.Printf("📋 System Information:\n")
		fmt.Printf("   • OS: %s\n", runtime.GOOS)
		fmt.Printf("   • Architecture: %s\n", runtime.GOARCH)
		fmt.Printf("   • Go version: %s\n", runtime.Version())
		fmt.Printf("   • CPU cores: %d\n", runtime.NumCPU())
		fmt.Println()
	}

	// Check Ollama health
	client := llm.NewOllamaClient()
	health, err := client.CheckOllamaHealth(ctx, doctorEndpoint)
	if err != nil {
		return fmt.Errorf("failed to check Ollama health: %w", err)
	}

	if !IsQuiet() {
		fmt.Printf("🤖 Ollama Status:\n")
		fmt.Printf("   • Endpoint: %s\n", health.Endpoint)
		
		if health.IsRunning {
			fmt.Printf("   • Status: ✅ Running\n")
			fmt.Printf("   • Version: %s\n", health.Version)
			fmt.Printf("   • Available models: %d\n", len(health.Models))
			
			if len(health.Models) > 0 {
				fmt.Printf("   • Models:\n")
				for _, model := range health.Models {
					fmt.Printf("     - %s\n", model)
				}
			}
		} else {
			fmt.Printf("   • Status: ❌ Not running\n")
			if health.ErrorMessage != "" {
				fmt.Printf("   • Error: %s\n", health.ErrorMessage)
			}
		}
		fmt.Println()
	}

	// Provide recommendations
	if !health.IsRunning {
		return showOllamaSetupInstructions()
	}

	if len(health.Models) == 0 {
		return showModelInstallInstructions()
	}

	// Check if recommended models are available
	recommendedModels := llm.GetRecommendedModels()
	hasRecommended := false
	
	for _, recommended := range recommendedModels {
		if recommended.Recommended {
			for _, available := range health.Models {
				if available == recommended.Name {
					hasRecommended = true
					break
				}
			}
		}
	}

	if !hasRecommended {
		if !IsQuiet() {
			fmt.Printf("💡 Recommendations:\n")
			fmt.Printf("   Consider installing a recommended model for better performance:\n")
			for _, model := range recommendedModels {
				if model.Recommended {
					fmt.Printf("   • %s (%s) - %s\n", model.Name, model.Size, model.UseCase)
					fmt.Printf("     Install: ollama pull %s\n", model.Name)
				}
			}
			fmt.Println()
		}
	}

	if !IsQuiet() {
		fmt.Printf("✅ System is ready for synthetic data generation!\n")
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Create a specification: fauxfoundry init my-spec.yaml\n")
		fmt.Printf("  2. Generate data: fauxfoundry generate --spec my-spec.yaml\n")
		fmt.Printf("  3. Or use the TUI: fauxfoundry tui\n")
	}

	return nil
}

func showOllamaSetupInstructions() error {
	if IsQuiet() {
		return fmt.Errorf("Ollama is not running")
	}

	fmt.Printf("🚨 Ollama Setup Required\n")
	fmt.Printf("═══════════════════════════\n\n")
	fmt.Printf("Ollama is not running or not installed. FauxFoundry requires Ollama to generate synthetic data.\n\n")

	instructions := llm.GetOllamaInstallInstructions()
	
	fmt.Printf("📦 Installation Instructions for %s:\n\n", runtime.GOOS)
	
	var osInstructions string
	switch runtime.GOOS {
	case "darwin":
		osInstructions = instructions["macos"]
	case "linux":
		osInstructions = instructions["linux"]
	case "windows":
		osInstructions = instructions["windows"]
	default:
		osInstructions = instructions["linux"] // fallback
	}

	fmt.Printf("%s\n\n", osInstructions)
	
	fmt.Printf("🐳 Alternative - Docker:\n\n")
	fmt.Printf("%s\n\n", instructions["docker"])
	
	fmt.Printf("After installation:\n")
	fmt.Printf("  1. Ensure Ollama is running: ollama serve\n")
	fmt.Printf("  2. Run health check again: fauxfoundry doctor\n")
	fmt.Printf("  3. Visit https://ollama.ai for more information\n")

	return fmt.Errorf("Ollama setup required")
}

func showModelInstallInstructions() error {
	if IsQuiet() {
		return fmt.Errorf("no models available")
	}

	fmt.Printf("📥 Model Installation Required\n")
	fmt.Printf("═════════════════════════════\n\n")
	fmt.Printf("Ollama is running but no models are installed. You need at least one model to generate data.\n\n")

	recommendedModels := llm.GetRecommendedModels()
	
	fmt.Printf("🌟 Recommended Models:\n\n")
	for _, model := range recommendedModels {
		if model.Recommended {
			fmt.Printf("• %s (%s)\n", model.Name, model.Size)
			fmt.Printf("  %s\n", model.Description)
			fmt.Printf("  Use case: %s\n", model.UseCase)
			fmt.Printf("  Install: ollama pull %s\n\n", model.Name)
		}
	}

	fmt.Printf("💡 Quick start:\n")
	fmt.Printf("  ollama pull llama3.1:8b  # Recommended for most users\n\n")
	
	fmt.Printf("After installing a model:\n")
	fmt.Printf("  1. Run health check: fauxfoundry doctor\n")
	fmt.Printf("  2. Create your first spec: fauxfoundry init test.yaml\n")

	return fmt.Errorf("model installation required")
}
