package rst

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLCodeBlock represents a code example in YAML-native format.
// This is the legacy format used in some steps files with action: blocks.
const YAMLCodeBlock DirectiveType = "yaml-code-block"

// YAMLActionItem represents an action item in a YAML steps file.
// This is the structure used in legacy steps files for code examples.
type YAMLActionItem struct {
	Pre      string `yaml:"pre"`
	Language string `yaml:"language"`
	Code     string `yaml:"code"`
	Copyable bool   `yaml:"copyable"`
}

// ParseYAMLStepsFile parses a YAML steps file and extracts code examples.
//
// This function handles the legacy YAML-native format where code examples
// are defined using action: blocks with language: and code: fields, rather
// than RST directives like .. code-block::.
//
// Example YAML format:
//
//	title: Download the file
//	stepnum: 1
//	action:
//	  - pre: "Run this command:"
//	    language: sh
//	    code: |
//	      curl -LO https://example.com/file.tgz
//
// Parameters:
//   - filePath: Path to the YAML steps file
//
// Returns:
//   - []Directive: Slice of directives representing code examples
//   - error: Any error encountered during parsing
func ParseYAMLStepsFile(filePath string) ([]Directive, error) {
	// Only process YAML files
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".yaml" && ext != ".yml" {
		return nil, nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var directives []Directive

	// Split by YAML document separator and parse each document
	documents := strings.Split(string(content), "\n---")
	lineNum := 1

	for _, doc := range documents {
		if strings.TrimSpace(doc) == "" || strings.TrimSpace(doc) == "..." {
			// Count lines in empty/end documents
			lineNum += strings.Count(doc, "\n")
			continue
		}

		var step YAMLStep
		if err := yaml.Unmarshal([]byte(doc), &step); err != nil {
			// Skip documents that don't parse as steps
			lineNum += strings.Count(doc, "\n")
			continue
		}

		// Extract code examples from action blocks
		// Action can be a single item or a list of items
		actions := extractActionsFromStep(step)
		for _, action := range actions {
			if action.Code != "" && action.Language != "" {
				directive := Directive{
					Type:     YAMLCodeBlock,
					Argument: action.Language, // Language goes in Argument like code-block
					Options:  make(map[string]string),
					Content:  strings.TrimSpace(action.Code),
					LineNum:  lineNum,
				}
				// Store language in options too for consistency
				directive.Options["language"] = action.Language
				directives = append(directives, directive)
			}
		}

		lineNum += strings.Count(doc, "\n") + 1 // +1 for the --- separator
	}

	return directives, nil
}

// extractActionsFromStep extracts action items from a YAMLStep.
// The Action field can be either a single map or a list of maps.
func extractActionsFromStep(step YAMLStep) []YAMLActionItem {
	if step.Action == nil {
		return nil
	}

	var actions []YAMLActionItem

	// Try as a list of maps first (most common)
	if actionList, ok := step.Action.([]interface{}); ok {
		for _, item := range actionList {
			if actionMap, ok := item.(map[string]interface{}); ok {
				action := parseActionMap(actionMap)
				actions = append(actions, action)
			}
		}
		return actions
	}

	// Try as a single map
	if actionMap, ok := step.Action.(map[string]interface{}); ok {
		action := parseActionMap(actionMap)
		actions = append(actions, action)
	}

	return actions
}

// parseActionMap converts a map[string]interface{} to a YAMLActionItem.
func parseActionMap(m map[string]interface{}) YAMLActionItem {
	var action YAMLActionItem

	if pre, ok := m["pre"].(string); ok {
		action.Pre = pre
	}
	if lang, ok := m["language"].(string); ok {
		action.Language = lang
	}
	if code, ok := m["code"].(string); ok {
		action.Code = code
	}
	if copyable, ok := m["copyable"].(bool); ok {
		action.Copyable = copyable
	}

	return action
}

