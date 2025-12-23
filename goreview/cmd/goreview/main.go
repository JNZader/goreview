// Package main is the entry point for the goreview CLI.
//
// This file is intentionally minimal - all logic lives in the commands package.
// The main function only initializes and executes the root command.
package main

import (
	"os"

	"github.com/JNZader/goreview/goreview/cmd/goreview/commands"
)

func main() {
	// Execute the root command
	// If there's an error, Cobra will print it and we exit with code 1
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
