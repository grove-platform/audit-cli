// Package report provides the parent command for generating reports.
//
// This package serves as the parent command for various reporting operations.
// Currently supports:
//   - testable-code: Analyze testable code examples on pages from analytics data
//
// Future subcommands could include other report types for documentation metrics.
package report

import (
	testablecode "github.com/grove-platform/audit-cli/commands/report/testable-code"
	"github.com/spf13/cobra"
)

// NewReportCommand creates the report parent command.
//
// This command serves as a parent for various reporting operations.
// It doesn't perform any operations itself but provides a namespace for subcommands.
func NewReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate reports from documentation data",
		Long: `Generate various reports from documentation data and analytics.

Currently supports:
  - testable-code: Analyze testable code examples on pages from analytics CSV data

Future subcommands may support other report types for documentation metrics.`,
	}

	// Add subcommands
	cmd.AddCommand(testablecode.NewTestableCodeCommand())

	return cmd
}

