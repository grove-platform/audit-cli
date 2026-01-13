package rst

import (
	"testing"

	"github.com/grove-platform/audit-cli/internal/language"
)

func TestDirective_ResolveLanguage(t *testing.T) {
	tests := []struct {
		name      string
		directive Directive
		want      string
	}{
		{
			name: "code-block with language argument",
			directive: Directive{
				Type:     CodeBlock,
				Argument: "python",
				Options:  map[string]string{},
			},
			want: "python",
		},
		{
			name: "code-block with language option",
			directive: Directive{
				Type:     CodeBlock,
				Argument: "",
				Options:  map[string]string{"language": "javascript"},
			},
			want: "javascript",
		},
		{
			name: "code-block argument takes priority over option",
			directive: Directive{
				Type:     CodeBlock,
				Argument: "python",
				Options:  map[string]string{"language": "javascript"},
			},
			want: "python",
		},
		{
			name: "code-block with no language returns undefined",
			directive: Directive{
				Type:     CodeBlock,
				Argument: "",
				Options:  map[string]string{},
			},
			want: language.Undefined,
		},
		{
			name: "literalinclude with language option",
			directive: Directive{
				Type:     LiteralInclude,
				Argument: "/path/to/file.txt",
				Options:  map[string]string{"language": "python"},
			},
			want: "python",
		},
		{
			name: "literalinclude infers from file extension",
			directive: Directive{
				Type:     LiteralInclude,
				Argument: "/path/to/example.py",
				Options:  map[string]string{},
			},
			want: "python",
		},
		{
			name: "literalinclude language option takes priority over extension",
			directive: Directive{
				Type:     LiteralInclude,
				Argument: "/path/to/example.py",
				Options:  map[string]string{"language": "javascript"},
			},
			want: "javascript",
		},
		{
			name: "literalinclude with unknown extension returns undefined",
			directive: Directive{
				Type:     LiteralInclude,
				Argument: "/path/to/file.xyz",
				Options:  map[string]string{},
			},
			want: language.Undefined,
		},
		{
			name: "io-code-block with language option",
			directive: Directive{
				Type:     IoCodeBlock,
				Argument: "",
				Options:  map[string]string{"language": "go"},
			},
			want: "go",
		},
		{
			name: "io-code-block with no language returns undefined",
			directive: Directive{
				Type:     IoCodeBlock,
				Argument: "",
				Options:  map[string]string{},
			},
			want: language.Undefined,
		},
		{
			name: "code-block normalizes language",
			directive: Directive{
				Type:     CodeBlock,
				Argument: "ts",
				Options:  map[string]string{},
			},
			want: "typescript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.directive.ResolveLanguage()
			if got != tt.want {
				t.Errorf("Directive.ResolveLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSubDirective_ResolveLanguage(t *testing.T) {
	tests := []struct {
		name          string
		subDir        SubDirective
		parentOptions map[string]string
		want          string
	}{
		{
			name: "sub-directive with own language option",
			subDir: SubDirective{
				Argument: "",
				Options:  map[string]string{"language": "python"},
				Content:  "print('hello')",
			},
			parentOptions: map[string]string{},
			want:          "python",
		},
		{
			name: "sub-directive infers from filepath",
			subDir: SubDirective{
				Argument: "/path/to/example.js",
				Options:  map[string]string{},
				Content:  "",
			},
			parentOptions: map[string]string{},
			want:          "javascript",
		},
		{
			name: "sub-directive language option takes priority over filepath",
			subDir: SubDirective{
				Argument: "/path/to/example.py",
				Options:  map[string]string{"language": "javascript"},
				Content:  "",
			},
			parentOptions: map[string]string{},
			want:          "javascript",
		},
		{
			name: "sub-directive falls back to parent language",
			subDir: SubDirective{
				Argument: "",
				Options:  map[string]string{},
				Content:  "some code",
			},
			parentOptions: map[string]string{"language": "go"},
			want:          "go",
		},
		{
			name: "sub-directive own language takes priority over parent",
			subDir: SubDirective{
				Argument: "",
				Options:  map[string]string{"language": "python"},
				Content:  "print('hello')",
			},
			parentOptions: map[string]string{"language": "go"},
			want:          "python",
		},
		{
			name: "sub-directive with no language returns undefined",
			subDir: SubDirective{
				Argument: "",
				Options:  map[string]string{},
				Content:  "some code",
			},
			parentOptions: map[string]string{},
			want:          language.Undefined,
		},
		{
			name: "sub-directive with nil parent options",
			subDir: SubDirective{
				Argument: "",
				Options:  map[string]string{},
				Content:  "some code",
			},
			parentOptions: nil,
			want:          language.Undefined,
		},
		{
			name: "sub-directive normalizes language",
			subDir: SubDirective{
				Argument: "",
				Options:  map[string]string{"language": "ts"},
				Content:  "const x = 1",
			},
			parentOptions: map[string]string{},
			want:          "typescript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.subDir.ResolveLanguage(tt.parentOptions)
			if got != tt.want {
				t.Errorf("SubDirective.ResolveLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}

