// Package resolve provides the parent command for resolving documentation paths and URLs.
//
// This package serves as the parent command for resolution operations.
// Currently supports:
//   - url: Resolve source .txt files to their production URLs
//
// These commands help writers and tools map between source files and live documentation.
package resolve

import (
	"github.com/grove-platform/audit-cli/commands/resolve/url"
	"github.com/spf13/cobra"
)

// NewResolveCommand creates the resolve parent command.
//
// This command serves as a parent for various resolution operations.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewResolveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Resolve documentation paths and URLs",
		Long: `Resolve mappings between source files and production URLs.

Helps writers and tools understand the relationship between source .txt files
in the documentation monorepo and their corresponding live URLs.

Currently supports:
  - url: Resolve source .txt files to their production URLs`,
	}

	// Add subcommands
	cmd.AddCommand(url.NewURLCommand())

	return cmd
}

