package rst

import (
	"os"
	"strings"
)

// ExtractPageTitle returns the page's H1 title from an RST file.
//
// It finds the first section heading, supporting both underline-only headings:
//
//	Page Title
//	==========
//
// and overline+underline headings:
//
//	==========
//	Page Title
//	==========
//
// Directive lines (starting with "..") and RST field/option lines (starting
// with ":") are not considered valid titles.
//
// Parameters:
//   - filePath: Path to the source .txt file
//
// Returns:
//   - string: The title text, or an empty string if no heading is found
//   - error: Error if the file cannot be read
func ExtractPageTitle(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if !isHeadingUnderline(strings.TrimSpace(line)) {
			continue
		}

		// The title is the immediately preceding non-empty text line.
		if i == 0 {
			continue
		}
		candidate := strings.TrimSpace(lines[i-1])
		if candidate == "" {
			continue
		}
		// Skip directives, field lists, and overline rows.
		if strings.HasPrefix(candidate, "..") || strings.HasPrefix(candidate, ":") {
			continue
		}
		if isHeadingUnderline(candidate) {
			continue
		}
		return candidate, nil
	}

	return "", nil
}
