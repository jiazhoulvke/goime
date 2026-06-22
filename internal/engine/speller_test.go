package engine

import (
    "reflect"
    "testing"
)

func TestSpellerXiaohe(t *testing.T) {
    s := NewXiaoheSpeller()
    tests := []struct {
        input string
        want  []string
    }{
        {"uuru", []string{"shu", "ru"}},
        {"ni", []string{"ni"}},
        {"hf", []string{"hen"}},
        {"ji", []string{"ji"}},
        {"yc", []string{"yao"}},  // y + iao = yao
        {"ad", []string{"ai"}},   // zero-initial ai
        {"aj", []string{"an"}},   // zero-initial an
        {"ee", []string{"e"}},    // zero-initial e
        {"eg", []string{"eng"}},  // zero-initial eng
        {"oo", []string{"o"}},    // zero-initial o
        {"", nil},
    }
    for _, tc := range tests {
        got := s.ToPinyin(tc.input)
        if !reflect.DeepEqual(got, tc.want) {
            t.Errorf("ToPinyin(%q) = %v, want %v", tc.input, got, tc.want)
        }
    }
}

func TestSpellerFullPinyin(t *testing.T) {
    s := NewFullPinyinSpeller()
    tests := []struct {
        input string
        want  []string
    }{
        {"shuru", []string{"shuru"}},
        {"ni", []string{"ni"}},
        {"hao", []string{"hao"}},
        {"", nil},
    }
    for _, tc := range tests {
        got := s.ToPinyin(tc.input)
        if !reflect.DeepEqual(got, tc.want) {
            t.Errorf("ToPinyin(%q) = %v, want %v", tc.input, got, tc.want)
        }
    }
}

func TestSpellerInterface(t *testing.T) {
    var s Speller = NewXiaoheSpeller()
    if s.Name() != "xiaohe" {
        t.Errorf("expected xiaohe, got %s", s.Name())
    }
    s = NewFullPinyinSpeller()
    if s.Name() != "fullpin" {
        t.Errorf("expected fullpin, got %s", s.Name())
    }
}
