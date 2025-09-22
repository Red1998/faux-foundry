package cli

import (
	"github.com/spf13/cobra"
	"github.com/copyleftdev/faux-foundry/internal/tui"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal interface",
	Long: `Launch the interactive Terminal User Interface (TUI) for FauxFoundry. The TUI provides
a rich, keyboard-driven interface for creating specifications, monitoring generation progress,
and managing your synthetic data workflows.

Features:
  - Interactive specification editor with validation
  - Real-time generation monitoring with progress bars
  - Specification browser and file management
  - Settings and configuration management
  - Contextual help and keyboard shortcuts

Examples:
  # Launch TUI
  fauxfoundry tui

  # Launch TUI with specific specification
  fauxfoundry tui --spec customer.yaml`,
	RunE: runTUI,
}

func init() {
	tuiCmd.Flags().StringVarP(&specFile, "spec", "s", "", "load specific specification file")
}

func runTUI(cmd *cobra.Command, args []string) error {
	return tui.Run(specFile)
}
