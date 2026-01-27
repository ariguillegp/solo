package core

import "strings"

func FilterDirs(dirs []DirEntry, query string) []DirEntry {
	if query == "" {
		return dirs
	}

	query = strings.ToLower(query)
	var result []DirEntry

	for _, d := range dirs {
		name := strings.ToLower(d.Name)
		path := strings.ToLower(d.Path)

		if fuzzyMatch(name, query) || fuzzyMatch(path, query) {
			result = append(result, d)
		}
	}

	return result
}

func fuzzyMatch(text, pattern string) bool {
	pi := 0
	for ti := 0; ti < len(text) && pi < len(pattern); ti++ {
		if text[ti] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}
