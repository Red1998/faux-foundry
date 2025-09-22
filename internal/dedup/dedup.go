package dedup

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/copyleftdev/faux-foundry/pkg/types"
)

// Deduplicator handles record deduplication using canonical hashing
type Deduplicator struct {
	seen map[string]bool
	duplicateCount int
}

// NewDeduplicator creates a new deduplicator
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		seen: make(map[string]bool),
	}
}

// IsUnique checks if a record is unique and adds it to the seen set
func (d *Deduplicator) IsUnique(record types.Record) bool {
	hash := d.canonicalHash(record)
	
	if d.seen[hash] {
		d.duplicateCount++
		return false
	}
	
	d.seen[hash] = true
	return true
}

// FilterUnique filters a slice of records to only include unique ones
func (d *Deduplicator) FilterUnique(records []types.Record) []types.Record {
	unique := make([]types.Record, 0, len(records))
	
	for _, record := range records {
		if d.IsUnique(record) {
			unique = append(unique, record)
		}
	}
	
	return unique
}

// GetDuplicateCount returns the number of duplicates encountered
func (d *Deduplicator) GetDuplicateCount() int {
	return d.duplicateCount
}

// GetUniqueCount returns the number of unique records seen
func (d *Deduplicator) GetUniqueCount() int {
	return len(d.seen)
}

// Reset clears the deduplicator state
func (d *Deduplicator) Reset() {
	d.seen = make(map[string]bool)
	d.duplicateCount = 0
}

// canonicalHash creates a canonical hash of a record
func (d *Deduplicator) canonicalHash(record types.Record) string {
	// Create a canonical representation by sorting keys and values
	canonical := d.canonicalize(record)
	
	// Create JSON representation
	jsonBytes, err := json.Marshal(canonical)
	if err != nil {
		// Fallback to string representation if JSON fails
		return fmt.Sprintf("%v", canonical)
	}
	
	// Create SHA256 hash
	hash := sha256.Sum256(jsonBytes)
	return fmt.Sprintf("%x", hash)
}

// canonicalize creates a canonical representation of a record
func (d *Deduplicator) canonicalize(record types.Record) map[string]interface{} {
	canonical := make(map[string]interface{})
	
	// Get sorted keys for consistent ordering
	keys := make([]string, 0, len(record))
	for key := range record {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	// Process each field in sorted order
	for _, key := range keys {
		value := record[key]
		canonical[key] = d.canonicalizeValue(value)
	}
	
	return canonical
}

// canonicalizeValue creates a canonical representation of a value
func (d *Deduplicator) canonicalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		// Normalize strings by trimming whitespace and converting to lowercase for comparison
		// Note: This might be too aggressive for some use cases
		return strings.TrimSpace(v)
		
	case map[string]interface{}:
		// Recursively canonicalize nested objects
		return d.canonicalize(v)
		
	case []interface{}:
		// Canonicalize arrays by sorting if they contain comparable types
		canonical := make([]interface{}, len(v))
		for i, item := range v {
			canonical[i] = d.canonicalizeValue(item)
		}
		
		// Sort array if all elements are strings or numbers
		if d.isComparableArray(canonical) {
			sort.Slice(canonical, func(i, j int) bool {
				return fmt.Sprintf("%v", canonical[i]) < fmt.Sprintf("%v", canonical[j])
			})
		}
		
		return canonical
		
	case nil:
		return nil
		
	default:
		// For other types (numbers, booleans), return as-is
		return v
	}
}

// isComparableArray checks if an array contains only comparable types
func (d *Deduplicator) isComparableArray(arr []interface{}) bool {
	if len(arr) == 0 {
		return true
	}
	
	// Check if all elements are strings or numbers
	for _, item := range arr {
		switch item.(type) {
		case string, int, int64, float64, bool:
			continue
		default:
			return false
		}
	}
	
	return true
}

// BatchDeduplicator handles deduplication for batches of records
type BatchDeduplicator struct {
	*Deduplicator
	batchSize int
}

// NewBatchDeduplicator creates a new batch deduplicator
func NewBatchDeduplicator(batchSize int) *BatchDeduplicator {
	return &BatchDeduplicator{
		Deduplicator: NewDeduplicator(),
		batchSize:    batchSize,
	}
}

// ProcessBatch processes a batch of records and returns unique ones
func (bd *BatchDeduplicator) ProcessBatch(records []types.Record) []types.Record {
	return bd.FilterUnique(records)
}

// GetStats returns deduplication statistics
func (bd *BatchDeduplicator) GetStats() DeduplicationStats {
	return DeduplicationStats{
		UniqueRecords:    bd.GetUniqueCount(),
		DuplicateRecords: bd.GetDuplicateCount(),
		TotalProcessed:   bd.GetUniqueCount() + bd.GetDuplicateCount(),
		DeduplicationRate: float64(bd.GetDuplicateCount()) / float64(bd.GetUniqueCount()+bd.GetDuplicateCount()),
	}
}

// DeduplicationStats contains statistics about the deduplication process
type DeduplicationStats struct {
	UniqueRecords     int     `json:"unique_records"`
	DuplicateRecords  int     `json:"duplicate_records"`
	TotalProcessed    int     `json:"total_processed"`
	DeduplicationRate float64 `json:"deduplication_rate"`
}

// String returns a string representation of the stats
func (s DeduplicationStats) String() string {
	return fmt.Sprintf(
		"Unique: %d, Duplicates: %d, Total: %d, Rate: %.2f%%",
		s.UniqueRecords,
		s.DuplicateRecords,
		s.TotalProcessed,
		s.DeduplicationRate*100,
	)
}
