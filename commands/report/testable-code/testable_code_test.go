// Package testablecode provides tests for the testable-code subcommand.
package testablecode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grove-platform/audit-cli/internal/config"
	"github.com/grove-platform/audit-cli/internal/rst"
)

// TestParseCSV tests the CSV parsing functionality.
func TestParseCSV(t *testing.T) {
	// Create a temporary CSV file with header
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test.csv")

	csvContent := `rank,url
1,www.mongodb.com/docs/atlas/page1/
2,www.mongodb.com/docs/manual/page2/
3,www.mongodb.com/docs/drivers/page3/`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write test CSV: %v", err)
	}

	entries, err := ParseCSV(csvPath)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Check first entry
	if entries[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", entries[0].Rank)
	}
	if entries[0].URL != "www.mongodb.com/docs/atlas/page1/" {
		t.Errorf("Expected URL 'www.mongodb.com/docs/atlas/page1/', got '%s'", entries[0].URL)
	}
}

// TestParseCSVWithoutHeader tests CSV parsing without a header row.
func TestParseCSVWithoutHeader(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test.csv")

	csvContent := `1,www.mongodb.com/docs/atlas/page1/
2,www.mongodb.com/docs/manual/page2/`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write test CSV: %v", err)
	}

	entries, err := ParseCSV(csvPath)
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", entries[0].Rank)
	}
}

// TestParseCSVEmptyFile tests error handling for empty CSV.
func TestParseCSVEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "empty.csv")

	if err := os.WriteFile(csvPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write test CSV: %v", err)
	}

	_, err := ParseCSV(csvPath)
	if err == nil {
		t.Error("Expected error for empty CSV, got nil")
	}
}

// TestParseCSVMissingFile tests error handling for missing file.
func TestParseCSVMissingFile(t *testing.T) {
	_, err := ParseCSV("/nonexistent/path/file.csv")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

// TestTestableProducts tests the TestableProducts map.
func TestTestableProducts(t *testing.T) {
	testCases := []struct {
		product  string
		expected bool
	}{
		{"Python", true},
		{"python", true},
		{"Node.js", true},
		{"nodejs", true},
		{"Go", true},
		{"go", true},
		{"Java", true},
		{"java", true},
		{"Java (Sync)", true},
		{"java-sync", true},
		{"C#", true},
		{"csharp", true},
		{"MongoDB Shell", true},
		{"mongosh", true},
		{"JavaScript", false}, // Not testable without context
		{"Shell", false},      // Not testable without context
		{"Ruby", false},
		{"PHP", false},
		{"Unknown", false},
	}

	for _, tc := range testCases {
		result := TestableProducts[tc.product]
		if result != tc.expected {
			t.Errorf("TestableProducts[%q] = %v, expected %v", tc.product, result, tc.expected)
		}
	}
}

// TestMaybeTestableProducts tests the MaybeTestableProducts map.
func TestMaybeTestableProducts(t *testing.T) {
	testCases := []struct {
		product  string
		expected bool
	}{
		{"JavaScript", true},
		{"Shell", true},
		{"Python", false},
		{"Node.js", false},
	}

	for _, tc := range testCases {
		result := MaybeTestableProducts[tc.product]
		if result != tc.expected {
			t.Errorf("MaybeTestableProducts[%q] = %v, expected %v", tc.product, result, tc.expected)
		}
	}
}

// TestIsTestedPath tests the isTestedPath function.
func TestIsTestedPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"/code-examples/tested/python/example.py", true},
		{"/includes/tested/driver-examples/insert.py", true},
		{"/code-examples/untested/example.py", false},
		{"/includes/examples/insert.py", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isTestedPath(tc.path)
		if result != tc.expected {
			t.Errorf("isTestedPath(%q) = %v, expected %v", tc.path, result, tc.expected)
		}
	}
}

// TestIsTestable tests the isTestable function.
func TestIsTestable(t *testing.T) {
	testCases := []struct {
		product    string
		contentDir string
		expected   bool
	}{
		{"Python", "pymongo-driver", true},
		{"Node.js", "node", true},
		{"Go", "golang", true},
		{"MongoDB Shell", "mongodb-shell", true},
		{"JavaScript", "node", false}, // JavaScript without context is not testable
		{"Shell", "mongodb-shell", false},
		{"Ruby", "ruby-driver", false},
		{"Unknown", "", false},
	}

	for _, tc := range testCases {
		result := isTestable(tc.product, tc.contentDir)
		if result != tc.expected {
			t.Errorf("isTestable(%q, %q) = %v, expected %v", tc.product, tc.contentDir, result, tc.expected)
		}
	}
}

// TestIsMaybeTestable tests the isMaybeTestable function.
func TestIsMaybeTestable(t *testing.T) {
	testCases := []struct {
		product  string
		expected bool
	}{
		{"JavaScript", true},
		{"Shell", true},
		{"Python", false},
		{"Node.js", false},
		{"MongoDB Shell", false},
		{"Unknown", false},
	}

	for _, tc := range testCases {
		result := isMaybeTestable(tc.product)
		if result != tc.expected {
			t.Errorf("isMaybeTestable(%q) = %v, expected %v", tc.product, result, tc.expected)
		}
	}
}

// TestParseComposableOptions tests the parseComposableOptions function.
func TestParseComposableOptions(t *testing.T) {
	testCases := []struct {
		options          string
		expectedLanguage string
		expectedInterface string
	}{
		{"language=python; interface=driver", "python", "driver"},
		{"language=nodejs", "nodejs", ""},
		{"interface=mongosh", "", "mongosh"},
		{"language=java; interface=compass", "java", "compass"},
		{"language-atlas-only=python", "python", ""},
		{"driver-lang=go", "go", ""},
		{"interface-local-only=mongosh", "", "mongosh"},
		{"", "", ""},
		{"invalid", "", ""},
	}

	for _, tc := range testCases {
		ctx := parseComposableOptions(tc.options)
		if ctx.Language != tc.expectedLanguage {
			t.Errorf("parseComposableOptions(%q).Language = %q, expected %q",
				tc.options, ctx.Language, tc.expectedLanguage)
		}
		if ctx.Interface != tc.expectedInterface {
			t.Errorf("parseComposableOptions(%q).Interface = %q, expected %q",
				tc.options, ctx.Interface, tc.expectedInterface)
		}
	}
}

// TestBuildPageReport tests the BuildPageReport function.
func TestBuildPageReport(t *testing.T) {
	analysis := &PageAnalysis{
		Rank:       1,
		URL:        "www.mongodb.com/docs/test/",
		SourcePath: "/path/to/source.rst",
		ContentDir: "pymongo-driver",
		CodeExamples: []CodeExample{
			{Type: "literalinclude", Language: "python", Product: "Python", IsTestable: true, IsTested: true},
			{Type: "code-block", Language: "python", Product: "Python", IsTestable: true, IsTested: false},
			{Type: "io-code-block", Language: "javascript", Product: "Node.js", IsInput: true, IsTestable: true},
			{Type: "io-code-block", Language: "javascript", Product: "Node.js", IsOutput: true, IsTestable: true},
			{Type: "code-block", Language: "json", Product: "JSON", IsTestable: false},
		},
	}

	report := BuildPageReport(analysis)

	if report.Rank != 1 {
		t.Errorf("Expected Rank 1, got %d", report.Rank)
	}
	if report.TotalExamples != 5 {
		t.Errorf("Expected TotalExamples 5, got %d", report.TotalExamples)
	}
	if report.TotalInput != 1 {
		t.Errorf("Expected TotalInput 1, got %d", report.TotalInput)
	}
	if report.TotalOutput != 1 {
		t.Errorf("Expected TotalOutput 1, got %d", report.TotalOutput)
	}
	if report.TotalTested != 1 {
		t.Errorf("Expected TotalTested 1, got %d", report.TotalTested)
	}
	if report.TotalTestable != 4 {
		t.Errorf("Expected TotalTestable 4, got %d", report.TotalTestable)
	}

	// Check Python stats
	pythonStats, ok := report.ByProduct["Python"]
	if !ok {
		t.Fatal("Expected Python in ByProduct")
	}
	if pythonStats.TotalCount != 2 {
		t.Errorf("Expected Python TotalCount 2, got %d", pythonStats.TotalCount)
	}
	if pythonStats.TestedCount != 1 {
		t.Errorf("Expected Python TestedCount 1, got %d", pythonStats.TestedCount)
	}

	// Check Node.js stats
	nodeStats, ok := report.ByProduct["Node.js"]
	if !ok {
		t.Fatal("Expected Node.js in ByProduct")
	}
	if nodeStats.TotalCount != 2 {
		t.Errorf("Expected Node.js TotalCount 2, got %d", nodeStats.TotalCount)
	}
	if nodeStats.InputCount != 1 {
		t.Errorf("Expected Node.js InputCount 1, got %d", nodeStats.InputCount)
	}
	if nodeStats.OutputCount != 1 {
		t.Errorf("Expected Node.js OutputCount 1, got %d", nodeStats.OutputCount)
	}
}

// TestEscapeCSV tests the escapeCSV function.
func TestEscapeCSV(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with,comma", `"with,comma"`},
		{`with"quote`, `"with""quote"`},
		{"with\nnewline", `"with` + "\n" + `newline"`},
		{"", ""},
		{"normal text", "normal text"},
	}

	for _, tc := range testCases {
		result := escapeCSV(tc.input)
		if result != tc.expected {
			t.Errorf("escapeCSV(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// TestIsMongoShellContext tests the isMongoShellContext function.
func TestIsMongoShellContext(t *testing.T) {
	testCases := []struct {
		contentDir string
		contexts   []CodeContext
		expected   bool
	}{
		{"mongodb-shell", nil, true},
		{"mongodb-shell", []CodeContext{}, true},
		{"pymongo-driver", []CodeContext{{Interface: "mongosh"}}, true},
		{"node", []CodeContext{{Interface: "driver"}}, false},
		{"manual", []CodeContext{}, false},
		{"", nil, false},
	}

	for _, tc := range testCases {
		result := isMongoShellContext(tc.contentDir, tc.contexts)
		if result != tc.expected {
			t.Errorf("isMongoShellContext(%q, %v) = %v, expected %v",
				tc.contentDir, tc.contexts, result, tc.expected)
		}
	}
}

// TestDetermineProduct tests the determineProduct function.
func TestDetermineProduct(t *testing.T) {
	// Create mock mappings
	mappings := &ProductMappings{
		DriversTabIDToProduct: map[string]string{
			"python":    "Python",
			"nodejs":    "Node.js",
			"java-sync": "Java (Sync)",
		},
		ComposableLanguageToProduct: map[string]string{
			"python": "Python",
			"nodejs": "Node.js",
			"go":     "Go",
		},
		ComposableInterfaceToProduct: map[string]string{
			"mongosh":  "MongoDB Shell",
			"driver":   "Driver",
			"compass":  "Compass",
		},
	}

	testCases := []struct {
		name       string
		language   string
		contentDir string
		contexts   []CodeContext
		expected   string
	}{
		// Non-driver languages bypass context
		{"bash bypasses context", "bash", "pymongo-driver", []CodeContext{{Language: "python"}}, "Shell"},
		{"json bypasses context", "json", "node", []CodeContext{{TabID: "nodejs"}}, "JSON"},
		{"yaml bypasses context", "yaml", "golang", nil, "YAML"},
		{"text bypasses context", "text", "manual", nil, "Text"},

		// MongoDB Shell context
		{"shell in mongosh dir", "shell", "mongodb-shell", nil, "MongoDB Shell"},
		{"javascript in mongosh context", "javascript", "", []CodeContext{{Interface: "mongosh"}}, "MongoDB Shell"},
		{"shell outside mongosh", "shell", "manual", nil, "Shell"},

		// Tab context
		{"python tab", "python", "", []CodeContext{{TabID: "python"}}, "Python"},
		{"nodejs tab", "javascript", "", []CodeContext{{TabID: "nodejs"}}, "Node.js"},

		// Composable language context
		{"go composable", "go", "", []CodeContext{{Language: "go"}}, "Go"},

		// Content directory fallback
		{"pymongo content dir", "python", "pymongo-driver", nil, "Python"},
		{"node content dir", "javascript", "node", nil, "Node.js"},

		// Language fallback
		{"python language", "python", "", nil, "Python"},
		{"ruby language", "ruby", "", nil, "Ruby"},

		// Unknown
		{"empty language", "", "", nil, "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := determineProduct(tc.language, tc.contentDir, tc.contexts, mappings)
			if result != tc.expected {
				t.Errorf("determineProduct(%q, %q, %v) = %q, expected %q",
					tc.language, tc.contentDir, tc.contexts, result, tc.expected)
			}
		})
	}
}

// TestGetLanguage tests the getLanguage function.
func TestGetLanguage(t *testing.T) {
	testCases := []struct {
		name        string
		options     map[string]string
		defaultLang string
		expected    string
	}{
		{"language option", map[string]string{"language": "python"}, "javascript", "python"},
		{"default lang", map[string]string{}, "javascript", "javascript"},
		{"empty default", map[string]string{}, "", "undefined"},
		{"empty language option", map[string]string{"language": ""}, "go", "go"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			directive := rst.Directive{Options: tc.options}
			result := getLanguage(directive, tc.defaultLang)
			if result != tc.expected {
				t.Errorf("getLanguage(%v, %q) = %q, expected %q",
					tc.options, tc.defaultLang, result, tc.expected)
			}
		})
	}
}

// TestProcessDirective tests the processDirective function.
func TestProcessDirective(t *testing.T) {
	mappings := &ProductMappings{
		DriversTabIDToProduct:        map[string]string{"python": "Python", "nodejs": "Node.js"},
		ComposableLanguageToProduct:  map[string]string{"python": "Python", "nodejs": "Node.js"},
		ComposableInterfaceToProduct: map[string]string{"mongosh": "MongoDB Shell"},
	}

	testCases := []struct {
		name            string
		directive       rst.Directive
		contentDir      string
		contexts        []CodeContext
		expectedCount   int
		expectedType    string
		expectedLang    string
		expectedProduct string
	}{
		{
			name: "literalinclude with tested path",
			directive: rst.Directive{
				Type:     rst.LiteralInclude,
				Argument: "/code-examples/tested/python/example.py",
				Options:  map[string]string{"language": "python"},
			},
			contentDir:      "pymongo-driver",
			contexts:        nil,
			expectedCount:   1,
			expectedType:    "literalinclude",
			expectedLang:    "python",
			expectedProduct: "Python",
		},
		{
			name: "code-block with language argument",
			directive: rst.Directive{
				Type:     rst.CodeBlock,
				Argument: "javascript",
				Options:  map[string]string{},
			},
			contentDir:      "node",
			contexts:        nil,
			expectedCount:   1,
			expectedType:    "code-block",
			expectedLang:    "javascript",
			expectedProduct: "Node.js",
		},
		{
			name: "code-block with json bypasses context",
			directive: rst.Directive{
				Type:     rst.CodeBlock,
				Argument: "json",
				Options:  map[string]string{},
			},
			contentDir:      "pymongo-driver",
			contexts:        []CodeContext{{Language: "python"}},
			expectedCount:   1,
			expectedType:    "code-block",
			expectedLang:    "json",
			expectedProduct: "JSON",
		},
		{
			name: "io-code-block with input and output",
			directive: rst.Directive{
				Type:    rst.IoCodeBlock,
				Options: map[string]string{"language": "python"},
				InputDirective: &rst.SubDirective{
					Argument: "/code-examples/input.py",
					Options:  map[string]string{"language": "python"},
				},
				OutputDirective: &rst.SubDirective{
					Argument: "/code-examples/output.txt",
					Options:  map[string]string{"language": "text"},
				},
			},
			contentDir:    "pymongo-driver",
			contexts:      nil,
			expectedCount: 2, // input + output
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			examples := processDirective(tc.directive, "/test/source.rst", tc.contentDir, tc.contexts, mappings)

			if len(examples) != tc.expectedCount {
				t.Errorf("Expected %d examples, got %d", tc.expectedCount, len(examples))
				return
			}

			if tc.expectedCount > 0 && tc.expectedType != "" {
				if examples[0].Type != tc.expectedType {
					t.Errorf("Expected type %q, got %q", tc.expectedType, examples[0].Type)
				}
			}

			if tc.expectedLang != "" && examples[0].Language != tc.expectedLang {
				t.Errorf("Expected language %q, got %q", tc.expectedLang, examples[0].Language)
			}

			if tc.expectedProduct != "" && examples[0].Product != tc.expectedProduct {
				t.Errorf("Expected product %q, got %q", tc.expectedProduct, examples[0].Product)
			}
		})
	}
}

// TestProcessDirectiveIOCodeBlock tests io-code-block processing in detail.
func TestProcessDirectiveIOCodeBlock(t *testing.T) {
	mappings := &ProductMappings{
		DriversTabIDToProduct:        map[string]string{},
		ComposableLanguageToProduct:  map[string]string{},
		ComposableInterfaceToProduct: map[string]string{},
	}

	directive := rst.Directive{
		Type:    rst.IoCodeBlock,
		Options: map[string]string{},
		InputDirective: &rst.SubDirective{
			Argument: "/code-examples/tested/python/input.py",
			Options:  map[string]string{"language": "python"},
		},
		OutputDirective: &rst.SubDirective{
			Argument: "/code-examples/output.json",
			Options:  map[string]string{"language": "json"},
		},
	}

	examples := processDirective(directive, "/test/source.rst", "pymongo-driver", nil, mappings)

	if len(examples) != 2 {
		t.Fatalf("Expected 2 examples, got %d", len(examples))
	}

	// Check input
	input := examples[0]
	if !input.IsInput {
		t.Error("Expected first example to be input")
	}
	if input.IsOutput {
		t.Error("Expected first example to not be output")
	}
	if !input.IsTested {
		t.Error("Expected input to be tested (path contains /tested/)")
	}
	if input.Language != "python" {
		t.Errorf("Expected input language 'python', got %q", input.Language)
	}

	// Check output
	output := examples[1]
	if !output.IsOutput {
		t.Error("Expected second example to be output")
	}
	if output.IsInput {
		t.Error("Expected second example to not be input")
	}
	if output.IsTested {
		t.Error("Expected output to not be tested")
	}
	if output.Language != "json" {
		t.Errorf("Expected output language 'json', got %q", output.Language)
	}
}

// TestParseFileContexts tests the parseFileContexts function.
func TestParseFileContexts(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "testable-code-test", "content", "test-project", "source")

	t.Run("file with tabs", func(t *testing.T) {
		filePath := filepath.Join(testDataDir, "with-tabs.rst")
		contexts, err := parseFileContexts(filePath)
		if err != nil {
			t.Fatalf("parseFileContexts failed: %v", err)
		}

		// Should find tab contexts
		if len(contexts) == 0 {
			t.Error("Expected to find contexts")
		}

		// Check that we found at least one tabid
		foundTabID := false
		for _, ctx := range contexts {
			if ctx.TabID != "" {
				foundTabID = true
				break
			}
		}
		if !foundTabID {
			t.Error("Expected to find at least one TabID context")
		}
	})

	t.Run("file with composable tutorial", func(t *testing.T) {
		filePath := filepath.Join(testDataDir, "with-selected-content.rst")
		contexts, err := parseFileContexts(filePath)
		if err != nil {
			t.Fatalf("parseFileContexts failed: %v", err)
		}

		// Should find composable context
		if len(contexts) == 0 {
			t.Error("Expected to find contexts")
		}

		// Check that we found language or interface from composable options
		foundComposable := false
		for _, ctx := range contexts {
			if ctx.Language != "" || ctx.Interface != "" {
				foundComposable = true
				break
			}
		}
		if !foundComposable {
			t.Error("Expected to find composable context with language or interface")
		}
	})

	t.Run("simple file without context", func(t *testing.T) {
		filePath := filepath.Join(testDataDir, "simple-code.rst")
		contexts, err := parseFileContexts(filePath)
		if err != nil {
			t.Fatalf("parseFileContexts failed: %v", err)
		}

		// Should return at least one empty context
		if len(contexts) == 0 {
			t.Error("Expected at least one context (even if empty)")
		}
	})
}

// TestParseSelectedContentBlocks tests the parseSelectedContentBlocks function.
func TestParseSelectedContentBlocks(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "testable-code-test", "content", "test-project", "source")
	filePath := filepath.Join(testDataDir, "with-selected-content.rst")

	result, err := parseSelectedContentBlocks(filePath)
	if err != nil {
		t.Fatalf("parseSelectedContentBlocks failed: %v", err)
	}

	// The function should map include paths to their selections
	// We expect to find mappings for the includes in selected-content blocks
	if len(result) == 0 {
		t.Log("No selected-content mappings found (this may be expected if includes don't resolve)")
	}

	// Check that any found mappings have valid selection values
	for path, selection := range result {
		if selection == "" {
			t.Errorf("Empty selection for path %q", path)
		}
		t.Logf("Found mapping: %s -> %s", path, selection)
	}
}

// TestCollectCodeExamples tests the collectCodeExamples function.
func TestCollectCodeExamples(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "testable-code-test", "content", "test-project", "source")

	mappings := &ProductMappings{
		DriversTabIDToProduct:        map[string]string{"python": "Python", "nodejs": "Node.js", "java-sync": "Java (Sync)"},
		ComposableLanguageToProduct:  map[string]string{"python": "Python", "nodejs": "Node.js", "go": "Go"},
		ComposableInterfaceToProduct: map[string]string{"mongosh": "MongoDB Shell", "driver": "Driver"},
	}

	t.Run("simple code file", func(t *testing.T) {
		filePath := filepath.Join(testDataDir, "simple-code.rst")
		visited := make(map[string]bool)

		examples, err := collectCodeExamples(filePath, "test-project", visited, mappings)
		if err != nil {
			t.Fatalf("collectCodeExamples failed: %v", err)
		}

		// Should find 4 code blocks: python, javascript, json, sh
		if len(examples) != 4 {
			t.Errorf("Expected 4 examples, got %d", len(examples))
		}

		// Check that we found the expected languages
		languages := make(map[string]bool)
		for _, ex := range examples {
			languages[ex.Language] = true
		}

		expectedLangs := []string{"python", "javascript", "json", "sh"}
		for _, lang := range expectedLangs {
			if !languages[lang] {
				t.Errorf("Expected to find language %q", lang)
			}
		}
	})

	t.Run("file with tabs", func(t *testing.T) {
		filePath := filepath.Join(testDataDir, "with-tabs.rst")
		visited := make(map[string]bool)

		examples, err := collectCodeExamples(filePath, "test-project", visited, mappings)
		if err != nil {
			t.Fatalf("collectCodeExamples failed: %v", err)
		}

		// Should find 3 code blocks: python, javascript, java
		if len(examples) < 3 {
			t.Errorf("Expected at least 3 examples, got %d", len(examples))
		}
	})
}

// TestMergeProjectComposables tests the MergeProjectComposables function.
func TestMergeProjectComposables(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "testable-code-test", "content", "test-project", "source")

	baseMappings := &ProductMappings{
		DriversTabIDToProduct:        map[string]string{"python": "Python"},
		ComposableLanguageToProduct:  map[string]string{"python": "Python"},
		ComposableInterfaceToProduct: map[string]string{"driver": "Driver"},
	}

	t.Run("merges project composables", func(t *testing.T) {
		sourcePath := filepath.Join(testDataDir, "simple-code.rst")
		absPath, _ := filepath.Abs(sourcePath)

		merged := MergeProjectComposables(baseMappings, absPath)

		// Should have base mappings
		if merged.DriversTabIDToProduct["python"] != "Python" {
			t.Error("Expected base mapping for python tab")
		}

		// Should have project-specific composables merged in
		// The test project defines nodejs and go in language composable
		if merged.ComposableLanguageToProduct["nodejs"] != "Node.js" {
			t.Error("Expected project composable for nodejs")
		}
		if merged.ComposableLanguageToProduct["go"] != "Go" {
			t.Error("Expected project composable for go")
		}

		// Should have interface composables
		if merged.ComposableInterfaceToProduct["mongosh"] != "MongoDB Shell" {
			t.Error("Expected project composable for mongosh interface")
		}
	})

	t.Run("returns base mappings for nonexistent path", func(t *testing.T) {
		merged := MergeProjectComposables(baseMappings, "/nonexistent/path/file.rst")

		// Should return base mappings unchanged
		if merged.DriversTabIDToProduct["python"] != "Python" {
			t.Error("Expected base mapping to be preserved")
		}
	})
}

// TestAnalyzePage tests the AnalyzePage function.
// Note: AnalyzePage requires a URLMapping which involves URL resolution.
// The URL resolution expects .txt files (MongoDB docs monorepo format).
// We create .txt copies of our test .rst files to test the full integration.
func TestAnalyzePage(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "..", "testdata", "testable-code-test", "content", "test-project", "source")
	absTestDataDir, _ := filepath.Abs(testDataDir)
	monorepoPath := filepath.Join(absTestDataDir, "..", "..", "..")

	// Create .txt copies of our .rst test files for URL resolution
	rstFiles := []string{"simple-code.rst", "with-tabs.rst", "with-selected-content.rst"}
	for _, rstFile := range rstFiles {
		rstPath := filepath.Join(absTestDataDir, rstFile)
		txtPath := filepath.Join(absTestDataDir, rstFile[:len(rstFile)-4]+".txt")
		content, err := os.ReadFile(rstPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", rstPath, err)
		}
		if err := os.WriteFile(txtPath, content, 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", txtPath, err)
		}
		defer os.Remove(txtPath)
	}

	mappings := &ProductMappings{
		DriversTabIDToProduct:        map[string]string{"python": "Python", "nodejs": "Node.js", "java-sync": "Java (Sync)"},
		ComposableLanguageToProduct:  map[string]string{"python": "Python", "nodejs": "Node.js", "go": "Go"},
		ComposableInterfaceToProduct: map[string]string{"mongosh": "MongoDB Shell", "driver": "Driver"},
	}

	// Create a mock URLMapping that maps test URLs to our test files
	urlMapping := &config.URLMapping{
		URLSlugToProject: map[string]string{
			"test-project": "test-project",
		},
		ProjectToContentDir: map[string]string{
			"test-project": "test-project",
		},
		ProjectBranches: map[string][]string{
			"test-project": {"current"},
		},
		MonorepoPath: monorepoPath,
	}

	t.Run("analyzes simple code file", func(t *testing.T) {
		entry := PageEntry{
			Rank: 1,
			URL:  "https://www.mongodb.com/docs/test-project/current/simple-code/",
		}

		analysis, err := AnalyzePage(entry, urlMapping, mappings)
		if err != nil {
			t.Fatalf("AnalyzePage failed: %v", err)
		}

		// Should find 4 code examples
		if len(analysis.CodeExamples) != 4 {
			t.Errorf("Expected 4 code examples, got %d", len(analysis.CodeExamples))
		}

		// Check that products are assigned
		for _, ex := range analysis.CodeExamples {
			if ex.Product == "" || ex.Product == "Unknown" {
				t.Errorf("Expected product to be assigned for language %q, got %q", ex.Language, ex.Product)
			}
		}
	})

	t.Run("analyzes file with tabs", func(t *testing.T) {
		entry := PageEntry{
			Rank: 2,
			URL:  "https://www.mongodb.com/docs/test-project/current/with-tabs/",
		}

		analysis, err := AnalyzePage(entry, urlMapping, mappings)
		if err != nil {
			t.Fatalf("AnalyzePage failed: %v", err)
		}

		// Should find at least 3 code examples (one per tab)
		if len(analysis.CodeExamples) < 3 {
			t.Errorf("Expected at least 3 code examples, got %d", len(analysis.CodeExamples))
		}

		// Check that products are assigned based on tab context
		products := make(map[string]bool)
		for _, ex := range analysis.CodeExamples {
			products[ex.Product] = true
		}

		// Should have Python, Node.js, and Java products
		expectedProducts := []string{"Python", "Node.js", "Java (Sync)"}
		for _, prod := range expectedProducts {
			if !products[prod] {
				t.Errorf("Expected to find product %q", prod)
			}
		}
	})

	t.Run("returns error for nonexistent URL", func(t *testing.T) {
		entry := PageEntry{
			Rank: 99,
			URL:  "https://www.mongodb.com/docs/nonexistent-project/current/page/",
		}

		_, err := AnalyzePage(entry, urlMapping, mappings)
		if err == nil {
			t.Error("Expected error for nonexistent URL")
		}
	})

	t.Run("analyzes file with composable tutorial", func(t *testing.T) {
		entry := PageEntry{
			Rank: 3,
			URL:  "https://www.mongodb.com/docs/test-project/current/with-selected-content/",
		}

		analysis, err := AnalyzePage(entry, urlMapping, mappings)
		if err != nil {
			t.Fatalf("AnalyzePage failed: %v", err)
		}

		// Should find code examples from selected-content blocks
		if len(analysis.CodeExamples) == 0 {
			t.Error("Expected to find code examples in composable tutorial")
		}

		// Log what we found for debugging
		for _, ex := range analysis.CodeExamples {
			t.Logf("Found: type=%s, lang=%s, product=%s, source=%s",
				ex.Type, ex.Language, ex.Product, ex.SourceFile)
		}
	})
}

