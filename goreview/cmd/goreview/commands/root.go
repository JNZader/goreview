// Package commands contains all CLI commands for goreview.
//
// This package uses the Cobra library for CLI management.
// Each command is defined in its own file and registered in init().
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// cfgFile holds the path to the config file (from --config flag)
	cfgFile string

	// verbose enables detailed output
	verbose bool

	// quiet suppresses all output except errors
	quiet bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goreview",
	Short: "AI-powered code review tool",
	Long: `GoReview is a CLI tool that uses AI to review your code changes.

It analyzes diffs, identifies potential issues, and provides actionable feedback
on bugs, security vulnerabilities, performance problems, and best practices.

Examples:
  # Review staged changes
  goreview review

  # Review a specific commit
  goreview review --commit HEAD~1

  # Review changes compared to a branch
  goreview review --base main

  # Generate a commit message
  goreview commit

  # Show current configuration
  goreview config show`,

	// SilenceUsage prevents printing usage on errors
	// We want clean error messages, not the full help text
	SilenceUsage: true,

	// SilenceErrors lets us handle errors ourselves
	SilenceErrors: true,

	// PersistentPreRunE runs before any command (including subcommands)
	// Use this for initialization that all commands need
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flags are available to this command and all subcommands
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .goreview.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all output except errors")

	// Bind flags to viper for config file support
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
}

// initializeConfig reads in config file and ENV variables if set.
func initializeConfig() error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory and home directory
		viper.SetConfigName(".goreview")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	// Read environment variables that match
	// GOREVIEW_PROVIDER_NAME -> provider.name
	viper.SetEnvPrefix("GOREVIEW")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is not an error - we have defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	if verbose && !quiet {
		if viper.ConfigFileUsed() != "" {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	}

	return nil
}

// isVerbose returns true if verbose mode is enabled
func isVerbose() bool {
	return verbose && !quiet
}

// isQuiet returns true if quiet mode is enabled
func isQuiet() bool {
	return quiet
}
