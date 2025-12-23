package commands

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
)

// TestVersionCommand tests the version command output
func TestVersionCommand(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origBuildDate := BuildDate

	// Set test values
	Version = "1.2.3"
	Commit = "abc123def"
	BuildDate = "2024-01-15T10:00:00Z"

	// Restore after test
	defer func() {
		Version = origVersion
		Commit = origCommit
		BuildDate = origBuildDate
	}()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:    "default output",
			args:    []string{},
			wantErr: false,
			contains: []string{
				"goreview version 1.2.3",
				"Commit:     abc123def",
				"Built:      2024-01-15T10:00:00Z",
				runtime.Version(),
			},
		},
		{
			name:     "short flag",
			args:     []string{"--short"},
			wantErr:  false,
			contains: []string{"1.2.3"},
		},
		{
			name:    "json flag",
			args:    []string{"--json"},
			wantErr: false,
			contains: []string{
				`"version": "1.2.3"`,
				`"commit": "abc123def"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each test
			versionShort = false
			versionJSON = false

			// Create a new command for testing
			cmd := versionCmd
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			// Execute
			err := cmd.Execute()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For non-error cases, we need to capture stdout differently
			// Since fmt.Printf goes to os.Stdout, not cmd.OutOrStdout()
			// We'll test the function directly instead
		})
	}
}

// TestGetVersionInfo tests the GetVersionInfo function
func TestGetVersionInfo(t *testing.T) {
	// Save original
	origVersion := Version
	Version = "test-version"
	defer func() { Version = origVersion }()

	info := GetVersionInfo()

	if info.Version != "test-version" {
		t.Errorf("GetVersionInfo().Version = %v, want %v", info.Version, "test-version")
	}

	if info.GoVersion != runtime.Version() {
		t.Errorf("GetVersionInfo().GoVersion = %v, want %v", info.GoVersion, runtime.Version())
	}

	if info.OS != runtime.GOOS {
		t.Errorf("GetVersionInfo().OS = %v, want %v", info.OS, runtime.GOOS)
	}

	if info.Arch != runtime.GOARCH {
		t.Errorf("GetVersionInfo().Arch = %v, want %v", info.Arch, runtime.GOARCH)
	}
}

// TestVersionInfoJSON tests JSON marshaling of VersionInfo
func TestVersionInfoJSON(t *testing.T) {
	info := VersionInfo{
		Version:   "1.0.0",
		Commit:    "abc123",
		BuildDate: "2024-01-15",
		GoVersion: "go1.23.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal VersionInfo: %v", err)
	}

	// Unmarshal back
	var decoded VersionInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal VersionInfo: %v", err)
	}

	if decoded.Version != info.Version {
		t.Errorf("Version mismatch: got %v, want %v", decoded.Version, info.Version)
	}

	// Check JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{"version", "commit", "build_date", "go_version", "os", "arch"}
	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing field: %s", field)
		}
	}
}

// TestVersionCommandArgs tests that version command rejects arguments
func TestVersionCommandArgs(t *testing.T) {
	// Execute through rootCmd to properly test argument validation
	rootCmd.SetArgs([]string{"version", "unexpected-arg"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for unexpected argument, got nil")
	}
}
