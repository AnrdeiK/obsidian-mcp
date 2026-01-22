package vault

import (
	"reflect"
	"sort"
	"testing"
)

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "no tags",
			content: "This is a note without tags",
			want:    []string{},
		},
		{
			name:    "single tag",
			content: "This note has a #tag",
			want:    []string{"tag"},
		},
		{
			name:    "multiple tags",
			content: "Multiple #tags in #one note with #several tags",
			want:    []string{"tags", "one", "several"},
		},
		{
			name:    "duplicate tags",
			content: "#golang is great, I love #golang and #go",
			want:    []string{"golang", "go"},
		},
		{
			name:    "tags with different cases",
			content: "#GoLang #GOLANG #golang should be deduplicated",
			want:    []string{"golang"}, // All normalized to lowercase
		},
		{
			name:    "tags with numbers",
			content: "#tag1 #tag2 #2024",
			want:    []string{"tag1", "tag2", "2024"},
		},
		{
			name:    "tags at start of line",
			content: "#start\n#middle text\n#end",
			want:    []string{"start", "middle", "end"},
		},
		{
			name:    "tags with special chars (should not match)",
			content: "#tag-with-dash #tag_with_underscore",
			want:    []string{"tag", "tag_with_underscore"},
		},
		{
			name:    "empty string",
			content: "",
			want:    []string{},
		},
		{
			name:    "hashtag in code block",
			content: "```\n#include <stdio.h>\n```\n#actualtag",
			want:    []string{"include", "actualtag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTags(tt.content)

			// Sort both slices for comparison since order doesn't matter
			sort.Strings(got)
			sort.Strings(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTagsPerformance(t *testing.T) {
	// Generate large content with many tags
	content := ""
	for i := 0; i < 1000; i++ {
		content += "#tag" + string(rune(i%100)) + " "
	}

	tags := ExtractTags(content)
	if len(tags) == 0 {
		t.Error("Expected to extract tags from large content")
	}
}
