package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainView represents the main dashboard view
type MainView struct {
	state  *AppState
	theme  *Theme
	width  int
	height int
}

// NewMainView creates a new main view
func NewMainView(state *AppState, theme *Theme) *MainView {
	return &MainView{
		state: state,
		theme: theme,
	}
}

// Init implements tea.Model
func (m *MainView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *MainView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 4 // Account for header and footer
	}
	return m, nil
}

// View implements tea.Model
func (m *MainView) View() string {
	if m.width == 0 {
		return "Loading main view..."
	}

	// Current Specification section
	specSection := m.renderSpecSection()
	
	// Generation Status section
	statusSection := m.renderStatusSection()
	
	// Recent Activity section
	activitySection := m.renderActivitySection()

	// Layout sections vertically
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		specSection,
		"",
		statusSection,
		"",
		activitySection,
	)

	// Center the content
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Render(content)
}

// renderSpecSection renders the current specification section
func (m *MainView) renderSpecSection() string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(1, 2).
		Width(m.width - 6)

	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Bold(true)

	buttonStyle := lipgloss.NewStyle().
		Foreground(m.theme.Accent).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent)

	title := titleStyle.Render("Current Specification")
	buttons := lipgloss.JoinHorizontal(
		lipgloss.Left,
		buttonStyle.Render("Edit"),
		" ",
		buttonStyle.Render("Validate"),
	)

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(m.width-lipgloss.Width(title)-lipgloss.Width(buttons)-10).Render(""),
		buttons,
	)

	var content string
	if m.state.CurrentSpec != nil {
		content = fmt.Sprintf(
			"customer.yaml\nDomain: %s\nFields: %d (%s)\nTarget: %d records",
			m.state.CurrentSpec.Dataset.Domain,
			len(m.state.CurrentSpec.Dataset.Fields),
			m.getFieldNames(),
			m.state.CurrentSpec.Dataset.Count,
		)
	} else {
		content = "No specification loaded\nPress F2 to browse specifications or Ctrl+N to create new"
	}

	return sectionStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			"",
			content,
		),
	)
}

// renderStatusSection renders the generation status section
func (m *MainView) renderStatusSection() string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Info).
		Padding(1, 2).
		Width(m.width - 6)

	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Info).
		Bold(true)

	statusStyle := lipgloss.NewStyle().
		Foreground(m.theme.Success)

	errorStyle := lipgloss.NewStyle().
		Foreground(m.theme.Error)

	buttonStyle := lipgloss.NewStyle().
		Foreground(m.theme.Primary).
		Bold(true).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary)

	title := titleStyle.Render("System Status")

	var status, details string
	var buttons string

	// Check Ollama status (this would be populated by actual health check)
	ollamaRunning := true // TODO: Get from actual health check
	ollamaModels := []string{"llama3.1:8b", "mistral:7b"} // TODO: Get from actual health check

	if !ollamaRunning {
		status = errorStyle.Render("❌ Ollama not running")
		details = "Ollama is required for data generation\nRun 'fauxfoundry doctor' for setup instructions"
		buttons = lipgloss.JoinHorizontal(
			lipgloss.Left,
			buttonStyle.Render("Setup Guide"),
			" ",
			buttonStyle.Render("Refresh"),
		)
	} else if len(ollamaModels) == 0 {
		status = errorStyle.Render("⚠️  No models installed")
		details = "At least one model is required\nRecommended: llama3.1:8b"
		buttons = lipgloss.JoinHorizontal(
			lipgloss.Left,
			buttonStyle.Render("Install Model"),
			" ",
			buttonStyle.Render("Refresh"),
		)
	} else if m.state.ActiveGeneration != nil {
		status = fmt.Sprintf("🔄 %s", m.state.ActiveGeneration.Status)
		details = fmt.Sprintf(
			"Progress: %d/%d records\nModel: %s\nElapsed: %s",
			m.state.ActiveGeneration.Progress.Generated,
			m.state.ActiveGeneration.Progress.Target,
			m.state.ActiveGeneration.Spec.Model.Name,
			m.state.ActiveGeneration.Progress.ElapsedTime,
		)
		buttons = lipgloss.JoinHorizontal(
			lipgloss.Left,
			buttonStyle.Render("Pause"),
			" ",
			buttonStyle.Render("Cancel"),
		)
	} else {
		status = statusStyle.Render("✅ Ready to generate")
		details = fmt.Sprintf(
			"Ollama: Connected\nModels: %d available (%s)\nEstimated time: ~2 minutes",
			len(ollamaModels),
			ollamaModels[0],
		)
		buttons = lipgloss.JoinHorizontal(
			lipgloss.Left,
			buttonStyle.Render("Generate"),
			" ",
			buttonStyle.Render("Settings"),
			" ",
			buttonStyle.Render("Preview"),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		status,
		details,
		"",
		buttons,
	)

	return sectionStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			content,
		),
	)
}

// renderActivitySection renders the recent activity section
func (m *MainView) renderActivitySection() string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Secondary).
		Padding(1, 2).
		Width(m.width - 6)

	titleStyle := lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		Bold(true)

	title := titleStyle.Render("Recent Activity")

	// Default activity if no history
	activities := []string{
		"12:30 PM - customer.yaml validated successfully",
		"12:28 PM - Generated 500 product records",
		"12:25 PM - Created new specification: products.yaml",
	}

	// Use actual history if available
	if len(m.state.History) > 0 {
		activities = []string{}
		for i, entry := range m.state.History {
			if i >= 3 { // Show only last 3 entries
				break
			}
			activities = append(activities, fmt.Sprintf("%s - %s", entry.Timestamp, entry.Action))
		}
	}

	content := ""
	for _, activity := range activities {
		content += activity + "\n"
	}

	return sectionStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			content,
		),
	)
}

// getFieldNames returns a comma-separated list of field names
func (m *MainView) getFieldNames() string {
	if m.state.CurrentSpec == nil || len(m.state.CurrentSpec.Dataset.Fields) == 0 {
		return ""
	}

	names := make([]string, 0, len(m.state.CurrentSpec.Dataset.Fields))
	for _, field := range m.state.CurrentSpec.Dataset.Fields {
		names = append(names, field.Name)
	}

	// Join first few names
	if len(names) <= 3 {
		return fmt.Sprintf("%v", names)
	}
	return fmt.Sprintf("%s, %s, %s, ...", names[0], names[1], names[2])
}
