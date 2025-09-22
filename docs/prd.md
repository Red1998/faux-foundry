# PRD: FauxFoundry

**A CLI and TUI for synthetic, domain-aware data generation powered by local LLMs.**

## 1. Executive Summary

FauxFoundry enables teams to generate unique synthetic datasets from human-readable YAML specifications through both command-line and interactive terminal interfaces. It leverages local AI models (e.g., Ollama) to produce realistic, domain-aware data that respects schema constraints while ensuring exactly N unique records are delivered through efficient streaming with minimal validation overhead.

This tool is designed with system empathy: it maintains constant memory usage, degrades gracefully under errors, and provides both automation-friendly CLI commands and discoverable TUI workflows for different user needs.

## 2. Core Principles

- **Simplicity First**: YAML in, JSONL out. Nothing more, nothing less.
- **LLM-Centric**: Delegate complexity of domain semantics to the model.
- **Graceful Determinism**: Consistent prompts, retries until uniqueness, canonical hashing.
- **Streaming Always**: Constant memory footprint, back-pressure friendly.
- **Operator Empathy**: Easy flags, safe defaults, clear logs, non-zero exit codes when failing.

## 3. Target Users

- **Developers/QA**: Need quick fake data for tests and demos
- **Data Scientists**: Want synthetic versions of sensitive data for analysis
- **Ops & SREs**: Need reproducible, safe CLI that won't consume excessive memory or silently fail
- **Security Teams**: Demand guarantees that no real PII leaks into synthetic datasets

## 4. Inputs

### YAML Specification

- **model**: Backend configuration (Ollama URL, model name, batch size, sampling parameters)
- **dataset**: Target record count, domain description, field rules (type, enum, range, pattern)

**Example YAML:**

```yaml
model:
  endpoint: "http://localhost:11434"
  name: "llama3.1:8b"
  batch_size: 32
  temperature: 0.7

dataset:
  count: 1000
  domain: "E-commerce customer data"
  fields:
    - name: "email"
      type: "email"
      pattern: "@(gmail|yahoo|outlook)\\.com$"
    - name: "age"
      type: "integer"
      range: [18, 80]
    - name: "status"
      type: "enum"
      values: ["active", "inactive", "pending"]
```

### CLI Commands

#### Core Commands

- `generate`: Generate synthetic data from specification
- `validate`: Validate YAML specifications
- `init`: Initialize new specifications interactively
- `tui`: Launch interactive terminal interface

#### CLI Flags

- `--spec path.yaml` (required for generate): Path to YAML specification file
- `--count 5000` (optional): Override record count from spec
- `--output data.jsonl.gz` (optional): Output file (stdout by default, gzip if .gz extension)
- `--timeout 1h` (optional): Maximum execution time (default: 2h)
- `--seed 42` (optional): Reproducibility seed
- `--dry-run` (optional): Validate spec without generating data
- `--verbose` (optional): Enable detailed logging
- `--interactive` (optional): Launch TUI mode for guided workflows

## 5. Outputs

- **Format**: JSON Lines (JSONL) stream, optionally GZIP compressed
- **Uniqueness**: Each record canonicalized for uniqueness via hashing
- **Compatibility**: Compatible with jq, databases, parquet converters, and data pipelines

**Example Output:**

```jsonl
{"email": "john.doe@gmail.com", "age": 34, "status": "active"}
{"email": "jane.smith@yahoo.com", "age": 28, "status": "pending"}
{"email": "bob.wilson@outlook.com", "age": 45, "status": "inactive"}
```

## 6. Functional Requirements

### Core Generation Logic

- Parse YAML specification and validate schema
- Request batches from LLM until exactly N unique records are generated
- Deduplicate via canonical hashing; regenerate on collision
- Stream records to output (file/stdout), never buffering entire dataset

### Validation & Quality

- Minimal validation: ensure valid JSON, required fields present
- Respect field constraints (type, range, enum, pattern)
- Maintain data consistency across related fields

### Error Handling

- Exit with clear non-zero code if unable to reach N after retries
- Graceful handling of LLM timeouts and network issues
- Clear error messages with actionable guidance

## 7. Non-Functional Requirements

### Performance

- **Memory**: O(1) memory usage (streaming, not buffering)
- **Batch Size**: Configurable (default: 32 records per batch)
- **Concurrency**: Configurable parallel requests (default: 2 concurrent batches)
- **Throughput**: Target 100+ records/second on standard hardware

### Reliability

- **Timeouts**: Configurable timeouts with graceful cancellation
- **Retries**: Configurable retry budget per batch (default: 3 attempts)
- **Fault Tolerance**: Graceful degradation on partial failures
- **Interruption**: Clean shutdown on SIGINT/SIGTERM

### Maintainability

- **Codebase Size**: Small, focused codebase (~1k LOC)
- **Architecture**: Clear separation of concerns:
  - Spec parser
  - Generation engine
  - LLM client
  - Output writer
- **Testing**: Comprehensive unit and integration tests
- **Documentation**: Clear API documentation and examples

### Security/Privacy

- **PII Protection**: Never leak real PII data
- **Network Isolation**: No network calls beyond local model endpoint
- **Data Residency**: All processing happens locally
- **Audit Trail**: Optional logging of generation metadata (no data content)

## 8. System Empathy

### For Operators

- Clear, structured logging with configurable levels
- Dry-run option for validation without generation
- Graceful interrupts with cleanup
- Progress indicators for long-running operations

### For Developers

- Single static Go binary with no external dependencies
- Simple installation and deployment
- Clear error messages with debugging context
- Comprehensive CLI help and examples

### For Data Pipelines

- Deterministic behavior: always produces exactly N records or fails
- Never silently drops records or produces partial results
- Exit codes that clearly indicate success/failure reasons
- Streaming output compatible with Unix pipes

### For Future Maintainers

- Minimal moving parts and dependencies
- Explicit interfaces and clear abstractions
- Readable, well-documented code
- Comprehensive test coverage

## 9. Success Metrics

- **Adoption**: 100+ GitHub stars within 6 months
- **Performance**: Generate 1M records in under 10 minutes
- **Reliability**: 99.9% success rate for valid specifications
- **Developer Experience**: Setup to first synthetic data in under 5 minutes

## 10. Dependencies

### Required

- **Go 1.21+**: For building the application
- **Local LLM**: Ollama or compatible API endpoint

### Optional

- **Docker**: For containerized deployment
- **jq**: For output processing and validation

## 11. Testing Strategy

### Unit Tests

- YAML specification parsing and validation
- Record deduplication and hashing logic
- LLM client error handling and retries

### Integration Tests

- End-to-end generation workflows
- Output format validation
- Performance benchmarks

### Acceptance Tests

- Real-world specification examples
- Multi-domain data generation scenarios
- Error recovery and edge cases

## 12. Roadmap

### v0.1 (MVP)

- YAML → JSONL unique record generation
- Ollama backend support only
- Basic CLI with essential flags (`generate`, `validate`, `init`)
- Interactive TUI for guided workflows
- Streaming output with deduplication

### v0.2 (Enhanced)

- Multi-table relations and foreign keys
- Additional LLM backend support (OpenAI API, etc.)
- Advanced field types and constraints
- Performance optimizations

### v0.3 (Production Ready)

- Configuration file support
- Advanced logging and monitoring
- Docker containerization
- CI/CD pipeline and releases

### v0.4 (Extended)

- Lightweight validators (UUID, E164, date formats)
- Plugin system for custom field types
- Advanced TUI features (multi-pane layout, collaborative editing)
- Integration with popular data tools