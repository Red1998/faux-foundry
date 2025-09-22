package spec

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
	"github.com/copyleftdev/faux-foundry/pkg/types"
)

// Type aliases for cleaner code
type (
	Specification = types.Specification
	ModelConfig   = types.ModelConfig
	DatasetConfig = types.DatasetConfig
	Field         = types.Field
)

// LoadFromFile loads and parses a specification from a YAML file
func LoadFromFile(filename string) (*Specification, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	return ParseYAML(data)
}

// ParseYAML parses a specification from YAML bytes
func ParseYAML(data []byte) (*Specification, error) {
	var spec Specification
	
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Set defaults
	setDefaults(&spec)
	
	return &spec, nil
}

// SaveToFile saves a specification to a YAML file
func SaveToFile(spec *Specification, filename string) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// ToYAML converts a specification to YAML string
func ToYAML(spec *Specification) (string, error) {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	
	return string(data), nil
}

// setDefaults sets default values for missing fields
func setDefaults(spec *Specification) {
	// Model defaults
	if spec.Model.Endpoint == "" {
		spec.Model.Endpoint = "http://localhost:11434"
	}
	if spec.Model.Name == "" {
		spec.Model.Name = "llama3.1:8b"
	}
	if spec.Model.BatchSize == 0 {
		spec.Model.BatchSize = 32
	}
	if spec.Model.Temperature == 0 {
		spec.Model.Temperature = 0.7
	}
	if spec.Model.Timeout == "" {
		spec.Model.Timeout = "30s"
	}
	
	// Dataset defaults
	if spec.Dataset.Count == 0 {
		spec.Dataset.Count = 1000
	}
	if spec.Dataset.Domain == "" {
		spec.Dataset.Domain = "Generic data"
	}
	
	// Field defaults
	for i := range spec.Dataset.Fields {
		field := &spec.Dataset.Fields[i]
		if field.Type == "" {
			field.Type = "string"
		}
		// Set required to true by default for primary fields
		if field.Name == "id" || field.Name == "email" || field.Name == "name" {
			field.Required = true
		}
	}
}

// Validate validates a specification for correctness
func Validate(spec *Specification) error {
	if err := validateModel(&spec.Model); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}
	
	if err := validateDataset(&spec.Dataset); err != nil {
		return fmt.Errorf("dataset validation failed: %w", err)
	}
	
	return nil
}

// validateModel validates the model configuration
func validateModel(model *ModelConfig) error {
	if model.Endpoint == "" {
		return fmt.Errorf("model endpoint is required")
	}
	
	if model.Name == "" {
		return fmt.Errorf("model name is required")
	}
	
	if model.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive, got %d", model.BatchSize)
	}
	
	if model.BatchSize > 1000 {
		return fmt.Errorf("batch size too large (max 1000), got %d", model.BatchSize)
	}
	
	if model.Temperature < 0 || model.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2, got %.2f", model.Temperature)
	}
	
	// Validate endpoint format (basic check)
	if !strings.HasPrefix(model.Endpoint, "http://") && !strings.HasPrefix(model.Endpoint, "https://") {
		return fmt.Errorf("endpoint must be a valid HTTP/HTTPS URL")
	}
	
	return nil
}

// validateDataset validates the dataset configuration
func validateDataset(dataset *DatasetConfig) error {
	if dataset.Count <= 0 {
		return fmt.Errorf("record count must be positive, got %d", dataset.Count)
	}
	
	if dataset.Count > 10000000 { // 10M limit
		return fmt.Errorf("record count too large (max 10M), got %d", dataset.Count)
	}
	
	if dataset.Domain == "" {
		return fmt.Errorf("domain description is required")
	}
	
	if len(dataset.Fields) == 0 {
		return fmt.Errorf("at least one field is required")
	}
	
	if len(dataset.Fields) > 100 {
		return fmt.Errorf("too many fields (max 100), got %d", len(dataset.Fields))
	}
	
	// Validate fields
	fieldNames := make(map[string]bool)
	for i, field := range dataset.Fields {
		if err := validateField(&field, i); err != nil {
			return fmt.Errorf("field '%s': %w", field.Name, err)
		}
		
		// Check for duplicate field names
		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name: %s", field.Name)
		}
		fieldNames[field.Name] = true
	}
	
	return nil
}

// validateField validates a single field configuration
func validateField(field *Field, index int) error {
	if field.Name == "" {
		return fmt.Errorf("field name is required (field %d)", index)
	}
	
	// Validate field name format
	if !isValidFieldName(field.Name) {
		return fmt.Errorf("invalid field name '%s' (must be alphanumeric with underscores)", field.Name)
	}
	
	if field.Type == "" {
		return fmt.Errorf("field type is required")
	}
	
	// Validate field type
	validTypes := []string{
		"string", "text", "integer", "float", "boolean", "datetime", "date", "time",
		"email", "url", "uuid", "phone", "enum", "object", "array",
	}
	
	isValidType := false
	for _, validType := range validTypes {
		if field.Type == validType {
			isValidType = true
			break
		}
	}
	
	if !isValidType {
		return fmt.Errorf("invalid field type '%s', valid types: %v", field.Type, validTypes)
	}
	
	// Type-specific validation
	switch field.Type {
	case "enum":
		if len(field.Values) == 0 {
			return fmt.Errorf("enum type requires values")
		}
	case "integer", "float":
		if len(field.Range) > 0 && len(field.Range) != 2 {
			return fmt.Errorf("range must have exactly 2 values [min, max]")
		}
		if len(field.Range) == 2 && field.Range[0] >= field.Range[1] {
			return fmt.Errorf("range min (%d) must be less than max (%d)", field.Range[0], field.Range[1])
		}
	}
	
	// Validate pattern if provided
	if field.Pattern != "" {
		if _, err := regexp.Compile(field.Pattern); err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", field.Pattern, err)
		}
	}
	
	return nil
}

// isValidFieldName checks if a field name is valid (alphanumeric + underscores)
func isValidFieldName(name string) bool {
	if len(name) == 0 || len(name) > 50 {
		return false
	}
	
	// Must start with letter or underscore
	if !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z') || name[0] == '_') {
		return false
	}
	
	// Rest can be alphanumeric or underscore
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	
	return true
}
