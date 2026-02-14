package core

import (
	"sort"
	"strings"
)

type scoredMatch struct {
	score int
	ok    bool
}

func FilterDirs(dirs []DirEntry, query string) []DirEntry {
	if query == "" {
		return dirs
	}

	query = strings.ToLower(query)
	ranked := rankMatches(dirs, func(d DirEntry) (int, bool) {
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
	ranked := rankMatches(wts, func(wt Worktree) (int, bool) {
		name := strings.ToLower(wt.Name)
		branch := strings.ToLower(wt.Branch)
		branchSanitized := strings.ToLower(SanitizeWorktreeName(wt.Branch))

		score, ok := bestScore(
			matchScore(name, query, true),
			matchScore(branch, query, true),
		)
		if querySanitized != "" {
			score, ok = bestScore(scoredMatch{score: score, ok: ok}, matchScore(branchSanitized, querySanitized, true))
		}
		return score, ok
	})

	return ranked
}

func FilterTools(tools []string, query string) []string {
	if query == "" {
		return tools
	}

	query = strings.ToLower(query)
	ranked := rankMatches(tools, func(tool string) (int, bool) {
		name := strings.ToLower(tool)
		match := matchScore(name, query, true)
		return match.score, match.ok
	})

	return ranked
}

func FilterSessions(sessions []SessionInfo, query string) []SessionInfo {
	if query == "" {
		return sessions
	}

	query = strings.ToLower(query)
	ranked := rankMatches(sessions, func(session SessionInfo) (int, bool) {
		name := strings.ToLower(session.Name)
		path := strings.ToLower(session.DirPath)
		project := strings.ToLower(session.Project)
		branch := strings.ToLower(session.Branch)
		tool := strings.ToLower(session.Tool)
		return bestScore(
			matchScore(name, query, true),
			matchScore(path, query, false),
			matchScore(project, query, false),
			matchScore(branch, query, false),
			matchScore(tool, query, false),
		)
	})

	return ranked
}

func matchScore(text, pattern string, preferExactPrefix bool) scoredMatch {
	if text == "" || pattern == "" {
		return scoredMatch{}
	}

	if text == pattern {
		if preferExactPrefix {
			return scoredMatch{score: 1_000_000, ok: true}
		}
		return scoredMatch{score: 950_000, ok: true}
	}
	if strings.HasPrefix(text, pattern) {
		if preferExactPrefix {
			return scoredMatch{score: 900_000, ok: true}
		}
		return scoredMatch{score: 850_000, ok: true}
	}

	base, ok := fuzzySubsequenceScore(text, pattern)
	if !ok {
		return scoredMatch{}
	}

	return scoredMatch{score: base, ok: true}
}

func fuzzySubsequenceScore(text, pattern string) (int, bool) {
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
		return 0, false
	}

	gapPenalty := len(text) - len(pattern)
	if gapPenalty > 0 {
		score -= gapPenalty
	}
	return score, true
}

func bestScore(scores ...scoredMatch) (int, bool) {
	best := 0
	hasMatch := false
	for _, score := range scores {
		if !score.ok {
			continue
		}
		if !hasMatch || score.score > best {
			best = score.score
			hasMatch = true
		}
	}
	return best, hasMatch
}

func rankMatches[T any](items []T, scorer func(T) (int, bool)) []T {
	type rankedItem struct {
		item  T
		idx   int
		score int
	}

	ranked := make([]rankedItem, 0, len(items))
	for idx, item := range items {
		score, ok := scorer(item)
		if !ok {
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
