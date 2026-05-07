package editor

import (
	"testing"
)

func TestToBuffer(t *testing.T) {
	input := "line1\nline2\nline3"
	buf := toBuffer(input)
	if len(buf) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(buf))
	}
	if string(buf[0]) != "line1" || string(buf[1]) != "line2" || string(buf[2]) != "line3" {
		t.Errorf("Buffer content mismatch: %v", buf)
	}
}

func TestInsertRune(t *testing.T) {
	buf := [][]rune{[]rune("abc")}
	cur := Cursor{0, 1}
	insertRune(&buf, &cur, 'X')
	if string(buf[0]) != "aXbc" {
		t.Errorf("Expected 'aXbc', got '%s'", string(buf[0]))
	}
	if cur.Col != 2 {
		t.Errorf("Expected cursor col 2, got %d", cur.Col)
	}
}

func TestBackspace(t *testing.T) {
	buf := [][]rune{[]rune("abc")}
	cur := Cursor{0, 2}
	backspace(&buf, &cur)
	if string(buf[0]) != "ac" {
		t.Errorf("Expected 'ac', got '%s'", string(buf[0]))
	}
	if cur.Col != 1 {
		t.Errorf("Expected cursor col 1, got %d", cur.Col)
	}
}
