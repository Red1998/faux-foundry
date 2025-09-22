# Contributing to FauxFoundry

Thank you for your interest in contributing to FauxFoundry! This document provides guidelines and information for contributors.

## 🎯 Getting Started

### Prerequisites

- Go 1.21 or later
- [Ollama](https://ollama.ai) for testing LLM integration
- Git for version control

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/yourusername/faux-foundry
   cd faux-foundry
   ```

2. **Set up Development Environment**
   ```bash
   make setup
   make tools
   ```

3. **Verify Installation**
   ```bash
   make build
   make test
   make health
   ```

## 🔄 Development Workflow

### 1. Create a Feature Branch
```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes
- Write clean, well-documented code
- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes
```bash
# Run all tests
make test

# Run quality checks
make quality

# Test with examples
make examples

# Run integration tests
make test-integration
```

### 4. Commit Your Changes
```bash
git add .
git commit -m "feat: add your feature description"
```

We follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes
- `refactor:` - Code refactoring
- `test:` - Test additions or modifications
- `chore:` - Maintenance tasks

### 5. Push and Create Pull Request
```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub with:
- Clear description of changes
- Reference to any related issues
- Screenshots for UI changes
- Test results

## 📋 Code Guidelines

### Go Code Style
- Follow `gofmt` formatting
- Use `golangci-lint` for linting
- Write meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused

### Testing
- Write unit tests for all new functions
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for >80% code coverage

### Documentation
- Update README.md for user-facing changes
- Add inline code comments
- Update examples if needed
- Document new CLI flags or commands

## 🏗️ Project Structure

```
faux-foundry/
├── cmd/fauxfoundry/     # Main application entry point
├── internal/            # Internal packages
│   ├── cli/            # CLI commands
│   ├── tui/            # Terminal UI
│   ├── llm/            # LLM integration
│   ├── spec/           # Specification parsing
│   ├── dedup/          # Deduplication logic
│   └── output/         # Output writers
├── pkg/types/          # Shared type definitions
├── examples/           # Example specifications
├── docs/              # Documentation
└── tests/             # Test files
```

## 🎯 Areas for Contribution

### High Priority
- [ ] Additional LLM provider support (OpenAI, Anthropic, etc.)
- [ ] More healthcare EDI specifications
- [ ] Performance optimizations
- [ ] Enhanced TUI features
- [ ] Documentation improvements

### Medium Priority
- [ ] Additional output formats (CSV, Parquet, etc.)
- [ ] Data validation enhancements
- [ ] Template system improvements
- [ ] Monitoring and metrics
- [ ] Docker improvements

### Low Priority
- [ ] Web UI interface
- [ ] Cloud deployment options
- [ ] Advanced analytics
- [ ] Plugin system
- [ ] Multi-language support

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment Information**
   - OS and version
   - Go version
   - FauxFoundry version
   - Ollama version (if applicable)

2. **Steps to Reproduce**
   - Clear, numbered steps
   - Example specification files
   - Command line arguments used

3. **Expected vs Actual Behavior**
   - What you expected to happen
   - What actually happened
   - Error messages or logs

4. **Additional Context**
   - Screenshots if applicable
   - Related issues or PRs
   - Workarounds you've tried

## 💡 Feature Requests

For new features, please:

1. **Check Existing Issues** - Avoid duplicates
2. **Describe the Problem** - What need does this address?
3. **Propose a Solution** - How should it work?
4. **Consider Alternatives** - What other approaches exist?
5. **Additional Context** - Examples, mockups, etc.

## 🔍 Code Review Process

All contributions go through code review:

1. **Automated Checks** - CI/CD pipeline runs tests
2. **Maintainer Review** - Core team reviews code
3. **Community Feedback** - Other contributors may comment
4. **Approval and Merge** - Changes are merged after approval

### Review Criteria
- Code quality and style
- Test coverage
- Documentation completeness
- Performance impact
- Backward compatibility

## 🏥 Healthcare Data Guidelines

When contributing healthcare-related specifications:

### Compliance
- Ensure HIPAA compliance (no real PHI)
- Follow industry standards (HL7, X12, NCPDP)
- Use realistic but synthetic data patterns

### Quality Standards
- Accurate field relationships
- Proper data types and formats
- Realistic value distributions
- Comprehensive field coverage

### Documentation
- Explain medical terminology
- Reference relevant standards
- Provide usage examples
- Include field descriptions

## 🤝 Community Guidelines

### Be Respectful
- Use inclusive language
- Be constructive in feedback
- Help newcomers learn
- Celebrate contributions

### Be Professional
- Focus on technical merit
- Avoid personal attacks
- Keep discussions on-topic
- Follow the code of conduct

### Be Collaborative
- Share knowledge freely
- Credit others' work
- Seek consensus on major changes
- Support community growth

## 📞 Getting Help

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - Questions and general discussion
- **Documentation** - Check README and docs/ directory
- **Examples** - Review examples/ directory

## 🎉 Recognition

Contributors are recognized through:
- GitHub contributor graphs
- Release notes mentions
- README acknowledgments
- Community shout-outs

## 📄 License

By contributing to FauxFoundry, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to FauxFoundry!** 🎯

*Building tools for developers, by developers.*
