package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateReviewFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]interface{}
		args    []string
		wantErr bool
	}{
		{
			name:    "no mode specified",
			flags:   map[string]interface{}{},
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "staged mode",
			flags:   map[string]interface{}{"staged": true},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "commit mode",
			flags:   map[string]interface{}{"commit": "abc123"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "files mode",
			flags:   map[string]interface{}{},
			args:    []string{"file.go"},
			wantErr: false,
		},
		{
			name:    "multiple modes",
			flags:   map[string]interface{}{"staged": true, "commit": "abc123"},
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid format",
			flags:   map[string]interface{}{"staged": true, "format": "xml"},
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("staged", false, "")
			cmd.Flags().String("commit", "", "")
			cmd.Flags().String("branch", "", "")
			cmd.Flags().String("format", "markdown", "")

			for k, v := range tt.flags {
				switch val := v.(type) {
				case bool:
					cmd.Flags().Set(k, "true")
				case string:
					cmd.Flags().Set(k, val)
				}
			}

			err := validateReviewFlags(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateReviewFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetermineReviewMode(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]interface{}
		args     []string
		wantMode string
	}{
		{
			name:     "staged",
			flags:    map[string]interface{}{"staged": true},
			wantMode: "staged",
		},
		{
			name:     "commit",
			flags:    map[string]interface{}{"commit": "abc123"},
			wantMode: "commit",
		},
		{
			name:     "branch",
			flags:    map[string]interface{}{"branch": "main"},
			wantMode: "branch",
		},
		{
			name:     "files",
			args:     []string{"file1.go", "file2.go"},
			wantMode: "files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("staged", false, "")
			cmd.Flags().String("commit", "", "")
			cmd.Flags().String("branch", "", "")

			for k, v := range tt.flags {
				switch val := v.(type) {
				case bool:
					cmd.Flags().Set(k, "true")
				case string:
					cmd.Flags().Set(k, val)
				}
			}

			mode, _ := determineReviewMode(cmd, tt.args)
			if mode != tt.wantMode {
				t.Errorf("determineReviewMode() = %v, want %v", mode, tt.wantMode)
			}
		})
	}
}

func TestDetectFormatFromPath(t *testing.T) {
	tests := []struct {
		path   string
		format string
	}{
		{"report.json", "json"},
		{"report.sarif", "sarif"},
		{"report.md", "markdown"},
		{"report.markdown", "markdown"},
		{"report.txt", ""},
		{"report", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := DetectFormatFromPath(tt.path)
			if got != tt.format {
				t.Errorf("DetectFormatFromPath(%q) = %q, want %q", tt.path, got, tt.format)
			}
		})
	}
}
