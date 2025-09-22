package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/copyleftdev/faux-foundry/internal/spec"
)

var (
	template string
	force    bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [spec-file]",
	Short: "Initialize new specifications interactively",
	Long: `Initialize a new YAML specification file interactively. This command will guide you 
through creating a specification by asking questions about your data requirements.

If no filename is provided, it will create a specification based on the domain name.

Examples:
  # Create new specification interactively
  fauxfoundry init customer.yaml

  # Create from a template
  fauxfoundry init --template ecommerce customer.yaml

  # Force overwrite existing file
  fauxfoundry init --force customer.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&template, "template", "t", "", "template to use (ecommerce, user, product, etc.)")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing file")
}

func runInit(cmd *cobra.Command, args []string) error {
	var specPath string
	
	if len(args) > 0 {
		specPath = args[0]
	} else {
		specPath = "specification.yaml"
	}
	
	// Check if file exists and force flag
	if fileExists(specPath) && !force {
		return fmt.Errorf("file %s already exists (use --force to overwrite)", specPath)
	}
	
	// Ensure directory exists
	if dir := filepath.Dir(specPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	
	if !IsQuiet() {
		fmt.Printf("🚀 Creating new FauxFoundry specification: %s\n\n", specPath)
	}
	
	var specification *spec.Specification
	var err error
	
	if template != "" {
		// Create from template
		specification, err = createFromTemplate(template)
		if err != nil {
			return fmt.Errorf("failed to create from template: %w", err)
		}
		
		if !IsQuiet() {
			fmt.Printf("📋 Using template: %s\n", template)
		}
	} else {
		// Create interactively
		specification, err = createInteractively()
		if err != nil {
			return fmt.Errorf("failed to create specification: %w", err)
		}
	}
	
	// Save specification
	if err := spec.SaveToFile(specification, specPath); err != nil {
		return fmt.Errorf("failed to save specification: %w", err)
	}
	
	if !IsQuiet() {
		fmt.Printf("\n✅ Specification created successfully!\n")
		fmt.Printf("📁 File: %s\n", specPath)
		fmt.Printf("🎯 Domain: %s\n", specification.Dataset.Domain)
		fmt.Printf("📊 Fields: %d\n", len(specification.Dataset.Fields))
		fmt.Printf("🔢 Target records: %d\n", specification.Dataset.Count)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Review and customize the specification\n")
		fmt.Printf("  2. Validate: fauxfoundry validate %s\n", specPath)
		fmt.Printf("  3. Generate: fauxfoundry generate --spec %s\n", specPath)
	}
	
	return nil
}

func createFromTemplate(templateName string) (*spec.Specification, error) {
	templates := map[string]*spec.Specification{
		"ecommerce": {
			Model: spec.ModelConfig{
				Endpoint:    "http://localhost:11434",
				Name:        "llama3.1:8b",
				BatchSize:   32,
				Temperature: 0.7,
			},
			Dataset: spec.DatasetConfig{
				Count:  1000,
				Domain: "E-commerce customer data",
				Fields: []spec.Field{
					{Name: "email", Type: "email", Required: true, Pattern: "@(gmail|yahoo|outlook)\\.com$"},
					{Name: "age", Type: "integer", Required: true, Range: []int{18, 80}},
					{Name: "status", Type: "enum", Required: true, Values: []string{"active", "inactive", "pending"}},
					{Name: "created_at", Type: "datetime", Required: true, Description: "Account creation date"},
					{Name: "preferences", Type: "object", Description: "Customer preferences and settings"},
				},
			},
		},
		"user": {
			Model: spec.ModelConfig{
				Endpoint:    "http://localhost:11434",
				Name:        "llama3.1:8b",
				BatchSize:   32,
				Temperature: 0.7,
			},
			Dataset: spec.DatasetConfig{
				Count:  500,
				Domain: "User profile data",
				Fields: []spec.Field{
					{Name: "username", Type: "string", Required: true, Pattern: "^[a-zA-Z0-9_]{3,20}$"},
					{Name: "email", Type: "email", Required: true},
					{Name: "full_name", Type: "string", Required: true},
					{Name: "bio", Type: "text", Description: "User biography"},
					{Name: "avatar_url", Type: "url", Description: "Profile picture URL"},
				},
			},
		},
		"product": {
			Model: spec.ModelConfig{
				Endpoint:    "http://localhost:11434",
				Name:        "llama3.1:8b",
				BatchSize:   32,
				Temperature: 0.7,
			},
			Dataset: spec.DatasetConfig{
				Count:  2000,
				Domain: "Product catalog data",
				Fields: []spec.Field{
					{Name: "name", Type: "string", Required: true},
					{Name: "description", Type: "text", Required: true},
					{Name: "price", Type: "float", Required: true, Range: []int{1, 1000}},
					{Name: "category", Type: "enum", Required: true, Values: []string{"electronics", "clothing", "books", "home", "sports"}},
					{Name: "sku", Type: "string", Required: true, Pattern: "^[A-Z]{2}[0-9]{6}$"},
					{Name: "in_stock", Type: "boolean", Required: true},
				},
			},
		},
	}
	
	template, exists := templates[templateName]
	if !exists {
		available := make([]string, 0, len(templates))
		for name := range templates {
			available = append(available, name)
		}
		return nil, fmt.Errorf("unknown template '%s'. Available templates: %v", templateName, available)
	}
	
	return template, nil
}

func createInteractively() (*spec.Specification, error) {
	// TODO: Implement interactive specification creation
	// For now, return a basic template
	if !IsQuiet() {
		fmt.Printf("🔧 Interactive mode not yet fully implemented.\n")
		fmt.Printf("📋 Creating basic template - you can customize it manually.\n\n")
	}
	
	return &spec.Specification{
		Model: spec.ModelConfig{
			Endpoint:    "http://localhost:11434",
			Name:        "llama3.1:8b",
			BatchSize:   32,
			Temperature: 0.7,
		},
		Dataset: spec.DatasetConfig{
			Count:  1000,
			Domain: "Custom data domain",
			Fields: []spec.Field{
				{Name: "id", Type: "integer", Required: true, Description: "Unique identifier"},
				{Name: "name", Type: "string", Required: true, Description: "Name field"},
				{Name: "created_at", Type: "datetime", Required: true, Description: "Creation timestamp"},
			},
		},
	}, nil
}
