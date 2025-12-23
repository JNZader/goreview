package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const configFileName = ".goreview.yaml"

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
	// Check for existing config
	if _, err := os.Stat(configFileName); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return fmt.Errorf("configuration file already exists. Use --force to overwrite")
		}
	}

	// Detect project
	cwd, _ := os.Getwd()
	info, err := DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	var config map[string]interface{}

	// Interactive or non-interactive mode
	yes, _ := cmd.Flags().GetBool("yes")
	if yes {
		config = buildConfigFromFlags(cmd, info)
	} else {
		wizard := NewInitWizard(info)
		config, err = wizard.Run()
		if err != nil {
			return err
		}
	}

	// Generate YAML
	yamlConfig := buildYAMLConfig(config, info)

	// Write configuration file
	data, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFileName, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", configFileName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the configuration file")

	if config["provider"] == "ollama" {
		fmt.Println("  2. Ensure Ollama is running: ollama serve")
		fmt.Printf("  3. Pull the model: ollama pull %s\n", config["model"])
	} else {
		fmt.Println("  2. Set OPENAI_API_KEY environment variable")
	}

	fmt.Println("\nRun 'goreview review --staged' to review staged changes")

	return nil
}

func buildConfigFromFlags(cmd *cobra.Command, info *ProjectInfo) map[string]interface{} {
	config := info.SuggestDefaults()

	if provider, _ := cmd.Flags().GetString("provider"); provider != "" {
		config["provider"] = provider
	}
	if model, _ := cmd.Flags().GetString("model"); model != "" {
		config["model"] = model
	}
	if preset, _ := cmd.Flags().GetString("preset"); preset != "" {
		config["preset"] = preset
	}
	if excludes, _ := cmd.Flags().GetStringSlice("exclude"); len(excludes) > 0 {
		config["exclude"] = excludes
	}

	return config
}

func buildYAMLConfig(config map[string]interface{}, info *ProjectInfo) map[string]interface{} {
	yamlConfig := map[string]interface{}{
		"version": "1.0",
		"provider": map[string]interface{}{
			"name":  config["provider"],
			"model": config["model"],
		},
		"review": map[string]interface{}{
			"max_concurrency": 5,
			"timeout":         "5m",
		},
		"git": map[string]interface{}{
			"base_branch":     "main",
			"ignore_patterns": config["exclude"],
		},
		"rules": map[string]interface{}{
			"preset": config["preset"],
		},
		"cache": map[string]interface{}{
			"enabled":     true,
			"ttl":         "24h",
			"max_entries": 100,
		},
	}

	// Add Ollama-specific config
	if config["provider"] == "ollama" {
		yamlConfig["provider"].(map[string]interface{})["base_url"] = "http://localhost:11434"
	}

	// Add API key placeholder for OpenAI
	if config["provider"] == "openai" {
		yamlConfig["provider"].(map[string]interface{})["api_key"] = "${OPENAI_API_KEY}"
	}

	return yamlConfig
}
