package server

import "testing"

func TestSessionAppendSelection(t *testing.T) {
	s := NewSession("xiaohe")
	s.AppendSelection("nihao", "你好")
	s.AppendSelection("shijie", "世界")
	if len(s.Selections()) != 2 {
		t.Errorf("expected 2 selections, got %d", len(s.Selections()))
	}
}

func TestSessionReset(t *testing.T) {
	s := NewSession("xiaohe")
	s.Append("a")
	s.Append("b")
	if s.Buffer() != "ab" {
		t.Errorf("buffer = %q, want ab", s.Buffer())
	}
	s.Reset()
	if s.Buffer() != "" {
		t.Errorf("buffer should be empty after reset")
	}
}

func TestSessionBackspace(t *testing.T) {
	s := NewSession("xiaohe")
	s.Append("a")
	s.Append("b")
	s.Backspace()
	if s.Buffer() != "a" {
		t.Errorf("buffer = %q, want a", s.Buffer())
	}
}

func TestSessionClear(t *testing.T) {
	s := NewSession("xiaohe")
	s.Append("a")
	s.Clear()
	if s.Buffer() != "" {
		t.Errorf("buffer should be empty after clear")
	}
}

func TestSessionSetScheme(t *testing.T) {
	s := NewSession("xiaohe")
	if s.Scheme() != "xiaohe" {
		t.Errorf("expected xiaohe, got %s", s.Scheme())
	}
	s.SetScheme("fullpin")
	if s.Scheme() != "fullpin" {
		t.Errorf("expected fullpin, got %s", s.Scheme())
	}
}
