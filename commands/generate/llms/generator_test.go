package llms

import (
	"strings"
	"testing"
)

func TestToMarkdownURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		isRootIndex bool
		want        string
	}{
		{
			name: "regular page",
			url:  "https://www.mongodb.com/docs/manual/core/document/",
			want: "https://www.mongodb.com/docs/manual/core/document.md",
		},
		{
			name: "nested section index",
			url:  "https://www.mongodb.com/docs/manual/crud/",
			want: "https://www.mongodb.com/docs/manual/crud.md",
		},
		{
			name:        "root landing page uses index.md",
			url:         "https://www.mongodb.com/docs/manual/",
			isRootIndex: true,
			want:        "https://www.mongodb.com/docs/manual/index.md",
		},
		{
			name:        "versioned root landing page",
			url:         "https://www.mongodb.com/docs/atlas/cli/current/",
			isRootIndex: true,
			want:        "https://www.mongodb.com/docs/atlas/cli/current/index.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toMarkdownURL(tt.url, tt.isRootIndex)
			if got != tt.want {
				t.Errorf("toMarkdownURL(%q, %v) = %q, want %q", tt.url, tt.isRootIndex, got, tt.want)
			}
		})
	}
}

func TestRenderContent(t *testing.T) {
	pages := []PageEntry{
		{Title: "Documents", URL: "https://ex.com/a.md", Description: "About documents."},
		{Title: "No Desc", URL: "https://ex.com/b.md", Description: ""},
	}

	withDesc := renderContent("manual", pages, true)
	if !strings.Contains(withDesc, "- [Documents](https://ex.com/a.md): About documents.") {
		t.Errorf("expected description line, got:\n%s", withDesc)
	}
	// A page without a description must not emit a trailing ": ".
	if !strings.Contains(withDesc, "- [No Desc](https://ex.com/b.md)\n") ||
		strings.Contains(withDesc, "- [No Desc](https://ex.com/b.md):") {
		t.Errorf("page without description should have no trailing colon, got:\n%s", withDesc)
	}

	noDesc := renderContent("manual", pages, false)
	if strings.Contains(noDesc, "About documents.") {
		t.Errorf("descriptions should be omitted, got:\n%s", noDesc)
	}
	if !strings.HasPrefix(noDesc, "# manual\n\n") {
		t.Errorf("expected project header, got:\n%s", noDesc)
	}
}
