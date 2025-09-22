package output

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/copyleftdev/faux-foundry/pkg/types"
)

// Writer handles streaming output of generated records
type Writer interface {
	Write(record types.Record) error
	Close() error
	GetPath() string
}

// JSONLWriter writes records in JSON Lines format
type JSONLWriter struct {
	writer     io.WriteCloser
	encoder    *json.Encoder
	path       string
	compressed bool
	recordCount int
}

// NewJSONLWriter creates a new JSONL writer
func NewJSONLWriter(path string) (*JSONLWriter, error) {
	var writer io.WriteCloser
	var compressed bool

	if path == "" || path == "stdout" {
		writer = os.Stdout
		path = "stdout"
	} else {
		// Ensure directory exists
		if dir := filepath.Dir(path); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
		}

		// Open file
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}

		// Check if compression is needed
		if strings.HasSuffix(path, ".gz") {
			compressed = true
			gzWriter := gzip.NewWriter(file)
			writer = &gzipWriteCloser{gzWriter, file}
		} else {
			writer = file
		}
	}

	return &JSONLWriter{
		writer:     writer,
		encoder:    json.NewEncoder(writer),
		path:       path,
		compressed: compressed,
	}, nil
}

// gzipWriteCloser wraps gzip.Writer to close both gzip and underlying file
type gzipWriteCloser struct {
	gzWriter *gzip.Writer
	file     *os.File
}

func (g *gzipWriteCloser) Write(p []byte) (n int, err error) {
	return g.gzWriter.Write(p)
}

func (g *gzipWriteCloser) Close() error {
	if err := g.gzWriter.Close(); err != nil {
		g.file.Close()
		return err
	}
	return g.file.Close()
}

// Write writes a single record
func (w *JSONLWriter) Write(record types.Record) error {
	if err := w.encoder.Encode(record); err != nil {
		return fmt.Errorf("failed to encode record: %w", err)
	}
	w.recordCount++
	return nil
}

// WriteRecords writes multiple records
func (w *JSONLWriter) WriteRecords(records []types.Record) error {
	for _, record := range records {
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the writer
func (w *JSONLWriter) Close() error {
	if w.writer != os.Stdout {
		return w.writer.Close()
	}
	return nil
}

// GetPath returns the output path
func (w *JSONLWriter) GetPath() string {
	return w.path
}

// GetRecordCount returns the number of records written
func (w *JSONLWriter) GetRecordCount() int {
	return w.recordCount
}

// IsCompressed returns whether the output is compressed
func (w *JSONLWriter) IsCompressed() bool {
	return w.compressed
}

// StreamingWriter provides a streaming interface for writing records
type StreamingWriter struct {
	writer Writer
	buffer []types.Record
	bufferSize int
}

// NewStreamingWriter creates a new streaming writer with buffering
func NewStreamingWriter(path string, bufferSize int) (*StreamingWriter, error) {
	writer, err := NewJSONLWriter(path)
	if err != nil {
		return nil, err
	}

	return &StreamingWriter{
		writer:     writer,
		buffer:     make([]types.Record, 0, bufferSize),
		bufferSize: bufferSize,
	}, nil
}

// Write adds a record to the buffer and flushes if needed
func (s *StreamingWriter) Write(record types.Record) error {
	s.buffer = append(s.buffer, record)
	
	if len(s.buffer) >= s.bufferSize {
		return s.Flush()
	}
	
	return nil
}

// Flush writes all buffered records
func (s *StreamingWriter) Flush() error {
	if len(s.buffer) == 0 {
		return nil
	}

	for _, record := range s.buffer {
		if err := s.writer.Write(record); err != nil {
			return err
		}
	}

	s.buffer = s.buffer[:0] // Clear buffer
	return nil
}

// Close flushes remaining records and closes the writer
func (s *StreamingWriter) Close() error {
	if err := s.Flush(); err != nil {
		return err
	}
	return s.writer.Close()
}

// GetPath returns the output path
func (s *StreamingWriter) GetPath() string {
	return s.writer.GetPath()
}

// GetRecordCount returns the number of records written
func (s *StreamingWriter) GetRecordCount() int {
	if jsonlWriter, ok := s.writer.(*JSONLWriter); ok {
		return jsonlWriter.GetRecordCount()
	}
	return 0
}
