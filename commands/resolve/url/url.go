// Package url implements the url subcommand for resolving source files to production URLs.
package url

import (
	"fmt"

	"github.com/grove-platform/audit-cli/internal/config"
	"github.com/spf13/cobra"
)

// NewURLCommand creates the url subcommand.
//
// This command resolves source .txt files to their production URLs.
//
// Usage:
//
//	resolve url <source-file>
//	resolve url content/manual/manual/source/tutorial/install.txt
//
// Flags:
//   - --base-url: Override the base URL for resolution
func NewURLCommand() *cobra.Command {
	var baseURL string

	cmd := &cobra.Command{
		Use:   "url <source-file>",
		Short: "Resolve a source .txt file to its production URL",
		Long: `Resolve a source .txt file from the documentation monorepo to its production URL.

This command takes a path to a .txt source file and outputs the corresponding
production URL on mongodb.com.

How It Works:
  1. Identifies the project from the file path (e.g., content/manual/...)
  2. Extracts the page path relative to the source directory
  3. Constructs the production URL based on the project's URL slug

File Path Resolution:
  Paths can be specified as:
    1. Absolute path: /full/path/to/file.txt
    2. Relative to monorepo root (if configured): content/manual/manual/source/index.txt
    3. Relative to current directory: ./source/index.txt

Examples:
  # Resolve a file to its production URL
  resolve url content/manual/manual/source/tutorial/install.txt
  # Output: https://www.mongodb.com/docs/manual/tutorial/install/

  # Resolve an index file
  resolve url content/atlas/source/index.txt
  # Output: https://www.mongodb.com/docs/atlas/

  # Override the base URL
  resolve url content/manual/manual/source/reference/method.txt --base-url https://docs-staging.mongodb.com`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve file path (supports absolute, monorepo-relative, or cwd-relative)
			filePath, err := config.ResolveFilePath(args[0])
			if err != nil {
				return err
			}
			return runResolveURL(filePath, baseURL)
		},
	}

	cmd.Flags().StringVar(&baseURL, "base-url", "https://www.mongodb.com/docs", "Base URL for production documentation")

	return cmd
}

// runResolveURL executes the URL resolution operation.
func runResolveURL(filePath string, baseURL string) error {
	// Resolve the file to a URL
	productionURL, err := ResolveFileToURL(filePath, baseURL)
	if err != nil {
		return fmt.Errorf("failed to resolve URL: %w", err)
	}

	// Print the result
	fmt.Println(productionURL)

	return nil
}

