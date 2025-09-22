package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
	noColor bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fauxfoundry",
	Short: "A CLI and TUI for synthetic, domain-aware data generation powered by local LLMs",
	Long: `FauxFoundry enables teams to generate unique synthetic datasets from human-readable 
YAML specifications. It leverages local AI models (e.g., Ollama) to produce realistic, 
domain-aware data that respects schema constraints while ensuring exactly N unique records 
are delivered through efficient streaming with minimal validation overhead.

This tool is designed with system empathy: it maintains constant memory usage, degrades 
gracefully under errors, and provides both automation-friendly CLI commands and discoverable 
TUI workflows for different user needs.`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.fauxfoundry.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(tuiCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		// viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".fauxfoundry" (without extension).
		// viper.AddConfigPath(home)
		// viper.SetConfigType("yaml")
		// viper.SetConfigName(".fauxfoundry")
		_ = home // Suppress unused variable warning for now
	}

	// viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	//     fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	// }
}

// Helper functions for global flags
func IsVerbose() bool {
	return verbose
}

func IsQuiet() bool {
	return quiet
}

func IsColorDisabled() bool {
	return noColor
}

func GetConfigFile() string {
	return cfgFile
}
