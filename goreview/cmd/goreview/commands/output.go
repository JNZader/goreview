package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteOutput writes the report to a file or stdout.
func WriteOutput(content, outputPath string) error {
	if outputPath == "" {
		fmt.Print(content)
		return nil
	}

	// Create parent directories
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Report written to: %s\n", outputPath)
	return nil
}

// DetectFormatFromPath infers the output format from file extension.
func DetectFormatFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "json"
	case ".sarif":
		return "sarif"
	case ".md", ".markdown":
		return "markdown"
	default:
		return ""
	}
}
