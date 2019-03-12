package markdown

import (
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "should render markdown properly",
			text:     "*List*\n\n* Foo\n* Bar",
			expected: "<p><em>List</em></p>\n\n<ul>\n<li>Foo<br /></li>\n<li>Bar<br /></li>\n</ul>\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRendered := RenderMarkdown(tt.text); gotRendered != tt.expected {
				t.Errorf("RenderMarkdown() = %v, want %v", gotRendered, tt.expected)
			}
		})
	}
}
