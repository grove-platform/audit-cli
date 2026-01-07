package rst

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAMLStepsFile(t *testing.T) {
	// Create a temporary YAML steps file
	tempDir, err := os.MkdirTemp("", "yaml-steps-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	yamlContent := `title: Download the file
stepnum: 1
ref: download-file
action:
   - pre: |
       Run this command:
     language: sh
     copyable: true
     code: |
       curl -LO https://example.com/file.tgz
   - pre: |
       For Linux:
     language: bash
     code: |
       wget https://example.com/file.tgz
---
title: Verify the file
stepnum: 2
ref: verify-file
action:
 - pre: |
     Check the signature:
   language: sh
   code: |
     gpg --verify file.tgz.sig file.tgz
...
`

	testFile := filepath.Join(tempDir, "steps-test.yaml")
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse the file
	directives, err := ParseYAMLStepsFile(testFile)
	if err != nil {
		t.Fatalf("ParseYAMLStepsFile failed: %v", err)
	}

	// Should find 3 code examples
	if len(directives) != 3 {
		t.Errorf("Expected 3 directives, got %d", len(directives))
	}

	// Verify directive types and languages
	expectedLangs := []string{"sh", "bash", "sh"}
	for i, d := range directives {
		if d.Type != YAMLCodeBlock {
			t.Errorf("Directive %d: expected type %s, got %s", i, YAMLCodeBlock, d.Type)
		}
		if d.Argument != expectedLangs[i] {
			t.Errorf("Directive %d: expected language %s, got %s", i, expectedLangs[i], d.Argument)
		}
		if d.Content == "" {
			t.Errorf("Directive %d: expected non-empty content", i)
		}
	}
}

func TestParseYAMLStepsFile_NonYAMLFile(t *testing.T) {
	// Create a temporary RST file
	tempDir, err := os.MkdirTemp("", "yaml-steps-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.rst")
	if err := os.WriteFile(testFile, []byte(".. code-block:: python\n\n   print('hello')"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse the file - should return nil for non-YAML files
	directives, err := ParseYAMLStepsFile(testFile)
	if err != nil {
		t.Fatalf("ParseYAMLStepsFile failed: %v", err)
	}

	if directives != nil {
		t.Errorf("Expected nil directives for non-YAML file, got %d", len(directives))
	}
}

func TestParseDirectives_IncludesYAMLCodeBlocks(t *testing.T) {
	// Create a temporary YAML steps file
	tempDir, err := os.MkdirTemp("", "yaml-steps-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	yamlContent := `title: Run command
stepnum: 1
action:
  - language: python
    code: |
      print("hello")
`

	testFile := filepath.Join(tempDir, "steps-test.yaml")
	if err := os.WriteFile(testFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Use ParseDirectives (the main entry point)
	directives, err := ParseDirectives(testFile)
	if err != nil {
		t.Fatalf("ParseDirectives failed: %v", err)
	}

	// Should find 1 code example
	if len(directives) != 1 {
		t.Errorf("Expected 1 directive, got %d", len(directives))
	}

	if len(directives) > 0 {
		if directives[0].Type != YAMLCodeBlock {
			t.Errorf("Expected type %s, got %s", YAMLCodeBlock, directives[0].Type)
		}
		if directives[0].Argument != "python" {
			t.Errorf("Expected language python, got %s", directives[0].Argument)
		}
	}
}

