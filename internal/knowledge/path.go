package knowledge

import (
	"fmt"
	"path"
	"strings"
)

// NormalizeFile accepts repository-relative paths even when they do not exist.
func NormalizeFile(value string) (string, error) {
	normalized := strings.ReplaceAll(value, `\`, "/")
	if strings.TrimSpace(normalized) == "" || strings.HasPrefix(normalized, "/") || strings.Contains(normalized, ":") || strings.ContainsRune(normalized, 0) {
		return "", fmt.Errorf("file path %q must be repository-relative", value)
	}
	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return "", fmt.Errorf("file path %q contains traversal", value)
		}
	}
	normalized = path.Clean(normalized)
	if normalized == "." {
		return "", fmt.Errorf("file path %q must name a file", value)
	}
	return normalized, nil
}

type Pattern struct{ segments []string }

func ParsePattern(value string) (Pattern, error) {
	if value == "" || strings.ContainsAny(value, `\:[]{}!`) || strings.HasPrefix(value, "/") || strings.ContainsRune(value, 0) {
		return Pattern{}, fmt.Errorf("invalid path pattern %q", value)
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || (part != "**" && strings.Contains(part, "**")) {
			return Pattern{}, fmt.Errorf("invalid path pattern %q", value)
		}
	}
	return Pattern{parts}, nil
}

// Match uses memoized segment matching, independent of host filesystem semantics.
func (p Pattern) Match(file string) bool {
	parts := strings.Split(file, "/")
	type state struct{ p, f int }
	cache := map[state]bool{}
	seen := map[state]bool{}
	var match func(int, int) bool
	match = func(pi, fi int) bool {
		s := state{pi, fi}
		if seen[s] {
			return cache[s]
		}
		seen[s] = true
		result := false
		if pi == len(p.segments) {
			result = fi == len(parts)
		} else if p.segments[pi] == "**" {
			result = match(pi+1, fi) || (fi < len(parts) && match(pi, fi+1))
		} else if fi < len(parts) {
			result = segmentMatch(p.segments[pi], parts[fi]) && match(pi+1, fi+1)
		}
		cache[s] = result
		return result
	}
	return match(0, 0)
}

func segmentMatch(pattern, value string) bool {
	p, v := []rune(pattern), []rune(value)
	previous := make([]bool, len(v)+1)
	previous[0] = true
	for _, char := range p {
		next := make([]bool, len(v)+1)
		if char == '*' {
			next[0] = previous[0]
		}
		for j := 1; j <= len(v); j++ {
			if char == '*' {
				next[j] = previous[j] || next[j-1]
			} else {
				next[j] = previous[j-1] && (char == '?' || char == v[j-1])
			}
		}
		previous = next
	}
	return previous[len(v)]
}
