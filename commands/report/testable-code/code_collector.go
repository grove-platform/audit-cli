package testablecode

import (
	"bufio"
	"os"
	"strings"

	"github.com/grove-platform/audit-cli/internal/config"
	lang "github.com/grove-platform/audit-cli/internal/language"
	"github.com/grove-platform/audit-cli/internal/projectinfo"
	"github.com/grove-platform/audit-cli/internal/rst"
)

// AnalyzePage analyzes a single page for code examples.
//
// This function resolves a URL to its source file in the monorepo, then collects
// all code examples from that file and any files it includes. The analysis includes:
//   - Identifying the directive type (literalinclude, code-block, io-code-block)
//   - Determining the product/language context for each example
//   - Checking if the example is tested (references tested code)
//   - Checking if the example is testable (based on product)
//
// The contentDir is extracted from the source path and used for product determination
// when no explicit context (tabs, composables) is available.
func AnalyzePage(entry PageEntry, urlMapping *config.URLMapping, mappings *ProductMappings) (*PageAnalysis, error) {
	// Resolve URL to source file
	sourcePath, contentDir, err := urlMapping.ResolveURL(entry.URL)
	if err != nil {
		return nil, err
	}

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, err
	}

	// Merge project-specific composables from snooty.toml
	// This allows projects like Atlas to define custom composables that override rstspec.toml
	mergedMappings := MergeProjectComposables(mappings, sourcePath)

	analysis := &PageAnalysis{
		Rank:       entry.Rank,
		URL:        entry.URL,
		SourcePath: sourcePath,
		ContentDir: contentDir,
	}

	// Collect code examples from the file and its includes
	visited := make(map[string]bool)
	examples, err := collectCodeExamples(sourcePath, contentDir, visited, mergedMappings)
	if err != nil {
		return nil, err
	}

	analysis.CodeExamples = examples
	return analysis, nil
}

// collectCodeExamples collects all code examples from a file and its includes.
//
// This is the public entry point that starts collection with no inherited context.
// It delegates to collectCodeExamplesWithContext, which does the actual work.
//
// WHY THIS WRAPPER EXISTS:
// The collection process is recursive - when we encounter an `.. include::` directive,
// we need to collect code examples from the included file too. However, included files
// may need to inherit context from their parent (e.g., a `selected-content` block).
//
// This wrapper provides a clean entry point for external callers (like AnalyzePage)
// who just want to say "analyze this file" without worrying about context inheritance.
// The recursive collectCodeExamplesWithContext handles passing context through the
// include chain internally.
//
// Flow:
//
//	collectCodeExamples(main.txt)           ← entry point, nil context
//	  └── collectCodeExamplesWithContext(main.txt, nil)
//	        └── collectCodeExamplesWithContext(included.rst, inherited context)
//	              └── collectCodeExamplesWithContext(nested.rst, inherited context)
func collectCodeExamples(filePath, contentDir string, visited map[string]bool, mappings *ProductMappings) ([]CodeExample, error) {
	return collectCodeExamplesWithContext(filePath, contentDir, visited, nil, mappings)
}

// collectCodeExamplesWithContext collects code examples with inherited context from parent.
//
// CONTENT INCLUSION TYPES HANDLED:
// This function recursively follows content inclusions to find all code examples.
// The rst.FindIncludeDirectives function finds `.. include::` directives and resolves
// paths using MongoDB-specific conventions:
//
//  1. Regular RST includes: `.. include:: /includes/foo.rst` → resolved directly
//  2. Steps files: `.. include:: /includes/steps/foo.rst` → resolved to steps-foo.yaml
//  3. Extracts files: `.. include:: /includes/extracts/foo.rst` → resolved to YAML with ref
//  4. Template variables: `.. include:: {{var}}` → resolved from replacement section
//
// For YAML files (steps, extracts), rst.ParseDirectives handles two formats:
//  1. RST-in-YAML: RST directives embedded in YAML content (e.g., `.. code-block::` in `content: |` blocks)
//  2. YAML-native: Legacy `action:` blocks with `language:` and `code:` fields (added January 2026)
//
// CONTEXT INHERITANCE:
// When a file is included via `.. include::` within a `.. selected-content::` block
// or a `.. tab::` block, the code examples in that included file should inherit the
// context (language/product) from the parent block. This is critical for accurate
// product attribution because:
//
//  1. Many driver docs use composable tutorials where the main file has
//     `.. selected-content:: :selections: python` and includes a shared file
//     that contains the actual code examples.
//
//  2. Without context inheritance, those code examples would be attributed to
//     their raw language (e.g., "python") rather than the proper product context
//     (e.g., "Python" driver).
//
// The parentContext parameter carries this inherited context through the include chain.
func collectCodeExamplesWithContext(filePath, contentDir string, visited map[string]bool, parentContext *CodeContext, mappings *ProductMappings) ([]CodeExample, error) {
	if visited[filePath] {
		return nil, nil
	}
	visited[filePath] = true

	var examples []CodeExample

	// Parse directives from the file
	directives, err := rst.ParseDirectives(filePath)
	if err != nil {
		return nil, err
	}

	// Parse selected-content blocks to get context for includes
	selectedContentMap, err := parseSelectedContentBlocks(filePath)
	if err != nil {
		selectedContentMap = make(map[string]string)
	}

	// Parse context blocks (tabs, composables) with their line ranges
	// This allows us to match each directive to its containing context block
	var contextBlocks []contextBlock
	var fileContext []CodeContext
	if parentContext != nil {
		// If we have a parent context, use it for all directives
		fileContext = []CodeContext{*parentContext}
	} else {
		// Parse context blocks to get line-range-aware context
		contextBlocks, err = parseContextBlocks(filePath)
		if err != nil {
			contextBlocks = nil
		}
		// Also get file-level context (composable-tutorial options apply to whole file)
		fileContext, err = parseFileContexts(filePath)
		if err != nil {
			fileContext = []CodeContext{{}}
		}
	}

	// Process each directive with its specific context
	for _, directive := range directives {
		// Find the context for this directive based on its line number
		contexts := findContextForLine(directive.LineNum, contextBlocks, fileContext)
		exs := processDirective(directive, filePath, contentDir, contexts, mappings)
		examples = append(examples, exs...)
	}

	// Follow includes with their selected-content context
	includeFiles, err := rst.FindIncludeDirectives(filePath)
	if err == nil {
		for _, includeFile := range includeFiles {
			// Check if this include has a selected-content or tab context
			var includeContext *CodeContext
			if selection, ok := selectedContentMap[includeFile]; ok {
				// Determine if this is a tabid or a composable language selection
				// by checking which mapping contains it
				if _, isTabID := mappings.DriversTabIDToProduct[selection]; isTabID {
					includeContext = &CodeContext{TabID: selection}
				} else {
					// Treat as composable language selection
					includeContext = &CodeContext{Language: selection}
				}
			} else if parentContext != nil {
				includeContext = parentContext
			}

			includedExamples, err := collectCodeExamplesWithContext(includeFile, contentDir, visited, includeContext, mappings)
			if err == nil {
				examples = append(examples, includedExamples...)
			}
		}
	}

	return examples, nil
}

// contextBlock represents a context-providing block (tab or selected-content) with its line range.
type contextBlock struct {
	context   CodeContext
	startLine int
	endLine   int // -1 means extends to end of file or next block at same level
}

// parseContextBlocks parses a file to extract tab and selected-content blocks with their line ranges.
// This allows matching code examples to their containing context block.
func parseContextBlocks(filePath string) ([]contextBlock, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var blocks []contextBlock
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Track open blocks by their indentation level
	type openBlock struct {
		context CodeContext
		start   int
		indent  int
	}
	var openBlocks []openBlock

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " ")
		trimmedLine := strings.TrimSpace(line)
		indent := len(line) - len(trimmed)

		// Close any blocks that have ended (non-empty line at same or less indentation)
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, ":") {
			for i := len(openBlocks) - 1; i >= 0; i-- {
				if indent <= openBlocks[i].indent {
					// This block has ended
					blocks = append(blocks, contextBlock{
						context:   openBlocks[i].context,
						startLine: openBlocks[i].start,
						endLine:   lineNum - 1,
					})
					openBlocks = openBlocks[:i]
				}
			}
		}

		// Check for tab directive
		if rst.TabDirectiveRegex.MatchString(trimmedLine) {
			openBlocks = append(openBlocks, openBlock{
				context: CodeContext{}, // TabID will be filled in when we find :tabid:
				start:   lineNum,
				indent:  indent,
			})
			continue
		}

		// Check for selected-content directive
		if rst.SelectedContentDirectiveRegex.MatchString(trimmedLine) {
			openBlocks = append(openBlocks, openBlock{
				context: CodeContext{}, // Selection will be filled in when we find :selections:
				start:   lineNum,
				indent:  indent,
			})
			continue
		}

		// Look for :tabid: option to fill in the most recent tab block
		if len(openBlocks) > 0 {
			if matches := rst.TabIDOptionRegex.FindStringSubmatch(line); len(matches) > 1 {
				openBlocks[len(openBlocks)-1].context.TabID = strings.TrimSpace(matches[1])
				continue
			}
			// Look for :selections: option
			if matches := rst.SelectionsOptionRegex.FindStringSubmatch(line); len(matches) > 1 {
				openBlocks[len(openBlocks)-1].context.Language = strings.TrimSpace(matches[1])
				continue
			}
		}
	}

	// Close any remaining open blocks (they extend to end of file)
	for _, ob := range openBlocks {
		blocks = append(blocks, contextBlock{
			context:   ob.context,
			startLine: ob.start,
			endLine:   lineNum,
		})
	}

	return blocks, scanner.Err()
}

// findContextForLine finds the context that applies to a given line number.
// It checks context blocks first (tabs, selected-content), then falls back to file-level context.
func findContextForLine(lineNum int, contextBlocks []contextBlock, fileContext []CodeContext) []CodeContext {
	// Check if this line is inside any context block
	for _, block := range contextBlocks {
		if lineNum >= block.startLine && lineNum <= block.endLine {
			// Found a containing block - use its context
			if block.context.TabID != "" || block.context.Language != "" || block.context.Interface != "" {
				return []CodeContext{block.context}
			}
		}
	}

	// Fall back to file-level context
	return fileContext
}

// CodeContext represents the context in which a code example appears.
//
// MongoDB documentation uses several mechanisms to provide context for code examples:
//
//  1. Driver Tab Sets (`.. tabs-drivers::` with `.. tab::` and `:tabid:`):
//     Used to show the same concept in multiple driver languages. The :tabid:
//     identifies which driver (e.g., "python", "nodejs", "java-sync").
//
//  2. Composable Tutorials (`.. composable-tutorial::` with `:options:`):
//     Used for tutorials that can be customized by language and interface.
//     Options like "language=python; interface=driver" specify the context.
//
//  3. Selected Content (`.. selected-content::` with `:selections:`):
//     Used within composable tutorials to show content for a specific selection.
//     The :selections: value identifies which option is active.
//
// These contexts are used to determine the product for code examples, which in turn
// determines whether the example is testable (has test infrastructure).
type CodeContext struct {
	TabID      string // From :tabid: option in .. tab:: directive
	Composable string // From composable-tutorial options (unused, kept for future)
	Interface  string // From interface composable (e.g., "mongosh", "driver", "compass")
	Language   string // From language composable (e.g., "python", "nodejs", "java")
}

// processDirective converts an RST directive to CodeExample(s).
//
// This function handles five types of code example directives:
//   - literalinclude: Transcludes code from an external file
//   - code-block: Inline code block with language specification
//   - code: Shorter alias for code-block (standard reStructuredText, parsed as code-block)
//   - io-code-block: Input/output code example with separate input and output blocks
//   - yaml-code-block: YAML-native code examples from legacy steps files (action: blocks)
//
// For each directive, it determines the product based on the language and context,
// checks if the example is tested (references tested code), and checks if it's testable.
func processDirective(directive rst.Directive, sourceFile, contentDir string, contexts []CodeContext, mappings *ProductMappings) []CodeExample {
	var examples []CodeExample

	switch directive.Type {
	case rst.LiteralInclude:
		ex := CodeExample{
			Type:       string(rst.LiteralInclude),
			FilePath:   directive.Argument,
			SourceFile: sourceFile,
		}
		ex.Language = directive.ResolveLanguage()
		ex.IsTested = isTestedPath(directive.Argument)
		ex.Product = determineProduct(ex.Language, contentDir, contexts, mappings)
		ex.IsTestable = isTestable(ex.Product, contentDir)
		ex.IsMaybeTestable = isMaybeTestable(ex.Product)
		examples = append(examples, ex)

	case rst.CodeBlock:
		ex := CodeExample{
			Type:       string(rst.CodeBlock),
			SourceFile: sourceFile,
		}
		ex.Language = getLanguage(directive, directive.Argument)
		ex.Product = determineProduct(ex.Language, contentDir, contexts, mappings)
		ex.IsTestable = isTestable(ex.Product, contentDir)
		ex.IsMaybeTestable = isMaybeTestable(ex.Product)
		examples = append(examples, ex)

	case rst.IoCodeBlock:
		// Process input directive
		if directive.InputDirective != nil {
			ex := CodeExample{
				Type:       string(rst.IoCodeBlock),
				IsInput:    true,
				FilePath:   directive.InputDirective.Argument,
				SourceFile: sourceFile,
			}
			ex.Language = directive.InputDirective.ResolveLanguage(directive.Options)
			ex.IsTested = isTestedPath(directive.InputDirective.Argument)
			ex.Product = determineProduct(ex.Language, contentDir, contexts, mappings)
			ex.IsTestable = isTestable(ex.Product, contentDir)
			ex.IsMaybeTestable = isMaybeTestable(ex.Product)
			examples = append(examples, ex)
		}

		// Process output directive
		if directive.OutputDirective != nil {
			ex := CodeExample{
				Type:       string(rst.IoCodeBlock),
				IsOutput:   true,
				FilePath:   directive.OutputDirective.Argument,
				SourceFile: sourceFile,
			}
			ex.Language = directive.OutputDirective.ResolveLanguage(directive.Options)
			ex.IsTested = isTestedPath(directive.OutputDirective.Argument)
			ex.Product = determineProduct(ex.Language, contentDir, contexts, mappings)
			ex.IsTestable = isTestable(ex.Product, contentDir)
			ex.IsMaybeTestable = isMaybeTestable(ex.Product)
			examples = append(examples, ex)
		}

	case rst.YAMLCodeBlock:
		// YAML-native code examples from legacy steps files (action: blocks)
		ex := CodeExample{
			Type:       string(rst.YAMLCodeBlock),
			SourceFile: sourceFile,
		}
		ex.Language = getLanguage(directive, directive.Argument)
		ex.Product = determineProduct(ex.Language, contentDir, contexts, mappings)
		ex.IsTestable = isTestable(ex.Product, contentDir)
		ex.IsMaybeTestable = isMaybeTestable(ex.Product)
		examples = append(examples, ex)
	}

	return examples
}

// getLanguage extracts the language from a directive.
// Checks the :language: option first, then falls back to defaultLang.
// If defaultLang is empty, returns lang.Undefined.
func getLanguage(directive rst.Directive, defaultLang string) string {
	if langOpt, ok := directive.Options["language"]; ok && langOpt != "" {
		return langOpt
	}
	if defaultLang != "" {
		return defaultLang
	}
	return lang.Undefined
}

// isTestedPath checks if a file path references tested code.
func isTestedPath(path string) bool {
	return strings.Contains(path, "/tested/")
}

// isTestable checks if a code example is testable based on its product and content directory.
//
// A code example is considered testable if it meets one of these criteria:
//  1. It's in a testable content directory (e.g., csharp, golang, java, mongodb-shell, node, pymongo-driver)
//  2. Its product (after context resolution) is in the TestableProducts list
//
// Context resolution determines the product through:
//   - Composable tutorial selected-content blocks (e.g., :selections: nodejs)
//   - Driver tab sets with :tabid: (e.g., :tabid: python)
//   - Content directory mapping (e.g., content in "node" dir → Node.js)
//
// Raw language values like "javascript" and "shell" are intentionally NOT testable
// because many code examples use these languages without being actual Driver/Shell examples.
// Only properly contextualized examples are considered testable.
//
// Note: Being in a testable content directory (like mongodb-shell) does NOT automatically
// make all code examples testable. System shell commands (sh, bash) in the MongoDB Shell
// docs are still not testable - only actual MongoDB Shell code is testable.
func isTestable(product, contentDir string) bool {
	// Check if product is testable
	return TestableProducts[product]
}

// isMaybeTestable checks if a code example is in the "grey area" - it uses a language
// that COULD be testable (javascript, shell) but lacks proper context to determine definitively.
//
// This applies to:
//   - "JavaScript" product: Could be Node.js driver code OR browser JavaScript
//   - "Shell" product: Could be MongoDB Shell code OR bash/system commands
//
// These examples are NOT counted as testable (to avoid false positives) but are flagged
// separately so they can be reviewed and potentially re-categorized.
func isMaybeTestable(product string) bool {
	return MaybeTestableProducts[product]
}

// determineProduct determines the product from language, content dir, and context.
//
// The logic handles several special cases:
//
//  1. Non-driver languages (bash, sh, json, yaml, etc.) bypass context inheritance
//     and are reported based on their actual language. This prevents shell commands
//     like "npm install" from being counted as Node.js driver examples.
//
// 2. MongoDB Shell languages (shell, javascript, js) have special handling:
//   - In MongoDB Shell context (mongosh content dir or mongosh interface) → "MongoDB Shell"
//   - "shell" outside MongoDB Shell context → "Shell" (not testable)
//   - "javascript/js" outside MongoDB Shell context → use driver context or "JavaScript"
func determineProduct(language, contentDir string, contexts []CodeContext, mappings *ProductMappings) string {
	// Check if this is a non-driver language that should bypass context inheritance.
	// These languages should be reported based on their actual language, not the
	// surrounding composable/tab context.
	if language != "" && lang.IsNonDriverLanguage(language) {
		return lang.GetProductFromLanguage(language)
	}

	// Check if we're in a MongoDB Shell context
	inMongoShellContext := isMongoShellContext(contentDir, contexts)

	// Handle MongoDB Shell languages specially
	if language != "" && lang.IsMongoShellLanguage(language) {
		if inMongoShellContext {
			return "MongoDB Shell"
		}
		// "shell" outside MongoDB Shell context is just a shell command
		langLower := strings.ToLower(language)
		if langLower == "shell" {
			return "Shell"
		}
		// "javascript" or "js" outside MongoDB Shell context - check for driver context
		// (fall through to normal context checking below)
	}

	// Check if we have a context with a specific product
	for _, ctx := range contexts {
		if ctx.TabID != "" {
			if product, ok := mappings.DriversTabIDToProduct[ctx.TabID]; ok {
				return product
			}
		}
		if ctx.Language != "" {
			if product, ok := mappings.ComposableLanguageToProduct[ctx.Language]; ok {
				return product
			}
		}
		if ctx.Interface != "" {
			if product, ok := mappings.ComposableInterfaceToProduct[ctx.Interface]; ok {
				return product
			}
		}
	}

	// Map content directory to product using shared mapping
	if product := projectinfo.GetProductFromContentDir(contentDir); product != "" {
		return product
	}

	// Fall back to language
	if language != "" {
		return lang.GetProductFromLanguage(language)
	}

	return "Unknown"
}

// isMongoShellContext checks if we're in a MongoDB Shell context based on
// content directory or composable/tab context.
func isMongoShellContext(contentDir string, contexts []CodeContext) bool {
	// Check content directory
	if contentDir == "mongodb-shell" {
		return true
	}

	// Check for mongosh interface in composable context
	for _, ctx := range contexts {
		if ctx.Interface == "mongosh" {
			return true
		}
	}

	return false
}

// parseSelectedContentBlocks parses a file to map include paths to their
// selected-content or tab context. Returns a map from include path to selection/tabid.
//
// WHY THIS EXISTS:
// In composable tutorials and tabbed content, include directives often appear inside
// context-providing blocks. For example:
//
//	.. selected-content::
//	   :selections: python
//
//	   .. include:: /includes/driver-examples/insert-one.rst
//
// This function builds a map from include paths to their context, which is then
// used during include processing to pass the correct context to determineProduct.
//
// IMPORTANT: Context inheritance only applies to driver-appropriate languages.
// Non-driver languages (see NonDriverLanguages in internal/language) bypass context entirely:
//   - A "python" code block in the included file → attributed to "Python" (from context)
//   - A "text" code block in the included file → attributed to "Text" (bypasses context)
//   - A "sh" code block in the included file → attributed to "Shell" (bypasses context)
//
// The function handles both:
//   - selected-content blocks with :selections: option
//   - tab blocks with :tabid: option
func parseSelectedContentBlocks(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	var currentSelection string
	var currentTabID string
	var inSelectedContent bool
	var inTab bool
	var blockIndent int

	for scanner.Scan() {
		line := scanner.Text()

		// Calculate indentation
		trimmed := strings.TrimLeft(line, " ")
		trimmedLine := strings.TrimSpace(line)
		indent := len(line) - len(trimmed)

		// Check for selected-content directive
		if rst.SelectedContentDirectiveRegex.MatchString(trimmedLine) {
			inSelectedContent = true
			inTab = false
			blockIndent = indent
			currentSelection = ""
			continue
		}

		// Check for tab directive
		if matches := rst.TabDirectiveRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			inTab = true
			inSelectedContent = false
			blockIndent = indent
			currentTabID = ""
			continue
		}

		// If we're in a selected-content block, look for :selections:
		if inSelectedContent && currentSelection == "" {
			if matches := rst.SelectionsOptionRegex.FindStringSubmatch(line); len(matches) > 1 {
				currentSelection = strings.TrimSpace(matches[1])
				continue
			}
		}

		// If we're in a tab block, look for :tabid:
		if inTab && currentTabID == "" {
			if matches := rst.TabIDOptionRegex.FindStringSubmatch(line); len(matches) > 1 {
				currentTabID = strings.TrimSpace(matches[1])
				continue
			}
		}

		// Check if we've exited the block (less or equal indentation on non-empty line)
		if trimmedLine != "" && indent <= blockIndent && !strings.HasPrefix(trimmed, ":") {
			// Check if this is a new directive at same level
			if strings.HasPrefix(trimmed, "..") {
				// Could be a new selected-content or tab, handled above
				// Or could be something else that ends our block
				if !rst.SelectedContentDirectiveRegex.MatchString(trimmedLine) && !rst.TabDirectiveRegex.MatchString(trimmedLine) {
					inSelectedContent = false
					inTab = false
					currentSelection = ""
					currentTabID = ""
				}
			}
		}

		// Look for include directives within the current context
		if matches := rst.IncludeDirectiveRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			includePath := strings.TrimSpace(matches[1])
			// Resolve the include path to match what FindIncludeDirectives returns
			resolvedPath, err := rst.ResolveIncludePath(filePath, includePath)
			if err == nil {
				if inSelectedContent && currentSelection != "" {
					result[resolvedPath] = currentSelection
				} else if inTab && currentTabID != "" {
					result[resolvedPath] = currentTabID
				}
			}
		}
	}

	return result, scanner.Err()
}

// parseFileContexts parses a file to extract tab and composable contexts.
//
// This function scans a file for context-providing directives and extracts their
// configuration. It looks for:
//
//  1. Tab directives with :tabid: - Used in driver tab sets to identify the driver
//  2. Composable tutorials with :options: - Used to specify language and interface
//
// KNOWN LIMITATION:
// This function extracts ALL contexts from the file into a flat list, without
// tracking which code examples are inside which context blocks. For files with
// multiple tabs (e.g., Python, Node.js, Java tabs), all contexts are collected
// and determineProduct uses the first matching one for ALL code examples.
//
// PRACTICAL IMPACT (audited January 2026):
// This limitation has effectively zero practical impact because:
//
//  1. Code blocks always have explicit language - An audit of ~4,000 files with
//     :tabid: directives found ZERO code blocks without explicit :language: inside
//     driver tabs. Writers consistently specify the language, which takes precedence
//     over tab context in determineProduct.
//
//  2. Literalinclude uses file extension - The 673 literalinclude directives found
//     inside driver tabs all have file extensions that correctly identify the
//     language (.py, .java, .js, etc.), so tab context is not needed.
//
//  3. JSON files are correctly identified - The only "mismatches" (19 cases) are
//     intentional JSON data files shown across multiple driver tabs, which should
//     be attributed to "JSON", not the driver.
//
// A more accurate implementation would track line ranges for each context block
// and match code examples to their containing block, but this adds significant
// complexity for zero practical benefit given the above findings.
func parseFileContexts(filePath string) ([]CodeContext, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var contexts []CodeContext
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for tab directive
		if rst.TabDirectiveRegex.MatchString(trimmedLine) {
			// Look for :tabid: on next lines
			for scanner.Scan() {
				nextLine := scanner.Text()
				if strings.TrimSpace(nextLine) == "" {
					break
				}
				if tabIDMatches := rst.TabIDOptionRegex.FindStringSubmatch(nextLine); len(tabIDMatches) > 1 {
					contexts = append(contexts, CodeContext{TabID: strings.TrimSpace(tabIDMatches[1])})
					break
				}
				// If not an option line, stop looking
				if !strings.HasPrefix(strings.TrimSpace(nextLine), ":") {
					break
				}
			}
		}

		// Check for composable-tutorial
		if rst.ComposableTutorialDirectiveRegex.MatchString(trimmedLine) {
			// Look for :options: on next lines
			for scanner.Scan() {
				nextLine := scanner.Text()
				if strings.TrimSpace(nextLine) == "" {
					break
				}
				if optMatches := rst.OptionsOptionRegex.FindStringSubmatch(nextLine); len(optMatches) > 1 {
					ctx := parseComposableOptions(optMatches[1])
					contexts = append(contexts, ctx)
					break
				}
				if !strings.HasPrefix(strings.TrimSpace(nextLine), ":") {
					break
				}
			}
		}
	}

	if len(contexts) == 0 {
		contexts = append(contexts, CodeContext{})
	}

	return contexts, scanner.Err()
}

// parseComposableOptions parses composable options string like "language=python; interface=driver".
//
// WHY WE CHECK MULTIPLE COMPOSABLE IDS:
// The MongoDB documentation uses several variants of language and interface composables,
// each with slightly different option sets for different contexts. Writers can add new
// composable definitions to their snooty.toml files at any time.
//
// PATTERN MATCHING STRATEGY:
// Rather than maintaining an exhaustive list of composable IDs, we use pattern matching:
//   - Any key containing "language" or "lang" → treated as language context
//   - Any key containing "interface" → treated as interface context
//
// This approach is future-proof: new composables following the naming convention
// (e.g., "language-new-variant", "my-language-selector", "interface-v2") will be
// automatically handled without code changes.
//
// KNOWN COMPOSABLES (audited January 2026, expected to grow):
//
// Language-like composables (68 total usages):
//   - language, language-no-dependencies, language-mongocryptd-only,
//   - language-no-interface, language-atlas-only, language-atlas-only-2,
//   - language-local-only, driver-language
//
// Interface-like composables (65 total usages):
//   - interface, interface-atlas-only, interface-default-atlas-cli,
//   - interface-local-only
//
// All these composables use the same option values (python, nodejs, mongosh, etc.),
// which are mapped to products via ProductMappings loaded from rstspec.toml.
func parseComposableOptions(options string) CodeContext {
	ctx := CodeContext{}
	parts := strings.Split(options, ";")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		// Handle all language-like composables using substring matching.
		// This catches: language, language-*, *-language, *-language-*, driver-lang, etc.
		if strings.Contains(key, "language") || strings.Contains(key, "lang") {
			ctx.Language = value
		}
		// Handle all interface-like composables using substring matching.
		// This catches: interface, interface-*, *-interface, *-interface-*, etc.
		if strings.Contains(key, "interface") {
			ctx.Interface = value
		}
	}
	return ctx
}
