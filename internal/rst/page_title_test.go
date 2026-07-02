package rst

import "testing"

func TestExtractPageTitle(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "underline-only heading",
			content: `Documents
=========

Body text.
`,
			want: "Documents",
		},
		{
			name: "overline and underline heading",
			content: `=========
Documents
=========

Body text.
`,
			want: "Documents",
		},
		{
			name: "title after meta directive",
			content: `.. meta::
   :description: A summary.

================================
Rotate Keys for Sharded Clusters
================================
`,
			want: "Rotate Keys for Sharded Clusters",
		},
		{
			name: "skips directives and field lists",
			content: `.. default-domain:: mongodb

My Page Title
=============
`,
			want: "My Page Title",
		},
		{
			name: "no heading",
			content: `.. include:: /includes/foo.rst

Just a paragraph with no heading.
`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.content)
			got, err := ExtractPageTitle(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExtractPageTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}
