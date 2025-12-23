package commands

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information - these are set at build time via ldflags
// See Makefile for how these are injected
var (
	// Version is the semantic version (e.g., "1.0.0")
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is the date the binary was built
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long: `Print detailed version information about the goreview binary.

This includes the version number, git commit hash, build date,
and Go runtime information.

Examples:
  # Print full version info
  goreview version

  # Print only version number
  goreview version --short

  # Print version as JSON
  goreview version --json`,

	// No arguments expected
	Args: cobra.NoArgs,

	// RunE returns an error (better than Run which panics)
	RunE: runVersion,
}

// Flags for version command
var (
	versionShort bool
	versionJSON  bool
)

func init() {
	// Register version command under root
	rootCmd.AddCommand(versionCmd)

	// Local flags for this command only
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "print only version number")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "output as JSON")
}

// VersionInfo holds all version information
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// runVersion implements the version command logic
func runVersion(cmd *cobra.Command, args []string) error {
	info := VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	// Short output - just version number
	if versionShort {
		fmt.Println(info.Version)
		return nil
	}

	// JSON output
	if versionJSON {
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal version info: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Default: human-readable output
	fmt.Printf("goreview version %s\n", info.Version)
	fmt.Printf("  Commit:     %s\n", info.Commit)
	fmt.Printf("  Built:      %s\n", info.BuildDate)
	fmt.Printf("  Go version: %s\n", info.GoVersion)
	fmt.Printf("  OS/Arch:    %s/%s\n", info.OS, info.Arch)

	return nil
}

// GetVersionInfo returns the current version info (useful for other packages)
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}
