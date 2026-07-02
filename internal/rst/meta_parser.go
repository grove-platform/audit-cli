// Package rst provides utilities for parsing reStructuredText documentation files.
package rst

import (
	"bufio"
	"os"
	"strings"
)

// ExtractMetaDescription reads the value of the :description: field from the
// first ".. meta::" directive in an RST file.
//
// The meta directive looks like:
//
//	.. meta::
//	   :robots: noindex, nosnippet
//	   :description: A short summary of the page.
//
// The description value may wrap across multiple indented continuation lines,
// which are joined with single spaces.
//
// Parameters:
//   - filePath: Path to the source .txt file
//
// Returns:
//   - string: The description text, or an empty string if there is no meta
//     directive or no :description: field
//   - error: Error if the file cannot be read
func ExtractMetaDescription(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Allow long lines (default token size can be too small for long descriptions).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	inMeta := false
	collecting := false
	var parts []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !inMeta {
			if strings.HasPrefix(trimmed, ".. meta::") {
				inMeta = true
			}
			continue
		}

		// Inside the meta directive.
		indented := line != "" && (line[0] == ' ' || line[0] == '\t')

		// A blank line does not end the directive on its own, but it does end a
		// multi-line description value.
		if trimmed == "" {
			if collecting {
				break
			}
			continue
		}

		// A non-indented, non-blank line ends the meta directive block.
		if !indented {
			break
		}

		if collecting {
			// Continuation lines are indented more deeply and are not new options.
			if strings.HasPrefix(trimmed, ":") {
				break
			}
			parts = append(parts, trimmed)
			continue
		}

		// Look for the :description: option.
		if strings.HasPrefix(trimmed, ":description:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, ":description:"))
			parts = append(parts, value)
			collecting = true
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.Join(parts, " ")), nil
}
