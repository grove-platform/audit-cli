// Package generate provides the parent command for generating documentation artifacts.
//
// This package serves as the parent command for generation operations.
// Currently supports:
//   - llms: Generate per-project llms.txt files
package generate

import (
	"github.com/grove-platform/audit-cli/commands/generate/llms"
	"github.com/spf13/cobra"
)

// NewGenerateCommand creates the generate parent command.
//
// This command serves as a parent for various generation operations.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate documentation artifacts",
		Long: `Generate artifacts derived from the documentation monorepo.

Currently supports:
  - llms: Generate per-project llms.txt files for progressive disclosure`,
	}

	cmd.AddCommand(llms.NewLLMSCommand())

	return cmd
}
