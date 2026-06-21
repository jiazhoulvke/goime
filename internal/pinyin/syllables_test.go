package pinyin

import "testing"

func TestIsValidSyllable(t *testing.T) {
	tests := []struct {
		syllable string
		valid    bool
	}{
		{"shu", true},
		{"ru", true},
		{"ni", true},
		{"hao", true},
		{"zhuang", true},
		{"x", false},
		{"sh", false},
		{"", false},
		{"abcde", false},
	}
	for _, tc := range tests {
		got := IsValidSyllable(tc.syllable)
		if got != tc.valid {
			t.Errorf("IsValidSyllable(%q) = %v, want %v", tc.syllable, got, tc.valid)
		}
	}
}

func TestAllSyllablesCount(t *testing.T) {
	n := len(AllSyllables())
	if n < 400 || n > 420 {
		t.Errorf("expected ~410 syllables, got %d", n)
	}
}
