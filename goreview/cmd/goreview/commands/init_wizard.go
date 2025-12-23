package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// InitWizard handles interactive initialization.
type InitWizard struct {
	reader *bufio.Reader
	info   *ProjectInfo
}

// NewInitWizard creates a new initialization wizard.
func NewInitWizard(info *ProjectInfo) *InitWizard {
	return &InitWizard{
		reader: bufio.NewReader(os.Stdin),
		info:   info,
	}
}

// Run executes the interactive wizard.
func (w *InitWizard) Run() (map[string]interface{}, error) {
	config := make(map[string]interface{})

	fmt.Println("\n┌─────────────────────────────────────┐")
	fmt.Println("│     GoReview Configuration Wizard   │")
	fmt.Println("└─────────────────────────────────────┘\n")

	// Show detected info
	w.showDetectedInfo()

	// Provider selection
	provider, err := w.selectProvider()
	if err != nil {
		return nil, err
	}
	config["provider"] = provider

	// Model selection
	model, err := w.selectModel(provider)
	if err != nil {
		return nil, err
	}
	config["model"] = model

	// API Key (if needed)
	if provider == "openai" {
		apiKey, err := w.promptAPIKey()
		if err != nil {
			return nil, err
		}
		config["api_key"] = apiKey
	}

	// Preset selection
	preset, err := w.selectPreset()
	if err != nil {
		return nil, err
	}
	config["preset"] = preset

	// Exclude patterns
	excludes := w.info.SuggestDefaults()["exclude"].([]string)
	config["exclude"] = excludes

	// Confirmation
	w.showSummary(config)
	if !w.confirm("Create configuration?") {
		return nil, fmt.Errorf("initialization cancelled")
	}

	return config, nil
}

func (w *InitWizard) showDetectedInfo() {
	fmt.Println("Detected project information:")
	fmt.Println("─────────────────────────────")

	if len(w.info.Languages) > 0 {
		fmt.Printf("  Languages:    %s\n", strings.Join(w.info.Languages, ", "))
	}
	if w.info.ProjectType != "" {
		fmt.Printf("  Project type: %s\n", w.info.ProjectType)
	}
	if len(w.info.Frameworks) > 0 {
		fmt.Printf("  Frameworks:   %s\n", strings.Join(w.info.Frameworks, ", "))
	}
	fmt.Printf("  Git repo:     %v\n", w.info.HasGit)
	fmt.Printf("  CI detected:  %v\n", w.info.HasCI)
	fmt.Println()
}

func (w *InitWizard) selectProvider() (string, error) {
	fmt.Println("Select AI provider:")
	fmt.Println("  [1] Ollama (local, free)")
	fmt.Println("  [2] OpenAI (cloud, requires API key)")
	fmt.Print("\nChoice [1]: ")

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "", "1":
		return "ollama", nil
	case "2":
		return "openai", nil
	default:
		return "ollama", nil
	}
}

func (w *InitWizard) selectModel(provider string) (string, error) {
	var options []string
	var defaultModel string

	switch provider {
	case "ollama":
		options = []string{"qwen2.5-coder:14b", "codellama", "deepseek-coder", "mistral"}
		defaultModel = "qwen2.5-coder:14b"
	case "openai":
		options = []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}
		defaultModel = "gpt-4"
	}

	fmt.Println("\nSelect model:")
	for i, opt := range options {
		def := ""
		if opt == defaultModel {
			def = " (recommended)"
		}
		fmt.Printf("  [%d] %s%s\n", i+1, opt, def)
	}
	fmt.Printf("\nChoice [1]: ")

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultModel, nil
	}

	idx := 0
	fmt.Sscanf(input, "%d", &idx)
	if idx > 0 && idx <= len(options) {
		return options[idx-1], nil
	}

	return defaultModel, nil
}

func (w *InitWizard) promptAPIKey() (string, error) {
	fmt.Print("\nEnter OpenAI API key: ")
	input, _ := w.reader.ReadString('\n')
	return strings.TrimSpace(input), nil
}

func (w *InitWizard) selectPreset() (string, error) {
	fmt.Println("\nSelect rule preset:")
	fmt.Println("  [1] minimal  - Only critical security rules")
	fmt.Println("  [2] standard - Recommended for most projects")
	fmt.Println("  [3] strict   - Maximum code quality checks")
	fmt.Print("\nChoice [2]: ")

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		return "minimal", nil
	case "", "2":
		return "standard", nil
	case "3":
		return "strict", nil
	default:
		return "standard", nil
	}
}

func (w *InitWizard) showSummary(config map[string]interface{}) {
	fmt.Println("\n┌─────────────────────────────────────┐")
	fmt.Println("│         Configuration Summary       │")
	fmt.Println("├─────────────────────────────────────┤")
	fmt.Printf("│  Provider: %-24s │\n", config["provider"])
	fmt.Printf("│  Model:    %-24s │\n", config["model"])
	fmt.Printf("│  Preset:   %-24s │\n", config["preset"])
	fmt.Println("└─────────────────────────────────────┘")
}

func (w *InitWizard) confirm(message string) bool {
	fmt.Printf("\n%s [Y/n]: ", message)
	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}
