package str

import (
	"path"
	"strings"
)

// MatchURLPath checks if a URL path matches a pattern
// Supports wildcard matching with *
func MatchURLPath(urlPath, pattern string) bool {
	// Simple implementation for exact match
	if urlPath == pattern {
		return true
	}
	
	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		matched, _ := path.Match(pattern, urlPath)
		return matched
	}
	
	return false
}
