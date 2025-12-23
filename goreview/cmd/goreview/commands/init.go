package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize goreview configuration",
	Long: `Initialize goreview configuration in your project.

This command creates a .goreview.yaml configuration file with
sensible defaults based on your project structure.

Examples:
  # Interactive initialization
  goreview init

  # Non-interactive with defaults
  goreview init --yes

  # Specify provider
  goreview init --provider ollama --model codellama`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Mode flags
	initCmd.Flags().BoolP("yes", "y", false, "Accept all defaults (non-interactive)")
	initCmd.Flags().Bool("force", false, "Overwrite existing configuration")

	// Provider flags
	initCmd.Flags().String("provider", "", "AI provider (ollama, openai)")
	initCmd.Flags().String("model", "", "Model to use")
	initCmd.Flags().String("api-key", "", "API key for provider")

	// Project flags
	initCmd.Flags().String("preset", "standard", "Rule preset (minimal, standard, strict)")
	initCmd.Flags().StringSlice("exclude", nil, "Patterns to exclude")
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Init command - implementation follows")
	return nil
}
