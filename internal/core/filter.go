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

func FilterWorktrees(wts []Worktree, query string) []Worktree {
	if query == "" {
		return wts
	}

	query = strings.ToLower(query)
	var result []Worktree

	for _, wt := range wts {
		name := strings.ToLower(wt.Name)
		branch := strings.ToLower(wt.Branch)

		if fuzzyMatch(name, query) || fuzzyMatch(branch, query) {
			result = append(result, wt)
		}
	}

	return result
}

func FilterTools(tools []string, query string) []string {
	if query == "" {
		return tools
	}

	query = strings.ToLower(query)
	var result []string

	for _, tool := range tools {
		name := strings.ToLower(tool)
		if fuzzyMatch(name, query) {
			result = append(result, tool)
		}
	}

	return result
}

func FilterSessions(sessions []SessionInfo, query string) []SessionInfo {
	if query == "" {
		return sessions
	}

	query = strings.ToLower(query)
	var result []SessionInfo

	for _, session := range sessions {
		name := strings.ToLower(session.Name)
		path := strings.ToLower(session.DirPath)
		tool := strings.ToLower(session.Tool)
		if fuzzyMatch(name, query) || fuzzyMatch(path, query) || fuzzyMatch(tool, query) {
			result = append(result, session)
		}
	}

	return result
}
