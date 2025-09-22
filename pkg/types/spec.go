package types

import (
	"fmt"
	"time"
)

// Specification represents a complete FauxFoundry specification
type Specification struct {
	Model   ModelConfig   `yaml:"model" json:"model"`
	Dataset DatasetConfig `yaml:"dataset" json:"dataset"`
}

// ModelConfig defines the LLM backend configuration
type ModelConfig struct {
	Endpoint    string  `yaml:"endpoint" json:"endpoint"`
	Name        string  `yaml:"name" json:"name"`
	BatchSize   int     `yaml:"batch_size" json:"batch_size"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	Timeout     string  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// DatasetConfig defines the dataset generation parameters
type DatasetConfig struct {
	Count  int     `yaml:"count" json:"count"`
	Domain string  `yaml:"domain" json:"domain"`
	Fields []Field `yaml:"fields" json:"fields"`
}

// Field defines a single field in the generated dataset
type Field struct {
	Name        string      `yaml:"name" json:"name"`
	Type        string      `yaml:"type" json:"type"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool        `yaml:"required,omitempty" json:"required,omitempty"`
	Pattern     string      `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Range       []int       `yaml:"range,omitempty" json:"range,omitempty"`
	Values      []string    `yaml:"values,omitempty" json:"values,omitempty"`
	Default     interface{} `yaml:"default,omitempty" json:"default,omitempty"`
}

// GenerationJob represents an active data generation job
type GenerationJob struct {
	ID           string        `json:"id"`
	Spec         Specification `json:"spec"`
	Status       JobStatus     `json:"status"`
	Progress     Progress      `json:"progress"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	OutputPath   string        `json:"output_path"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// JobStatus represents the current status of a generation job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusRunning    JobStatus = "running"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// Progress tracks the progress of data generation
type Progress struct {
	Generated    int     `json:"generated"`
	Target       int     `json:"target"`
	Duplicates   int     `json:"duplicates"`
	Rate         float64 `json:"rate"` // records per second
	ElapsedTime  string  `json:"elapsed_time"`
	EstimatedETA string  `json:"estimated_eta"`
}

// Record represents a single generated data record
type Record map[string]interface{}

// ValidationError represents a specification validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}
