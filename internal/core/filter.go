package core

import (
	"sort"
	"strings"
)

func FilterDirs(dirs []DirEntry, query string) []DirEntry {
	if query == "" {
		return dirs
	}

	query = strings.ToLower(query)
	ranked := rankMatches(dirs, func(d DirEntry) int {
		name := strings.ToLower(d.Name)
		path := strings.ToLower(d.Path)
		return bestScore(
			matchScore(name, query, true),
			matchScore(path, query, false),
		)
	})

	return ranked
}

func FilterWorktrees(wts []Worktree, query string) []Worktree {
	if query == "" {
		return wts
	}

	query = strings.ToLower(query)
	querySanitized := strings.ToLower(SanitizeWorktreeName(query))
	ranked := rankMatches(wts, func(wt Worktree) int {
		name := strings.ToLower(wt.Name)
		branch := strings.ToLower(wt.Branch)
		branchSanitized := strings.ToLower(SanitizeWorktreeName(wt.Branch))

		score := bestScore(
			matchScore(name, query, true),
			matchScore(branch, query, true),
		)
		if querySanitized != "" {
			score = bestScore(score, matchScore(branchSanitized, querySanitized, true))
		}
		return score
	})

	return ranked
}

func FilterTools(tools []string, query string) []string {
	if query == "" {
		return tools
	}

	query = strings.ToLower(query)
	ranked := rankMatches(tools, func(tool string) int {
		name := strings.ToLower(tool)
		return matchScore(name, query, true)
	})

	return ranked
}

func FilterSessions(sessions []SessionInfo, query string) []SessionInfo {
	if query == "" {
		return sessions
	}

	query = strings.ToLower(query)
	ranked := rankMatches(sessions, func(session SessionInfo) int {
		name := strings.ToLower(session.Name)
		path := strings.ToLower(session.DirPath)
		tool := strings.ToLower(session.Tool)
		return bestScore(
			matchScore(name, query, true),
			matchScore(path, query, false),
			matchScore(tool, query, false),
		)
	})

	return ranked
}

const noMatchScore = -1

func matchScore(text, pattern string, preferExactPrefix bool) int {
	if text == "" || pattern == "" {
		return noMatchScore
	}

	if text == pattern {
		if preferExactPrefix {
			return 1_000_000
		}
		return 950_000
	}
	if strings.HasPrefix(text, pattern) {
		if preferExactPrefix {
			return 900_000
		}
		return 850_000
	}

	base := fuzzySubsequenceScore(text, pattern)
	if base == noMatchScore {
		return noMatchScore
	}

	return base
}

func fuzzySubsequenceScore(text, pattern string) int {
	pi := 0
	lastMatch := -2
	score := 0

	for ti := 0; ti < len(text) && pi < len(pattern); ti++ {
		if text[ti] != pattern[pi] {
			continue
		}

		score += 100
		if lastMatch+1 == ti {
			score += 35
		}
		if ti == pi {
			score += 20
		}
		if ti == 0 {
			score += 15
		}
		lastMatch = ti
		pi++
	}

	if pi != len(pattern) {
		return noMatchScore
	}

	gapPenalty := len(text) - len(pattern)
	if gapPenalty > 0 {
		score -= gapPenalty
	}
	return score
}

func bestScore(scores ...int) int {
	best := noMatchScore
	for _, score := range scores {
		if score > best {
			best = score
		}
	}
	return best
}

func rankMatches[T any](items []T, scorer func(T) int) []T {
	type rankedItem struct {
		item  T
		idx   int
		score int
	}

	ranked := make([]rankedItem, 0, len(items))
	for idx, item := range items {
		score := scorer(item)
		if score == noMatchScore {
			continue
		}
		ranked = append(ranked, rankedItem{item: item, idx: idx, score: score})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})

	result := make([]T, 0, len(ranked))
	for _, entry := range ranked {
		result = append(result, entry.item)
	}
	return result
}
