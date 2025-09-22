# Design Specification: FauxFoundry

**A comprehensive design specification for FauxFoundry's CLI and TUI interfaces**

## 1. Executive Summary

FauxFoundry provides both a traditional CLI interface for automation and scripting, and a rich Terminal User Interface (TUI) for interactive data generation workflows. The design emphasizes discoverability, real-time feedback, and operator empathy while maintaining the tool's core principle of simplicity.

## 2. Design Principles

### Core UX Principles
- **Progressive Disclosure**: Start simple, reveal complexity as needed
- **Real-time Feedback**: Show progress, validation, and results immediately
- **Keyboard-First**: Optimized for keyboard navigation with mouse support
- **Contextual Help**: Always-available guidance without overwhelming the interface
- **Error Prevention**: Validate inputs before execution, clear error recovery

### Visual Design Principles
- **Minimal Distraction**: Clean, focused interface with purposeful use of color
- **Information Hierarchy**: Clear visual hierarchy for different types of information
- **Consistent Patterns**: Reusable UI patterns across all workflows
- **Accessibility**: High contrast, clear typography, screen reader friendly

## 3. Architecture Overview

### Interface Modes
1. **CLI Mode**: Traditional command-line interface for automation
2. **TUI Mode**: Interactive terminal interface for guided workflows
3. **Hybrid Mode**: CLI commands with TUI feedback and progress

### Component Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                    FauxFoundry Interface                    │
├─────────────────────────────────────────────────────────────┤
│  CLI Layer          │  TUI Layer          │  Shared Core    │
│  ┌─────────────┐    │  ┌─────────────┐    │  ┌─────────────┐ │
│  │ Cobra CLI   │    │  │ Bubble Tea  │    │  │ Spec Parser │ │
│  │ Commands    │    │  │ Components  │    │  │ LLM Client  │ │
│  │ Flags       │    │  │ Views       │    │  │ Dedup Logic │ │
│  │ Validation  │    │  │ Models      │    │  │ Output      │ │
│  └─────────────┘    │  └─────────────┘    │  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 4. CLI Interface Design

### Command Structure
```bash
fauxfoundry [global-flags] <command> [command-flags] [args]
```

### Core Commands

#### `generate` - Generate synthetic data
```bash
# Basic usage
fauxfoundry generate --spec customer.yaml --count 1000

# With output options
fauxfoundry generate --spec customer.yaml --output data.jsonl.gz --verbose

# Interactive mode
fauxfoundry generate --interactive
```

#### `validate` - Validate specifications
```bash
# Validate spec file
fauxfoundry validate customer.yaml

# Dry run with validation
fauxfoundry validate --dry-run customer.yaml
```

#### `init` - Initialize new specifications
```bash
# Create new spec interactively
fauxfoundry init customer.yaml

# From template
fauxfoundry init --template ecommerce customer.yaml
```

#### `tui` - Launch interactive interface
```bash
# Launch TUI mode
fauxfoundry tui

# Launch with specific spec
fauxfoundry tui --spec customer.yaml
```

### Global Flags
- `--config`: Configuration file path
- `--verbose`: Enable verbose logging
- `--quiet`: Suppress non-essential output
- `--no-color`: Disable colored output
- `--help`: Show help information

## 5. TUI Interface Design

### Main Application Layout
```
┌─ FauxFoundry v0.1.0 ──────────────────────────────────────────────────────┐
│ [F1] Help  [F2] Specs  [F3] Generate  [F4] Monitor  [F10] Quit            │
├───────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  ┌─ Current Specification ─────────────────────────────────────────────┐  │
│  │ customer.yaml                                    [Edit] [Validate]  │  │
│  │ Domain: E-commerce customer data                                    │  │
│  │ Fields: 5 (email, age, status, created_at, preferences)            │  │
│  │ Target: 1,000 records                                              │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                           │
│  ┌─ Generation Status ─────────────────────────────────────────────────┐  │
│  │ ● Ready to generate                                                 │  │
│  │ Model: llama3.1:8b (Connected)                                     │  │
│  │ Estimated time: ~2 minutes                                         │  │
│  │                                                                     │  │
│  │ [Generate] [Settings] [Preview]                                     │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                           │
│  ┌─ Recent Activity ───────────────────────────────────────────────────┐  │
│  │ 12:30 PM - customer.yaml validated successfully                    │  │
│  │ 12:28 PM - Generated 500 product records                           │  │
│  │ 12:25 PM - Created new specification: products.yaml                │  │
│  └─────────────────────────────────────────────────────────────────────┘  │
│                                                                           │
└─ Status: Ready │ Memory: 45MB │ Uptime: 5m 23s ────────────────────────────┘
```

### TUI Components

#### 1. Specification Editor
```
┌─ Specification Editor: customer.yaml ─────────────────────────────────────┐
│                                                                           │
│ Model Configuration:                                                      │
│ ┌─────────────────────────────────────────────────────────────────────┐   │
│ │ Endpoint: http://localhost:11434                    [Test Connection] │   │
│ │ Model: llama3.1:8b                                 [Browse Models ▼] │   │
│ │ Batch Size: 32                                     [────────────────] │   │
│ │ Temperature: 0.7                                   [████████░░░░░░░░] │   │
│ └─────────────────────────────────────────────────────────────────────┘   │
│                                                                           │
│ Dataset Configuration:                                                    │
│ ┌─────────────────────────────────────────────────────────────────────┐   │
│ │ Count: 1000                                        [────────────────] │   │
│ │ Domain: E-commerce customer data                   [────────────────] │   │
│ └─────────────────────────────────────────────────────────────────────┘   │
│                                                                           │
│ Fields:                                                    [Add Field +] │
│ ┌─────────────────────────────────────────────────────────────────────┐   │
│ │ ✓ email      │ email    │ @(gmail|yahoo|outlook)\.com$  │ [Edit] [×] │   │
│ │ ✓ age        │ integer  │ range: [18, 80]               │ [Edit] [×] │   │
│ │ ✓ status     │ enum     │ active, inactive, pending     │ [Edit] [×] │   │
│ │ ✓ created_at │ datetime │ 2020-01-01 to now            │ [Edit] [×] │   │
│ │ ✓ preferences│ object   │ nested customer preferences   │ [Edit] [×] │   │
│ └─────────────────────────────────────────────────────────────────────┘   │
│                                                                           │
│ [Save] [Save As...] [Validate] [Preview] [Cancel]                        │
└───────────────────────────────────────────────────────────────────────────┘
```

#### 2. Generation Monitor
```
┌─ Generation Progress ──────────────────────────────────────────────────────┐
│                                                                           │
│ Generating customer.yaml → data.jsonl.gz                                 │
│                                                                           │
│ Progress: ████████████████████████████████████████████░░░░░░░░ 847/1000  │
│ Time: 1m 23s elapsed, ~15s remaining                                     │
│                                                                           │
│ ┌─ Live Statistics ─────────────────────────────────────────────────────┐ │
│ │ Records Generated: 847                                                │ │
│ │ Unique Records: 847 (100%)                                           │ │
│ │ Duplicates Rejected: 12                                              │ │
│ │ Generation Rate: 10.2 records/sec                                    │ │
│ │ Memory Usage: 52MB (constant)                                        │ │
│ │ LLM Requests: 27 batches                                             │ │
│ └───────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│ ┌─ Recent Records ──────────────────────────────────────────────────────┐ │
│ │ {"email":"sarah.jones@gmail.com","age":29,"status":"active",...}     │ │
│ │ {"email":"mike.wilson@yahoo.com","age":34,"status":"pending",...}    │ │
│ │ {"email":"lisa.brown@outlook.com","age":42,"status":"inactive",...}  │ │
│ └───────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│ [Pause] [Cancel] [View Output] [Save Progress]                           │
└───────────────────────────────────────────────────────────────────────────┘
```

#### 3. Specification Browser
```
┌─ Specification Browser ───────────────────────────────────────────────────┐
│                                                                           │
│ Filter: [All Types ▼] [Search: ________________] [Sort: Modified ▼]       │
│                                                                           │
│ ┌─ Specifications ──────────────────────────────────────────────────────┐ │
│ │ 📄 customer.yaml        E-commerce customers    1,000 records   2m ago│ │
│ │ 📄 products.yaml        Product catalog         5,000 records   1h ago│ │
│ │ 📄 orders.yaml          Order transactions      2,500 records   3h ago│ │
│ │ 📄 reviews.yaml         Product reviews         10,000 records  1d ago│ │
│ │ 📄 inventory.yaml       Inventory tracking      1,500 records   2d ago│ │
│ └───────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│ ┌─ Preview: customer.yaml ──────────────────────────────────────────────┐ │
│ │ Domain: E-commerce customer data                                      │ │
│ │ Fields: email, age, status, created_at, preferences                   │ │
│ │ Model: llama3.1:8b                                                    │ │
│ │ Last Generated: 2 minutes ago (1,000 records)                        │ │
│ │                                                                       │ │
│ │ Sample Output:                                                        │ │
│ │ {"email":"john.doe@gmail.com","age":34,"status":"active",...}        │ │
│ └───────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│ [New] [Edit] [Duplicate] [Generate] [Delete] [Export]                    │
└───────────────────────────────────────────────────────────────────────────┘
```

#### 4. Settings Panel
```
┌─ Settings ─────────────────────────────────────────────────────────────────┐
│                                                                           │
│ ┌─ General ─────────────────────────────────────────────────────────────┐ │
│ │ Default Output Directory: ~/data/                 [Browse...]         │ │
│ │ Auto-save Specifications: ☑ Enabled                                  │ │
│ │ Confirmation Prompts: ☑ Enabled                                      │ │
│ │ Theme: Dark ▼                                                         │ │
│ └───────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│ ┌─ Performance ─────────────────────────────────────────────────────────┐ │
│ │ Default Batch Size: 32                            [────────────────]  │ │
│ │ Max Concurrent Requests: 2                        [────────────────]  │ │
│ │ Request Timeout: 30s                              [────────────────]  │ │
│ │ Retry Attempts: 3                                 [────────────────]  │ │
│ └───────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│ ┌─ LLM Backends ────────────────────────────────────────────────────────┐ │
│ │ ☑ Ollama (http://localhost:11434)                [Test] [Configure]   │ │
│ │ ☐ OpenAI API                                      [Test] [Configure]   │ │
│ │ ☐ Custom Endpoint                                 [Test] [Configure]   │ │
│ └───────────────────────────────────────────────────────────────────────┘ │
│                                                                           │
│ [Save] [Reset to Defaults] [Cancel]                                      │
└───────────────────────────────────────────────────────────────────────────┘
```

## 6. Workflow Design

### Workflow 1: Quick Generation
```
User Journey: Generate data from existing spec
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Launch TUI  │ -> │ Select Spec │ -> │ Configure   │ -> │ Generate    │
│             │    │ from Browser│    │ Parameters  │    │ & Monitor   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Workflow 2: Spec Creation
```
User Journey: Create new specification from scratch
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ New Spec    │ -> │ Choose      │ -> │ Define      │ -> │ Test &      │
│ Wizard      │    │ Template    │    │ Fields      │    │ Validate    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Workflow 3: Iterative Refinement
```
User Journey: Refine spec based on output quality
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Generate    │ -> │ Review      │ -> │ Adjust      │ -> │ Re-generate │
│ Sample      │    │ Output      │    │ Spec        │    │ & Compare   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

## 7. Keyboard Navigation

### Global Shortcuts
- `F1`: Help/Documentation
- `F2`: Specification Browser
- `F3`: Generate Data
- `F4`: Monitor Active Generation
- `F10`: Quit Application
- `Ctrl+N`: New Specification
- `Ctrl+O`: Open Specification
- `Ctrl+S`: Save Current Spec
- `Ctrl+Q`: Quit
- `Ctrl+C`: Cancel Current Operation

### Navigation Shortcuts
- `Tab/Shift+Tab`: Navigate between components
- `Enter`: Activate/Select
- `Escape`: Cancel/Go Back
- `Arrow Keys`: Navigate lists/menus
- `Page Up/Down`: Scroll large content
- `Home/End`: Jump to beginning/end

### Context-Specific Shortcuts
- `Space`: Toggle checkboxes/selections
- `Delete`: Remove selected item
- `F5`: Refresh/Reload
- `Ctrl+F`: Search/Filter
- `Ctrl+D`: Duplicate item

## 8. Error Handling & Feedback

### Error States
1. **Connection Errors**: Clear indication when LLM backend is unavailable
2. **Validation Errors**: Inline validation with specific field highlighting
3. **Generation Errors**: Graceful handling with retry options
4. **File System Errors**: Clear messages for permission/disk space issues

### Success Feedback
1. **Progress Indicators**: Real-time progress bars and statistics
2. **Completion Notifications**: Clear success messages with next actions
3. **Visual Confirmations**: Color-coded status indicators
4. **Sound Feedback**: Optional audio cues for completion (configurable)

## 9. Accessibility Features

### Visual Accessibility
- High contrast color schemes
- Configurable font sizes
- Color-blind friendly palettes
- Clear visual hierarchy

### Keyboard Accessibility
- Full keyboard navigation
- Logical tab order
- Visible focus indicators
- Standard keyboard shortcuts

### Screen Reader Support
- Semantic markup for TUI components
- Descriptive labels and help text
- Status announcements
- Progress updates

## 10. Technical Implementation

### TUI Framework
- **Primary**: Bubble Tea (Go) for reactive TUI components
- **Styling**: Lip Gloss for consistent visual styling
- **Input**: Bubbles for form components and inputs

### Component Library
```go
// Core TUI Components
type Application struct {
    router    *Router
    state     *AppState
    theme     *Theme
    shortcuts *KeyMap
}

type View interface {
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() string
    Init() tea.Cmd
}

// Reusable Components
- SpecEditor
- ProgressMonitor
- SpecBrowser
- SettingsPanel
- HelpViewer
- StatusBar
```

### State Management
```go
type AppState struct {
    CurrentSpec     *Specification
    ActiveGeneration *GenerationJob
    Settings        *UserSettings
    History         []HistoryEntry
    Notifications   []Notification
}
```

## 11. Testing Strategy

### TUI Testing
- **Unit Tests**: Individual component behavior
- **Integration Tests**: Component interaction and state management
- **Visual Tests**: Screenshot comparison for UI consistency
- **Accessibility Tests**: Screen reader compatibility

### User Experience Testing
- **Usability Testing**: Task completion and user satisfaction
- **Performance Testing**: Responsiveness under load
- **Accessibility Testing**: Keyboard navigation and screen readers

## 12. Future Enhancements

### Phase 2 Features
- **Multi-pane Layout**: Side-by-side spec editing and preview
- **Collaborative Features**: Shared specifications and comments
- **Advanced Visualizations**: Data distribution charts and statistics
- **Plugin System**: Custom field types and validators

### Phase 3 Features
- **Web Interface**: Browser-based version of TUI
- **API Integration**: REST API for external tool integration
- **Cloud Sync**: Specification synchronization across devices
- **Advanced Analytics**: Generation performance and quality metrics

---

This design specification provides a comprehensive foundation for implementing both CLI and TUI interfaces that prioritize user experience while maintaining the tool's core principles of simplicity and reliability.
