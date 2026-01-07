package language

import (
	"testing"
)

func TestGetExtensionFromLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{"python", "python", ".py"},
		{"Python uppercase", "Python", ".py"},
		{"javascript", "javascript", ".js"},
		{"js shorthand", "js", ".js"},
		{"typescript", "typescript", ".ts"},
		{"ts shorthand", "ts", ".ts"},
		{"go", "go", ".go"},
		{"golang alias", "golang", ".go"},
		{"java", "java", ".java"},
		{"csharp", "csharp", ".cs"},
		{"c# alias", "c#", ".cs"},
		{"cs alias", "cs", ".cs"},
		{"cpp", "cpp", ".cpp"},
		{"c++ alias", "c++", ".cpp"},
		{"ruby", "ruby", ".rb"},
		{"rb shorthand", "rb", ".rb"},
		{"rust", "rust", ".rs"},
		{"rs shorthand", "rs", ".rs"},
		{"shell", "shell", ".sh"},
		{"sh shorthand", "sh", ".sh"},
		{"bash", "bash", ".sh"},
		{"json", "json", ".json"},
		{"yaml", "yaml", ".yaml"},
		{"yml alias", "yml", ".yaml"},
		{"text", "text", ".txt"},
		{"txt alias", "txt", ".txt"},
		{"empty string", "", ".txt"},
		{"none", "none", ".txt"},
		{"unknown language", "unknownlang", ".txt"},
		{"whitespace", "  python  ", ".py"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExtensionFromLanguage(tt.language)
			if got != tt.want {
				t.Errorf("GetExtensionFromLanguage(%q) = %q, want %q", tt.language, got, tt.want)
			}
		})
	}
}

func TestGetLanguageFromExtension(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"python file", "example.py", Python},
		{"javascript file", "script.js", JavaScript},
		{"typescript file", "app.ts", TypeScript},
		{"go file", "main.go", Go},
		{"java file", "Main.java", Java},
		{"csharp file", "Program.cs", CSharp},
		{"cpp file", "main.cpp", CPP},
		{"c file", "main.c", C},
		{"ruby file", "script.rb", Ruby},
		{"rust file", "main.rs", Rust},
		{"shell file", "script.sh", Shell},
		{"bash file", "script.bash", Shell},
		{"json file", "config.json", JSON},
		{"yaml file", "config.yaml", YAML},
		{"yml file", "config.yml", YAML},
		{"xml file", "data.xml", XML},
		{"html file", "index.html", HTML},
		{"css file", "styles.css", CSS},
		{"sql file", "query.sql", SQL},
		{"text file", "readme.txt", Text},
		{"php file", "index.php", PHP},
		{"full path", "/path/to/file.py", Python},
		{"unknown extension", "file.xyz", ""},
		{"no extension", "Makefile", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLanguageFromExtension(tt.filePath)
			if got != tt.want {
				t.Errorf("GetLanguageFromExtension(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{"python", "python", Python},
		{"Python uppercase", "Python", Python},
		{"py shorthand", "py", Python},
		{"javascript", "javascript", JavaScript},
		{"js shorthand", "js", JavaScript},
		{"typescript", "typescript", TypeScript},
		{"ts shorthand", "ts", TypeScript},
		{"go", "go", Go},
		{"golang alias", "golang", Go},
		{"csharp", "csharp", CSharp},
		{"c# alias", "c#", CSharp},
		{"cs alias", "cs", CSharp},
		{"cpp", "cpp", CPP},
		{"c++ alias", "c++", CPP},
		{"shell", "shell", Shell},
		{"sh shorthand", "sh", Shell},
		{"yaml", "yaml", YAML},
		{"yml alias", "yml", YAML},
		{"empty string", "", Undefined},
		{"none", "none", Undefined},
		{"unknown language", "unknownlang", "unknownlang"},
		{"whitespace", "  python  ", Python},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.language)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.language, got, tt.want)
			}
		})
	}
}

func TestGetProductFromLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{"python", "python", "Python"},
		{"Python uppercase", "Python", "Python"},
		{"javascript", "javascript", "JavaScript"},
		{"js shorthand", "js", "JavaScript"},
		{"typescript", "typescript", "TypeScript"},
		{"ts shorthand", "ts", "TypeScript"},
		{"go", "go", "Go"},
		{"golang alias", "golang", "Go"},
		{"java", "java", "Java"},
		{"csharp", "csharp", "C#"},
		{"c# alias", "c#", "C#"},
		{"mongosh", "mongosh", "MongoDB Shell"},
		{"bash", "bash", "Shell"},
		{"sh", "sh", "Shell"},
		{"shell", "shell", "Shell"},
		{"console", "console", "Shell"},
		{"json", "json", "JSON"},
		{"yaml", "yaml", "YAML"},
		{"yml alias", "yml", "YAML"},
		{"unknown returns original", "unknownlang", "unknownlang"},
		{"whitespace trimmed", "  python  ", "Python"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetProductFromLanguage(tt.language)
			if got != tt.want {
				t.Errorf("GetProductFromLanguage(%q) = %q, want %q", tt.language, got, tt.want)
			}
		})
	}
}

func TestIsNonDriverLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     bool
	}{
		{"bash is non-driver", "bash", true},
		{"sh is non-driver", "sh", true},
		{"console is non-driver", "console", true},
		{"text is non-driver", "text", true},
		{"json is non-driver", "json", true},
		{"yaml is non-driver", "yaml", true},
		{"xml is non-driver", "xml", true},
		{"ini is non-driver", "ini", true},
		{"toml is non-driver", "toml", true},
		{"properties is non-driver", "properties", true},
		{"sql is non-driver", "sql", true},
		{"none is non-driver", "none", true},
		{"http is non-driver", "http", true},
		{"python is driver", "python", false},
		{"javascript is driver", "javascript", false},
		{"go is driver", "go", false},
		{"java is driver", "java", false},
		{"shell is driver", "shell", false}, // shell has special handling
		{"case insensitive", "BASH", true},
		{"whitespace trimmed", "  json  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNonDriverLanguage(tt.language)
			if got != tt.want {
				t.Errorf("IsNonDriverLanguage(%q) = %v, want %v", tt.language, got, tt.want)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name           string
		languageArg    string
		languageOption string
		filePath       string
		want           string
	}{
		{
			name:           "explicit language argument takes priority",
			languageArg:    "python",
			languageOption: "javascript",
			filePath:       "/path/to/file.go",
			want:           "python",
		},
		{
			name:           "language option when no argument",
			languageArg:    "",
			languageOption: "javascript",
			filePath:       "/path/to/file.py",
			want:           "javascript",
		},
		{
			name:           "infer from filepath when no explicit language",
			languageArg:    "",
			languageOption: "",
			filePath:       "/path/to/example.py",
			want:           "python",
		},
		{
			name:           "infer from .js extension",
			languageArg:    "",
			languageOption: "",
			filePath:       "code/snippet.js",
			want:           "javascript",
		},
		{
			name:           "infer from .go extension",
			languageArg:    "",
			languageOption: "",
			filePath:       "main.go",
			want:           "go",
		},
		{
			name:           "infer from .java extension",
			languageArg:    "",
			languageOption: "",
			filePath:       "Example.java",
			want:           "java",
		},
		{
			name:           "unknown extension returns undefined",
			languageArg:    "",
			languageOption: "",
			filePath:       "/path/to/file.xyz",
			want:           "undefined",
		},
		{
			name:           "no inputs returns undefined",
			languageArg:    "",
			languageOption: "",
			filePath:       "",
			want:           "undefined",
		},
		{
			name:           "language argument normalized",
			languageArg:    "ts",
			languageOption: "",
			filePath:       "",
			want:           "typescript",
		},
		{
			name:           "language option normalized",
			languageArg:    "",
			languageOption: "ts",
			filePath:       "",
			want:           "typescript",
		},
		{
			name:           "filepath extension normalized",
			languageArg:    "",
			languageOption: "",
			filePath:       "/path/to/file.yml",
			want:           "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.languageArg, tt.languageOption, tt.filePath)
			if got != tt.want {
				t.Errorf("Resolve(%q, %q, %q) = %q, want %q",
					tt.languageArg, tt.languageOption, tt.filePath, got, tt.want)
			}
		})
	}
}

func TestIsMongoShellLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     bool
	}{
		{"shell is mongo shell", "shell", true},
		{"javascript is mongo shell", "javascript", true},
		{"js is mongo shell", "js", true},
		{"python is not mongo shell", "python", false},
		{"bash is not mongo shell", "bash", false},
		{"mongosh is not in list", "mongosh", false}, // mongosh is handled separately
		{"case insensitive", "SHELL", true},
		{"case insensitive js", "JavaScript", true},
		{"whitespace trimmed", "  shell  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMongoShellLanguage(tt.language)
			if got != tt.want {
				t.Errorf("IsMongoShellLanguage(%q) = %v, want %v", tt.language, got, tt.want)
			}
		})
	}
}

