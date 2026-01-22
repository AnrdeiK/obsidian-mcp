package vault

import (
	"regexp"
	"strings"
)

// tagRegex matches hashtags in markdown content
// Pattern: #(\w+) matches # followed by one or more word characters
var tagRegex = regexp.MustCompile(`#(\w+)`)

// ExtractTags finds all unique tags in the given content
// Tags are identified by the # prefix followed by word characters
// Returns a deduplicated slice of tag names (without the # prefix)
func ExtractTags(content string) []string {
	matches := tagRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return []string{}
	}

	// Use map for deduplication
	tagMap := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			tag := strings.ToLower(match[1]) // Normalize to lowercase
			tagMap[tag] = struct{}{}
		}
	}

	// Convert map to slice
	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags
}
