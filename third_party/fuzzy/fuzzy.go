package fuzzy

import "strings"

type Match struct {
	Str            string
	Index          int
	MatchedIndexes []int
	Score          int
}

type Matches []Match

func (m Matches) Len() int      { return len(m) }
func (m Matches) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m Matches) Less(i, j int) bool {
	if m[i].Score == m[j].Score {
		return m[i].Index < m[j].Index
	}
	return m[i].Score > m[j].Score
}

func Find(pattern string, data []string) Matches {
	matches := FindNoSort(pattern, data)
	// simple insertion sort to avoid importing sort in tiny shim
	for i := 1; i < len(matches); i++ {
		j := i
		for j > 0 && matches.Less(j, j-1) {
			matches.Swap(j, j-1)
			j--
		}
	}
	return matches
}

func FindNoSort(pattern string, data []string) Matches {
	pattern = strings.ToLower(pattern)
	if pattern == "" {
		return nil
	}
	results := make(Matches, 0)
	for idx, s := range data {
		ok, pos := subseq(strings.ToLower(s), pattern)
		if ok {
			results = append(results, Match{Str: s, Index: idx, MatchedIndexes: pos, Score: len(pos)})
		}
	}
	return results
}

func subseq(text, pattern string) (bool, []int) {
	pi := 0
	matched := make([]int, 0, len(pattern))
	for ti := 0; ti < len(text) && pi < len(pattern); ti++ {
		if text[ti] == pattern[pi] {
			matched = append(matched, ti)
			pi++
		}
	}
	return pi == len(pattern), matched
}
