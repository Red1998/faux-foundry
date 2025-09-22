package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/copyleftdev/faux-foundry/internal/spec"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [spec-file]",
	Short: "Validate YAML specifications",
	Long: `Validate one or more YAML specification files for syntax and semantic correctness.
This command checks that the specification is well-formed, all required fields are present,
and field constraints are valid.

Examples:
  # Validate a single specification
  fauxfoundry validate customer.yaml

  # Validate multiple specifications
  fauxfoundry validate customer.yaml products.yaml orders.yaml

  # Validate with dry-run (same as validate)
  fauxfoundry validate --dry-run customer.yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "same as validate (included for consistency)")
}

func runValidate(cmd *cobra.Command, args []string) error {
	var hasErrors bool
	
	for _, specPath := range args {
		if err := validateSingleSpec(specPath); err != nil {
			hasErrors = true
			if !IsQuiet() {
				fmt.Fprintf(os.Stderr, "❌ %s: %v\n", specPath, err)
			}
		} else {
			if !IsQuiet() {
				fmt.Printf("✅ %s: valid\n", specPath)
			}
		}
	}
	
	if hasErrors {
		return fmt.Errorf("validation failed for one or more specifications")
	}
	
	if !IsQuiet() {
		fmt.Printf("\n🎉 All specifications are valid!\n")
	}
	
	return nil
}

func validateSingleSpec(specPath string) error {
	// Check if file exists
	if !fileExists(specPath) {
		return fmt.Errorf("file not found")
	}
	
	// Load specification
	specification, err := spec.LoadFromFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}
	
	// Validate specification
	if err := spec.Validate(specification); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// If verbose, show details
	if IsVerbose() {
		fmt.Printf("  Domain: %s\n", specification.Dataset.Domain)
		fmt.Printf("  Fields: %d\n", len(specification.Dataset.Fields))
		fmt.Printf("  Target records: %d\n", specification.Dataset.Count)
		fmt.Printf("  Model: %s\n", specification.Model.Name)
		fmt.Printf("  Endpoint: %s\n", specification.Model.Endpoint)
		fmt.Printf("  Batch size: %d\n", specification.Model.BatchSize)
		fmt.Printf("  Temperature: %.2f\n", specification.Model.Temperature)
		
		if len(specification.Dataset.Fields) > 0 {
			fmt.Printf("  Field details:\n")
			for _, field := range specification.Dataset.Fields {
				fmt.Printf("    - %s (%s)", field.Name, field.Type)
				if field.Required {
					fmt.Printf(" [required]")
				}
				if field.Pattern != "" {
					fmt.Printf(" pattern: %s", field.Pattern)
				}
				if len(field.Range) == 2 {
					fmt.Printf(" range: [%d, %d]", field.Range[0], field.Range[1])
				}
				if len(field.Values) > 0 {
					fmt.Printf(" values: %v", field.Values)
				}
				fmt.Println()
			}
		}
	}
	
	return nil
}
