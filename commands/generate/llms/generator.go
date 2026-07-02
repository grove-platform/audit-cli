// Package llms provides generation of per-project llms.txt files.
package llms

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	resolveurl "github.com/grove-platform/audit-cli/commands/resolve/url"
	"github.com/grove-platform/audit-cli/internal/projectinfo"
	"github.com/grove-platform/audit-cli/internal/rst"
	"github.com/grove-platform/audit-cli/internal/snooty"
)

// CharLimit is the maximum recommended size (in characters) for an llms.txt file.
const CharLimit = 50000

// defaultExclusions are content-directory children that are not real docs
// projects and should never produce an llms.txt file.
var defaultExclusions = map[string]bool{
	"404":               true,
	"docs-platform":     true,
	"meta":              true,
	"table-of-contents": true,
	"code-examples":     true,
	// Deprecated projects: no useful content for agents.
	"app-services": true,
	"realm":        true,
}

// PageEntry holds the data needed to render one llms.txt line.
type PageEntry struct {
	Title       string
	URL         string // production URL with .md appended
	Description string // meta description ("" if none)
	SourcePath  string
}

// ProjectResult is the outcome of generating one project's llms.txt.
type ProjectResult struct {
	Project     string
	Version     string // "" for non-versioned projects
	Pages       []PageEntry
	OutputPath  string
	CharsWith   int // character count including descriptions
	CharsNoDesc int // character count omitting descriptions
	MissingDesc int // number of pages lacking a meta description
}

// Options configures a generation run.
type Options struct {
	MonorepoPath   string
	BaseURL        string
	OutputDir      string
	ForProject     string // limit to a single content-dir name; "" for all
	NoDescriptions bool   // omit descriptions from the written files
}

// Generate builds llms.txt files for the current + non-versioned pages of each
// documentation project and writes them under opts.OutputDir. It returns one
// ProjectResult per project processed.
func Generate(opts Options) ([]*ProjectResult, error) {
	contentDir := filepath.Join(opts.MonorepoPath, "content")
	if _, err := os.Stat(contentDir); err != nil {
		return nil, fmt.Errorf("content directory not found: %s", contentDir)
	}

	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read content directory: %w", err)
	}

	var results []*ProjectResult
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		project := entry.Name()
		if defaultExclusions[project] {
			continue
		}
		if opts.ForProject != "" && project != opts.ForProject {
			continue
		}

		projectDir := filepath.Join(contentDir, project)
		sourceDir, version, err := currentSourceDir(projectDir)
		if err != nil {
			return nil, fmt.Errorf("project %s: %w", project, err)
		}
		if sourceDir == "" {
			// No resolvable current source directory; skip.
			continue
		}

		result, err := generateProject(project, version, sourceDir, opts)
		if err != nil {
			return nil, fmt.Errorf("project %s: %w", project, err)
		}
		if result != nil {
			results = append(results, result)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Project < results[j].Project
	})
	return results, nil
}

// currentSourceDir returns the source directory to use for a project along with
// its version label. Non-versioned projects (content/<project>/source) return
// an empty version. Versioned projects return the current version's source dir.
func currentSourceDir(projectDir string) (sourceDir string, version string, err error) {
	// Non-versioned project.
	directSource := filepath.Join(projectDir, "source")
	if info, statErr := os.Stat(directSource); statErr == nil && info.IsDir() {
		return directSource, "", nil
	}

	// Versioned project: pick the current version.
	versions, err := projectinfo.DiscoverAllVersions(projectDir)
	if err != nil || len(versions) == 0 {
		return "", "", nil
	}
	for _, v := range versions {
		if projectinfo.IsCurrentVersion(v) {
			candidate := filepath.Join(projectDir, v, "source")
			if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
				return candidate, v, nil
			}
		}
	}
	return "", "", nil
}

// generateProject collects pages for a single project and writes its llms.txt.
func generateProject(project, version, sourceDir string, opts Options) (*ProjectResult, error) {
	var pages []PageEntry

	// Load the project's snooty constants once so {+name+} substitutions in
	// page titles can be resolved. Absence of constants is not fatal.
	constants := loadConstants(sourceDir)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip partial/include and code-example directories: these are not
			// standalone pages and don't have their own production URLs.
			name := info.Name()
			if name == "includes" || name == "code-examples" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".txt" {
			return nil
		}

		url, err := resolveurl.ResolveFileToURL(path, opts.BaseURL)
		if err != nil {
			// A page we can't map to a URL isn't useful in llms.txt; skip it.
			return nil
		}

		title, err := rst.ExtractPageTitle(path)
		if err != nil {
			return err
		}
		if title == "" {
			// Without a title there's nothing meaningful to link; skip.
			return nil
		}
		title = snooty.ResolveSubstitutions(title, constants)

		description, err := rst.ExtractMetaDescription(path)
		if err != nil {
			return err
		}
		description = snooty.ResolveSubstitutions(description, constants)

		// The project's root landing page has no "<root>.md" markdown form; its
		// markdown lives at "<root>/index.md" instead. Nested section index
		// pages already resolve to a normal "<section>.md" URL.
		isRootIndex := path == filepath.Join(sourceDir, "index.txt")

		pages = append(pages, PageEntry{
			Title:       title,
			URL:         toMarkdownURL(url, isRootIndex),
			Description: description,
			SourcePath:  path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, nil
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].URL < pages[j].URL
	})

	result := &ProjectResult{
		Project:    project,
		Version:    version,
		Pages:      pages,
		CharsWith:  utf8.RuneCountInString(renderContent(project, pages, true)),
		CharsNoDesc: utf8.RuneCountInString(renderContent(project, pages, false)),
	}
	for _, p := range pages {
		if p.Description == "" {
			result.MissingDesc++
		}
	}

	// Write the file.
	content := renderContent(project, pages, !opts.NoDescriptions)
	outPath := filepath.Join(opts.OutputDir, project, "llms.txt")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
		return nil, err
	}
	result.OutputPath = outPath

	return result, nil
}

// loadConstants reads the substitution constants from the project's snooty.toml,
// which sits in the directory containing the source directory. Returns nil if
// the file is missing or cannot be parsed.
func loadConstants(sourceDir string) map[string]string {
	snootyPath := filepath.Join(filepath.Dir(sourceDir), "snooty.toml")
	if _, err := os.Stat(snootyPath); err != nil {
		return nil
	}
	config, err := snooty.ParseFile(snootyPath)
	if err != nil {
		return nil
	}
	return config.Constants
}

// renderContent builds the llms.txt body. When withDesc is true, descriptions
// are appended as ": <description>" when present.
func renderContent(project string, pages []PageEntry, withDesc bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", project))
	for _, p := range pages {
		if withDesc && p.Description != "" {
			b.WriteString(fmt.Sprintf("- [%s](%s): %s\n", p.Title, p.URL, p.Description))
		} else {
			b.WriteString(fmt.Sprintf("- [%s](%s)\n", p.Title, p.URL))
		}
	}
	return b.String()
}

// toMarkdownURL converts a production page URL to its Markdown (.md) form.
//
// Regular pages and nested section indexes use the "<page>.md" form (the URL
// with its trailing slash replaced by ".md"). The project's root landing page
// has no "<root>.md" form and instead uses "<root>/index.md".
func toMarkdownURL(url string, isRootIndex bool) string {
	if isRootIndex {
		return strings.TrimSuffix(url, "/") + "/index.md"
	}
	return strings.TrimSuffix(url, "/") + ".md"
}
