# Changelog

All notable changes to FauxFoundry will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive CI/CD pipeline with GitHub Actions
- Docker support with multi-stage builds
- Professional Makefile with 60+ targets
- Advanced timeout handling with intelligent retry strategies
- Healthcare EDI specifications (X12 837, NCPDP D.0)
- Logo and professional branding

### Changed
- Repository ownership transferred to copyleftdev
- Enhanced README with comprehensive documentation
- Improved error handling and user feedback

### Fixed
- Module import paths updated to copyleftdev organization
- Build system compatibility across platforms

## [0.1.0] - 2024-01-15

### Added
- Initial release of FauxFoundry
- CLI framework with Cobra
- Terminal User Interface (TUI) with Bubble Tea
- YAML specification parsing and validation
- Ollama LLM integration
- Real-time data generation with progress tracking
- Deduplication system with canonical hashing
- Streaming JSONL output with optional GZIP compression
- Example specifications for various domains
- Health check system (`fauxfoundry doctor`)
- Template-based specification initialization

### Core Features
- **CLI Commands**:
  - `generate` - Generate synthetic data from specifications
  - `validate` - Validate YAML specifications
  - `init` - Create new specifications from templates
  - `tui` - Launch interactive terminal interface
  - `doctor` - System health diagnostics

- **Data Generation**:
  - Support for 14+ field types (string, integer, float, email, etc.)
  - Field constraints (patterns, ranges, enums)
  - Realistic, domain-aware data generation
  - 100% unique record guarantee
  - Batch processing with configurable sizes

- **Healthcare Support**:
  - Medical insurance verification data
  - EDI X12 healthcare transactions
  - HIPAA-compliant synthetic data patterns
  - Healthcare-specific field types and validations

- **Developer Experience**:
  - Rich TUI with keyboard shortcuts (F1-F10)
  - Real-time validation and error reporting
  - Comprehensive help system
  - Example specifications included
  - Streaming output for memory efficiency

### Technical Details
- **Language**: Go 1.21+
- **Architecture**: Modular design with internal packages
- **Dependencies**: Minimal external dependencies
- **Performance**: ~1-2 records/second with LLM generation
- **Memory**: Constant memory usage via streaming
- **Platforms**: Linux, macOS, Windows support

### Documentation
- Comprehensive README with usage examples
- Inline code documentation
- Example specifications for multiple domains
- Health check and setup guidance

---

## Release Notes Format

### Types of Changes
- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` for vulnerability fixes

### Version Numbering
FauxFoundry follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Release Process
1. Update CHANGELOG.md with new version
2. Tag release in Git: `git tag -a v1.0.0 -m "Release v1.0.0"`
3. Push tags: `git push origin --tags`
4. GitHub Actions automatically builds and publishes release
5. Update documentation and examples as needed

---

**FauxFoundry** - Generate synthetic data with confidence 🎯

*Created by [copyleftdev](https://github.com/copyleftdev)*
