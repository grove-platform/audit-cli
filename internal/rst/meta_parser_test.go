package rst

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "page.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestExtractMetaDescription(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "description present with other options",
			content: `.. meta::
   :robots: noindex, nosnippet
   :description: Definition and structure of documents.

====
Docs
====
`,
			want: "Definition and structure of documents.",
		},
		{
			name: "description only",
			content: `.. meta::
   :description: A short summary.

Title
=====
`,
			want: "A short summary.",
		},
		{
			name: "multi-line description is joined",
			content: `.. meta::
   :description: This description wraps
      across multiple lines.

Title
=====
`,
			want: "This description wraps across multiple lines.",
		},
		{
			name: "no meta directive",
			content: `Title
=====

Some content.
`,
			want: "",
		},
		{
			name: "meta without description",
			content: `.. meta::
   :robots: noindex

Title
=====
`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.content)
			got, err := ExtractMetaDescription(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExtractMetaDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}
