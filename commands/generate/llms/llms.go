package llms

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/grove-platform/audit-cli/internal/config"
	"github.com/spf13/cobra"
)

// NewLLMSCommand creates the "generate llms" subcommand.
//
// Usage:
//
//	generate llms [monorepo-path] [flags]
//
// It generates one llms.txt per documentation project (current + non-versioned
// pages), then prints a summary of each file's character count both with and
// without meta descriptions, flagging any that exceed the 50k character limit.
func NewLLMSCommand() *cobra.Command {
	var baseURL string
	var outputDir string
	var forProject string
	var noDescriptions bool

	cmd := &cobra.Command{
		Use:   "llms [monorepo-path]",
		Short: "Generate per-project llms.txt files",
		Long: `Generate an llms.txt file for each documentation project.

For every project under the monorepo's content/ directory, this command
enumerates the pages of its current version (and non-versioned projects),
extracts each page's title and meta description, resolves its production URL
(with .md appended), and writes a project llms.txt in the format:

  - [Page Title](https://www.mongodb.com/docs/manual/core/document.md): Description.

After generating the files it prints a summary showing each project's character
count both WITH and WITHOUT descriptions, flagging any file over 50,000
characters so you can decide whether descriptions fit for larger projects.

Pages without a meta description are emitted without the trailing ": description".

Examples:
  # Generate for all projects using the configured monorepo path
  generate llms

  # Generate for a single project
  generate llms --for-project atlas

  # Omit descriptions (useful for oversized projects)
  generate llms --for-project cloud-docs --no-descriptions`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdLineArg := ""
			if len(args) == 1 {
				cmdLineArg = args[0]
			}
			monorepoPath, err := config.GetMonorepoPath(cmdLineArg)
			if err != nil {
				return err
			}

			results, err := Generate(Options{
				MonorepoPath:   monorepoPath,
				BaseURL:        baseURL,
				OutputDir:      outputDir,
				ForProject:     forProject,
				NoDescriptions: noDescriptions,
			})
			if err != nil {
				return err
			}
			return printSummary(results, outputDir, noDescriptions)
		},
	}

	cmd.Flags().StringVar(&baseURL, "base-url", "https://www.mongodb.com/docs", "Base URL for production documentation")
	cmd.Flags().StringVar(&outputDir, "output-dir", "llms-output", "Directory to write per-project llms.txt files into")
	cmd.Flags().StringVar(&forProject, "for-project", "", "Limit generation to a single project (content directory name)")
	cmd.Flags().BoolVar(&noDescriptions, "no-descriptions", false, "Omit meta descriptions from the written files")

	return cmd
}

// printSummary writes the per-project character-count report to stdout.
func printSummary(results []*ProjectResult, outputDir string, noDescriptions bool) error {
	if len(results) == 0 {
		fmt.Println("No projects generated (no matching content found).")
		return nil
	}

	fmt.Printf("Generated %d llms.txt file(s) in %s/\n\n", len(results), outputDir)

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tVERSION\tPAGES\tNO_DESC\tCHARS(w/ desc)\tCHARS(no desc)\tOVER 50k?")

	var over []string
	for _, r := range results {
		version := r.Version
		if version == "" {
			version = "-"
		}
		flag := ""
		// Which count applies to the file we actually wrote?
		written := r.CharsWith
		if noDescriptions {
			written = r.CharsNoDesc
		}
		if written > CharLimit {
			flag = "YES"
			over = append(over, r.Project)
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\t%d\t%s\n",
			r.Project, version, len(r.Pages), r.MissingDesc, r.CharsWith, r.CharsNoDesc, flag)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if len(over) > 0 {
		fmt.Printf("\n%d project(s) exceed the %d-character limit for the written files: %v\n",
			len(over), CharLimit, over)
		if !noDescriptions {
			fmt.Println("Consider re-running these with --no-descriptions, or compare the CHARS(no desc) column above.")
		}
	} else {
		fmt.Printf("\nAll files are within the %d-character limit.\n", CharLimit)
	}

	return nil
}
