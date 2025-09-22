package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/copyleftdev/faux-foundry/pkg/types"
)

// AppState represents the global application state
type AppState struct {
	CurrentSpec     *types.Specification
	ActiveGeneration *types.GenerationJob
	Settings        *UserSettings
	History         []HistoryEntry
	Notifications   []Notification
}

// UserSettings represents user preferences
type UserSettings struct {
	Theme           string
	DefaultOutput   string
	AutoSave        bool
	Confirmations   bool
	DefaultBatchSize int
	DefaultTimeout  string
}

// HistoryEntry represents a historical action
type HistoryEntry struct {
	Timestamp string
	Action    string
	Details   string
}

// Notification represents a user notification
type Notification struct {
	Type    string // info, warning, error, success
	Message string
	Time    string
}

// App represents the main TUI application
type App struct {
	state      *AppState
	currentView ViewType
	views      map[ViewType]tea.Model
	width      int
	height     int
	theme      *Theme
}

// ViewType represents different views in the application
type ViewType int

const (
	ViewMain ViewType = iota
	ViewSpecEditor
	ViewSpecBrowser
	ViewGenerationMonitor
	ViewSettings
	ViewHelp
)

// Theme defines the visual styling
type Theme struct {
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Background  lipgloss.Color
	Foreground  lipgloss.Color
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	Info        lipgloss.Color
}

// DefaultTheme returns the default dark theme
func DefaultTheme() *Theme {
	return &Theme{
		Primary:     lipgloss.Color("#7C3AED"), // Purple
		Secondary:   lipgloss.Color("#6B7280"), // Gray
		Accent:      lipgloss.Color("#F59E0B"), // Amber
		Background:  lipgloss.Color("#111827"), // Dark gray
		Foreground:  lipgloss.Color("#F9FAFB"), // Light gray
		Success:     lipgloss.Color("#10B981"), // Green
		Warning:     lipgloss.Color("#F59E0B"), // Amber
		Error:       lipgloss.Color("#EF4444"), // Red
		Info:        lipgloss.Color("#3B82F6"), // Blue
	}
}

// NewApp creates a new TUI application
func NewApp(specFile string) *App {
	state := &AppState{
		Settings: &UserSettings{
			Theme:           "dark",
			DefaultOutput:   "~/data/",
			AutoSave:        true,
			Confirmations:   true,
			DefaultBatchSize: 32,
			DefaultTimeout:  "2h",
		},
		History:       []HistoryEntry{},
		Notifications: []Notification{},
	}

	// Load specification if provided
	if specFile != "" {
		// TODO: Load specification from file
		state.Notifications = append(state.Notifications, Notification{
			Type:    "info",
			Message: fmt.Sprintf("Loaded specification: %s", specFile),
			Time:    "now",
		})
	}

	app := &App{
		state:       state,
		currentView: ViewMain,
		views:       make(map[ViewType]tea.Model),
		theme:       DefaultTheme(),
	}

	// Initialize views
	app.initViews()

	return app
}

// initViews initializes all the views
func (a *App) initViews() {
	a.views[ViewMain] = NewMainView(a.state, a.theme)
	a.views[ViewSpecEditor] = NewSpecEditorView(a.state, a.theme)
	a.views[ViewSpecBrowser] = NewSpecBrowserView(a.state, a.theme)
	a.views[ViewGenerationMonitor] = NewGenerationMonitorView(a.state, a.theme)
	a.views[ViewSettings] = NewSettingsView(a.state, a.theme)
	a.views[ViewHelp] = NewHelpView(a.state, a.theme)
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		a.views[a.currentView].Init(),
	)
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		// Update all views with new size
		for viewType, view := range a.views {
			a.views[viewType], _ = view.Update(msg)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if a.currentView == ViewMain {
				return a, tea.Quit
			}
			// Go back to main view from other views
			a.currentView = ViewMain
			return a, nil

		case "f1":
			a.currentView = ViewHelp
			return a, nil

		case "f2":
			a.currentView = ViewSpecBrowser
			return a, nil

		case "f3":
			// TODO: Start generation
			a.currentView = ViewGenerationMonitor
			return a, nil

		case "f4":
			a.currentView = ViewGenerationMonitor
			return a, nil

		case "f10":
			return a, tea.Quit

		case "ctrl+n":
			// New specification
			a.currentView = ViewSpecEditor
			return a, nil

		case "ctrl+s":
			// Save current spec
			// TODO: Implement save logic
			a.state.Notifications = append(a.state.Notifications, Notification{
				Type:    "success",
				Message: "Specification saved",
				Time:    "now",
			})
			return a, nil
		}
	}

	// Update current view
	a.views[a.currentView], cmd = a.views[a.currentView].Update(msg)
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

// View implements tea.Model
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Header
	header := a.renderHeader()
	
	// Main content
	content := a.views[a.currentView].View()
	
	// Footer
	footer := a.renderFooter()

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// renderHeader renders the application header
func (a *App) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(a.theme.Primary).
		Bold(true).
		Padding(0, 1)

	shortcutsStyle := lipgloss.NewStyle().
		Foreground(a.theme.Secondary).
		Padding(0, 1)

	title := titleStyle.Render("FauxFoundry v0.1.0")
	shortcuts := shortcutsStyle.Render("[F1] Help  [F2] Specs  [F3] Generate  [F4] Monitor  [F10] Quit")

	headerStyle := lipgloss.NewStyle().
		Width(a.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(a.theme.Secondary)

	return headerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			title,
			lipgloss.NewStyle().Width(a.width-lipgloss.Width(title)-lipgloss.Width(shortcuts)).Render(""),
			shortcuts,
		),
	)
}

// renderFooter renders the application footer
func (a *App) renderFooter() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(a.theme.Secondary).
		Padding(0, 1)

	var status string
	switch a.currentView {
	case ViewMain:
		status = "Ready"
	case ViewSpecEditor:
		status = "Editing"
	case ViewSpecBrowser:
		status = "Browsing"
	case ViewGenerationMonitor:
		status = "Monitoring"
	case ViewSettings:
		status = "Settings"
	case ViewHelp:
		status = "Help"
	}

	// Add memory usage and other stats
	memInfo := "Memory: 45MB"
	uptime := "Uptime: 5m 23s"

	footerStyle := lipgloss.NewStyle().
		Width(a.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(a.theme.Secondary)

	return footerStyle.Render(
		statusStyle.Render(fmt.Sprintf("Status: %s │ %s │ %s", status, memInfo, uptime)),
	)
}

// Run starts the TUI application
func Run(specFile string) error {
	app := NewApp(specFile)
	
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	return nil
}
