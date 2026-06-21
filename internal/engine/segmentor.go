package engine

import (
	"sort"

	"github.com/jiazhoulvke/goime/internal/pinyin"
)

// excludedSyllables contains syllables that are technically valid pinyin
// (like interjections "o" and "ng") but should not appear as standalone
// syllables in normal pinyin segmentation.
var excludedSyllables = map[string]struct{}{
	"o":  {},
	"ng": {},
}

// isSegmentableSyllable returns true if s is a valid pinyin syllable
// that should be considered in segmentation.
func isSegmentableSyllable(s string) bool {
	if _, excluded := excludedSyllables[s]; excluded {
		return false
	}
	return pinyin.IsValidSyllable(s)
}

// Segment 将拼音字符串按所有合法音节边界切分
func Segment(s string) [][]string {
	if s == "" {
		return [][]string{}
	}
	result := [][]string{}
	backtrack(s, 0, nil, &result)
	// Sort results: more syllables first, then longer first syllable first
	sort.Slice(result, func(i, j int) bool {
		if len(result[i]) != len(result[j]) {
			return len(result[i]) > len(result[j])
		}
		for k := 0; k < len(result[i]) && k < len(result[j]); k++ {
			if len(result[i][k]) != len(result[j][k]) {
				return len(result[i][k]) > len(result[j][k])
			}
		}
		return false
	})
	return result
}

func backtrack(s string, start int, cur []string, result *[][]string) {
	if start >= len(s) {
		seg := make([]string, len(cur))
		copy(seg, cur)
		*result = append(*result, seg)
		return
	}
	for end := start + 1; end <= len(s) && end-start <= 6; end++ {
		syl := s[start:end]
		if isSegmentableSyllable(syl) {
			backtrack(s, end, append(cur, syl), result)
		}
	}
}
