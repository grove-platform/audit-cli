// Package language provides utilities for working with programming language identifiers.
//
// This package provides:
//   - Canonical language name constants
//   - File extension constants and mappings
//   - Language normalization (e.g., "ts" -> "typescript")
//   - Language inference from file extensions
package language

import (
	"path/filepath"
	"strings"
)

// Language constants define canonical language names used throughout the tool.
// These are used for normalization and file extension mapping.
const (
	Bash       = "bash"
	C          = "c"
	CPP        = "cpp"
	CSharp     = "csharp"
	Console    = "console"
	CSS        = "css"
	Go         = "go"
	HTML       = "html"
	Java       = "java"
	JavaScript = "javascript"
	JSON       = "json"
	Kotlin     = "kotlin"
	PHP        = "php"
	PowerShell = "powershell"
	PS5        = "ps5"
	Python     = "python"
	Ruby       = "ruby"
	Rust       = "rust"
	Scala      = "scala"
	Shell      = "shell"
	SQL        = "sql"
	Swift      = "swift"
	Text       = "text"
	TypeScript = "typescript"
	Undefined  = "undefined"
	XML        = "xml"
	YAML       = "yaml"
)

// File extension constants define the file extensions for each language.
// Used when generating output filenames for extracted code examples.
const (
	BashExtension       = ".sh"
	CExtension          = ".c"
	CPPExtension        = ".cpp"
	CSharpExtension     = ".cs"
	ConsoleExtension    = ".sh"
	CSSExtension        = ".css"
	GoExtension         = ".go"
	HTMLExtension       = ".html"
	JavaExtension       = ".java"
	JavaScriptExtension = ".js"
	JSONExtension       = ".json"
	KotlinExtension     = ".kt"
	PHPExtension        = ".php"
	PowerShellExtension = ".ps1"
	PS5Extension        = ".ps1"
	PythonExtension     = ".py"
	RubyExtension       = ".rb"
	RustExtension       = ".rs"
	ScalaExtension      = ".scala"
	ShellExtension      = ".sh"
	SQLExtension        = ".sql"
	SwiftExtension      = ".swift"
	TextExtension       = ".txt"
	TypeScriptExtension = ".ts"
	UndefinedExtension  = ".txt"
	XMLExtension        = ".xml"
	YAMLExtension       = ".yaml"
)

// GetExtensionFromLanguage returns the appropriate file extension for a given language.
//
// This function maps language identifiers to their corresponding file extensions.
// Handles various language name variants (e.g., "ts" -> ".ts", "c++" -> ".cpp", "golang" -> ".go").
// Returns ".txt" for unknown or undefined languages.
//
// Parameters:
//   - language: The language identifier (case-insensitive)
//
// Returns:
//   - string: The file extension including the leading dot (e.g., ".js", ".py")
func GetExtensionFromLanguage(language string) string {
	lang := strings.ToLower(strings.TrimSpace(language))

	langExtensionMap := map[string]string{
		Bash:       BashExtension,
		C:          CExtension,
		CPP:        CPPExtension,
		CSharp:     CSharpExtension,
		Console:    ConsoleExtension,
		CSS:        CSSExtension,
		Go:         GoExtension,
		HTML:       HTMLExtension,
		Java:       JavaExtension,
		JavaScript: JavaScriptExtension,
		JSON:       JSONExtension,
		Kotlin:     KotlinExtension,
		PHP:        PHPExtension,
		PowerShell: PowerShellExtension,
		PS5:        PS5Extension,
		Python:     PythonExtension,
		Ruby:       RubyExtension,
		Rust:       RustExtension,
		Scala:      ScalaExtension,
		Shell:      ShellExtension,
		SQL:        SQLExtension,
		Swift:      SwiftExtension,
		Text:       TextExtension,
		TypeScript: TypeScriptExtension,
		Undefined:  UndefinedExtension,
		XML:        XMLExtension,
		YAML:       YAMLExtension,
		"c++":      CPPExtension,
		"c#":       CSharpExtension,
		"cs":       CSharpExtension,
		"golang":   GoExtension,
		"js":       JavaScriptExtension,
		"kt":       KotlinExtension,
		"py":       PythonExtension,
		"rb":       RubyExtension,
		"rs":       RustExtension,
		"sh":       ShellExtension,
		"ts":       TypeScriptExtension,
		"txt":      TextExtension,
		"ps1":      PowerShellExtension,
		"yml":      YAMLExtension,
		"":         UndefinedExtension,
		"none":     UndefinedExtension,
	}

	if extension, exists := langExtensionMap[lang]; exists {
		return extension
	}

	return UndefinedExtension
}

// GetLanguageFromExtension infers the language from a file extension.
//
// This function maps file extensions to their corresponding language names.
// Returns empty string if the extension is not recognized.
//
// Parameters:
//   - filePath: The file path to extract the extension from
//
// Returns:
//   - string: The language name, or empty string if not recognized
func GetLanguageFromExtension(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	extensionMap := map[string]string{
		".py":    Python,
		".js":    JavaScript,
		".ts":    TypeScript,
		".go":    Go,
		".java":  Java,
		".cs":    CSharp,
		".cpp":   CPP,
		".c":     C,
		".rb":    Ruby,
		".rs":    Rust,
		".swift": Swift,
		".kt":    Kotlin,
		".scala": Scala,
		".sh":    Shell,
		".bash":  Shell,
		".ps1":   PowerShell,
		".json":  JSON,
		".yaml":  YAML,
		".yml":   YAML,
		".xml":   XML,
		".html":  HTML,
		".css":   CSS,
		".sql":   SQL,
		".txt":   Text,
		".php":   PHP,
	}
	if lang, ok := extensionMap[ext]; ok {
		return lang
	}
	return ""
}

// Resolve determines the language for a code example using a fallback chain.
//
// This function implements a priority-based language detection:
//  1. languageArg - explicit language from directive argument (e.g., .. code-block:: python)
//  2. languageOption - explicit language from :language: option
//  3. filePath - infer from file extension (for literalinclude, io-code-block)
//  4. "undefined" as final fallback
//
// The result is normalized before being returned.
//
// Parameters:
//   - languageArg: Language from directive argument (empty if argument is a filepath)
//   - languageOption: The value of the :language: option (may be empty)
//   - filePath: The filepath to infer language from extension (may be empty)
//
// Returns:
//   - string: The normalized language name
func Resolve(languageArg, languageOption, filePath string) string {
	// Priority 1: explicit language argument
	lang := languageArg

	// Priority 2: :language: option
	if lang == "" {
		lang = languageOption
	}

	// Priority 3: infer from file extension
	if lang == "" && filePath != "" {
		lang = GetLanguageFromExtension(filePath)
	}

	// Final fallback to undefined
	if lang == "" {
		lang = Undefined
	}

	return Normalize(lang)
}

// Normalize normalizes a language string to a canonical form.
//
// This function converts various language name variants to their canonical forms:
//   - "ts" -> "typescript"
//   - "c++" -> "cpp"
//   - "golang" -> "go"
//   - "js" -> "javascript"
//   - etc.
//
// Parameters:
//   - language: The language identifier (case-insensitive)
//
// Returns:
//   - string: The normalized language name, or the original string if no normalization is defined
func Normalize(language string) string {
	lang := strings.ToLower(strings.TrimSpace(language))

	normalizeMap := map[string]string{
		Bash:       Bash,
		C:          C,
		CPP:        CPP,
		CSharp:     CSharp,
		Console:    Console,
		CSS:        CSS,
		Go:         Go,
		HTML:       HTML,
		Java:       Java,
		JavaScript: JavaScript,
		JSON:       JSON,
		Kotlin:     Kotlin,
		PHP:        PHP,
		PowerShell: PowerShell,
		PS5:        PS5,
		Python:     Python,
		Ruby:       Ruby,
		Rust:       Rust,
		Scala:      Scala,
		Shell:      Shell,
		SQL:        SQL,
		Swift:      Swift,
		Text:       Text,
		TypeScript: TypeScript,
		XML:        XML,
		YAML:       YAML,
		"c++":      CPP,
		"c#":       CSharp,
		"cs":       CSharp,
		"golang":   Go,
		"js":       JavaScript,
		"kt":       Kotlin,
		"py":       Python,
		"rb":       Ruby,
		"rs":       Rust,
		"sh":       Shell,
		"ts":       TypeScript,
		"txt":      Text,
		"ps1":      PowerShell,
		"yml":      YAML,
		"":         Undefined,
		"none":     Undefined,
	}

	if normalized, exists := normalizeMap[lang]; exists {
		return normalized
	}

	return lang
}

// LanguageToProduct maps language identifiers to their display product names.
// This is used for reporting purposes when a language needs to be displayed
// as a product category.
var LanguageToProduct = map[string]string{
	"python":     "Python",
	"javascript": "JavaScript",
	"js":         "JavaScript",
	"typescript": "TypeScript",
	"ts":         "TypeScript",
	"go":         "Go",
	"golang":     "Go",
	"java":       "Java",
	"csharp":     "C#",
	"c#":         "C#",
	"cs":         "C#",
	"cpp":        "C++",
	"c++":        "C++",
	"c":          "C",
	"ruby":       "Ruby",
	"rb":         "Ruby",
	"rust":       "Rust",
	"rs":         "Rust",
	"swift":      "Swift",
	"kotlin":     "Kotlin",
	"kt":         "Kotlin",
	"scala":      "Scala",
	"php":        "PHP",
	"mongosh":    "MongoDB Shell",
	"bash":       "Shell",
	"sh":         "Shell",
	"shell":      "Shell",
	"console":    "Shell",
	"powershell": "PowerShell",
	"ps1":        "PowerShell",
	"json":       "JSON",
	"yaml":       "YAML",
	"yml":        "YAML",
	"xml":        "XML",
	"html":       "HTML",
	"css":        "CSS",
	"sql":        "SQL",
	"ini":        "INI",
	"toml":       "TOML",
	"properties": "Properties",
	"text":       "Text",
	"txt":        "Text",
	"none":       "Text",
}

// GetProductFromLanguage maps a language string to a display product name.
//
// This function converts language identifiers to their display names for reporting:
//   - "python" -> "Python"
//   - "js" -> "JavaScript"
//   - "mongosh" -> "MongoDB Shell"
//   - etc.
//
// Parameters:
//   - lang: The language identifier (case-insensitive)
//
// Returns:
//   - string: The display product name, or the original language if no mapping exists
func GetProductFromLanguage(lang string) string {
	langLower := strings.ToLower(strings.TrimSpace(lang))
	if product, ok := LanguageToProduct[langLower]; ok {
		return product
	}
	return lang
}

// NonDriverLanguages lists languages that should NOT inherit context from
// composables or tabs.
//
// WHY THIS EXISTS:
// Driver documentation often includes code examples that are NOT driver code:
//   - Shell commands to install packages (bash, sh)
//   - Configuration files (json, yaml, xml, ini, toml)
//   - SQL queries for comparison
//   - HTTP requests showing API calls
//
// Without this list, a bash command like "npm install mongodb" inside a Node.js
// driver tab would be incorrectly attributed to "Node.js" and counted as testable.
// By checking this list first, we ensure these examples are reported based on their
// actual language and are NOT considered testable.
//
// WHY "shell" AND "javascript" ARE EXCLUDED:
// These languages have special handling because they CAN be valid MongoDB Shell
// examples when in a MongoDB Shell context. See MongoShellLanguages and the
// special handling in determineProduct().
var NonDriverLanguages = map[string]bool{
	"bash":       true,
	"sh":         true,
	"console":    true,
	"text":       true,
	"json":       true,
	"yaml":       true,
	"xml":        true,
	"ini":        true,
	"toml":       true,
	"properties": true,
	"sql":        true,
	"none":       true,
	"http":       true,
}

// IsNonDriverLanguage checks if a language should NOT inherit context from
// composables or tabs.
//
// Parameters:
//   - language: The language identifier (case-insensitive)
//
// Returns:
//   - bool: true if the language is a non-driver language
func IsNonDriverLanguage(language string) bool {
	return NonDriverLanguages[strings.ToLower(strings.TrimSpace(language))]
}

// MongoShellLanguages lists languages that are valid for MongoDB Shell examples.
//
// WHY THIS EXISTS:
// MongoDB Shell (mongosh) code examples use "shell", "javascript", or "js" as
// their language. However, these same languages are used in other contexts:
//   - "shell" is used for bash/system shell commands
//   - "javascript" is used for browser JavaScript or Node.js
//
// To correctly identify MongoDB Shell examples, we need to check BOTH:
//  1. The language is in this list, AND
//  2. We're in a MongoDB Shell context (mongosh content dir or mongosh interface)
//
// If both conditions are met, the example is attributed to "MongoDB Shell" and
// is testable. Otherwise:
//   - "shell" → "Shell" (not testable, it's a system shell command)
//   - "javascript"/"js" → falls through to driver context or "JavaScript"
var MongoShellLanguages = map[string]bool{
	"shell":      true,
	"javascript": true,
	"js":         true,
}

// IsMongoShellLanguage checks if a language is valid for MongoDB Shell examples.
//
// Parameters:
//   - language: The language identifier (case-insensitive)
//
// Returns:
//   - bool: true if the language could be a MongoDB Shell language
func IsMongoShellLanguage(language string) bool {
	return MongoShellLanguages[strings.ToLower(strings.TrimSpace(language))]
}

