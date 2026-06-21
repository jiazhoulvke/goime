package engine

import (
	"reflect"
	"testing"
)

func TestSegment(t *testing.T) {
	tests := []struct {
		input string
		want  [][]string
	}{
		{"shuru", [][]string{{"shu", "ru"}}},
		{"nihao", [][]string{{"ni", "hao"}}},
		{"xian", [][]string{{"xi", "an"}, {"xian"}}},
		{"fangan", [][]string{{"fang", "an"}, {"fan", "gan"}}},
		{"a", [][]string{{"a"}}},
		{"", [][]string{}},
		{"xx", [][]string{}},
	}
	for _, tc := range tests {
		got := Segment(tc.input)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Segment(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
