package editor

import (
	"testing"
)

// ── existing core tests ───────────────────────────────────────────────────────

func TestToBuffer(t *testing.T) {
	input := "line1\nline2\nline3"
	buf := toBuffer(input)
	if len(buf) != 3 {
		t.Errorf("expected 3 lines, got %d", len(buf))
	}
	if string(buf[0]) != "line1" || string(buf[1]) != "line2" || string(buf[2]) != "line3" {
		t.Errorf("buffer content mismatch: %v", buf)
	}
}

func TestInsertRune(t *testing.T) {
	buf := [][]rune{[]rune("abc")}
	cur := Cursor{0, 1}
	insertRune(&buf, &cur, 'X')
	if string(buf[0]) != "aXbc" {
		t.Errorf("expected 'aXbc', got '%s'", string(buf[0]))
	}
	if cur.Col != 2 {
		t.Errorf("expected cursor col 2, got %d", cur.Col)
	}
}

func TestBackspace(t *testing.T) {
	buf := [][]rune{[]rune("abc")}
	cur := Cursor{0, 2}
	backspace(&buf, &cur)
	if string(buf[0]) != "ac" {
		t.Errorf("expected 'ac', got '%s'", string(buf[0]))
	}
	if cur.Col != 1 {
		t.Errorf("expected cursor col 1, got %d", cur.Col)
	}
}

// ── deleteSelection ───────────────────────────────────────────────────────────

func TestDeleteSelectionSameLine(t *testing.T) {
	buf := [][]rune{[]rune("hello world")}
	cur := Cursor{0, 0}
	sel := Selection{Active: true, StartRow: 0, StartCol: 2, EndRow: 0, EndCol: 7}
	deleteSelection(&buf, &cur, &sel)
	if string(buf[0]) != "heorld" {
		t.Errorf("expected 'heorld', got '%s'", string(buf[0]))
	}
	if cur.Row != 0 || cur.Col != 2 {
		t.Errorf("expected cursor at 0,2, got %d,%d", cur.Row, cur.Col)
	}
	if sel.Active {
		t.Error("selection should be cleared after delete")
	}
}

func TestDeleteSelectionMultiLine(t *testing.T) {
	buf := [][]rune{[]rune("hello"), []rune("world"), []rune("foo")}
	cur := Cursor{0, 0}
	sel := Selection{Active: true, StartRow: 0, StartCol: 3, EndRow: 1, EndCol: 3}
	deleteSelection(&buf, &cur, &sel)
	if len(buf) != 2 {
		t.Fatalf("expected 2 lines after multi-line delete, got %d", len(buf))
	}
	if string(buf[0]) != "helld" {
		t.Errorf("expected 'helld', got '%s'", string(buf[0]))
	}
	if string(buf[1]) != "foo" {
		t.Errorf("expected 'foo', got '%s'", string(buf[1]))
	}
	if cur.Row != 0 || cur.Col != 3 {
		t.Errorf("expected cursor at 0,3, got %d,%d", cur.Row, cur.Col)
	}
}

func TestDeleteSelectionInactive(t *testing.T) {
	buf := [][]rune{[]rune("hello")}
	cur := Cursor{0, 2}
	sel := Selection{Active: false}
	deleteSelection(&buf, &cur, &sel)
	if string(buf[0]) != "hello" {
		t.Error("buffer should not change when selection is inactive")
	}
}

func TestDeleteSelectionInvertedBounds(t *testing.T) {
	// Selection drawn right-to-left: EndRow/Col is before StartRow/Col
	buf := [][]rune{[]rune("abcdef")}
	cur := Cursor{0, 0}
	sel := Selection{Active: true, StartRow: 0, StartCol: 5, EndRow: 0, EndCol: 2}
	deleteSelection(&buf, &cur, &sel)
	if string(buf[0]) != "abf" {
		t.Errorf("expected 'abf', got '%s'", string(buf[0]))
	}
}

// ── normalizeSelection ────────────────────────────────────────────────────────

func TestNormalizeSelectionForward(t *testing.T) {
	sel := Selection{Active: true, StartRow: 0, StartCol: 2, EndRow: 1, EndCol: 5}
	sr, sc, er, ec := normalizeSelection(&sel)
	if sr != 0 || sc != 2 || er != 1 || ec != 5 {
		t.Errorf("forward selection should be unchanged, got %d,%d → %d,%d", sr, sc, er, ec)
	}
}

func TestNormalizeSelectionInverted(t *testing.T) {
	sel := Selection{Active: true, StartRow: 2, StartCol: 5, EndRow: 0, EndCol: 1}
	sr, sc, er, ec := normalizeSelection(&sel)
	if sr != 0 || sc != 1 || er != 2 || ec != 5 {
		t.Errorf("inverted selection should be flipped, got %d,%d → %d,%d", sr, sc, er, ec)
	}
}

func TestNormalizeSelectionSameRowInverted(t *testing.T) {
	sel := Selection{Active: true, StartRow: 1, StartCol: 8, EndRow: 1, EndCol: 3}
	sr, sc, er, ec := normalizeSelection(&sel)
	if sr != 1 || sc != 3 || er != 1 || ec != 8 {
		t.Errorf("same-row inverted should swap cols, got %d,%d → %d,%d", sr, sc, er, ec)
	}
}

// ── insertNewline ─────────────────────────────────────────────────────────────

func TestInsertNewlineMiddle(t *testing.T) {
	buf := [][]rune{[]rune("hello world")}
	cur := Cursor{0, 5}
	insertNewline(&buf, &cur)
	if len(buf) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(buf))
	}
	if string(buf[0]) != "hello" || string(buf[1]) != " world" {
		t.Errorf("expected 'hello' / ' world', got '%s' / '%s'", string(buf[0]), string(buf[1]))
	}
	if cur.Row != 1 || cur.Col != 0 {
		t.Errorf("expected cursor at 1,0, got %d,%d", cur.Row, cur.Col)
	}
}

func TestInsertNewlineAtStart(t *testing.T) {
	buf := [][]rune{[]rune("hello")}
	cur := Cursor{0, 0}
	insertNewline(&buf, &cur)
	if len(buf) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(buf))
	}
	if string(buf[0]) != "" || string(buf[1]) != "hello" {
		t.Errorf("expected '' / 'hello', got '%s' / '%s'", string(buf[0]), string(buf[1]))
	}
}

func TestInsertNewlineAtEnd(t *testing.T) {
	buf := [][]rune{[]rune("hello")}
	cur := Cursor{0, 5}
	insertNewline(&buf, &cur)
	if len(buf) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(buf))
	}
	if string(buf[0]) != "hello" || string(buf[1]) != "" {
		t.Errorf("expected 'hello' / '', got '%s' / '%s'", string(buf[0]), string(buf[1]))
	}
}

// ── addLineTab / removeLineTab ────────────────────────────────────────────────

func TestAddLineTabNoSelection(t *testing.T) {
	buf := [][]rune{[]rune("hello")}
	cur := Cursor{0, 0}
	sel := Selection{Active: false}
	addLineTab(&buf, &cur, &sel)
	if string(buf[0]) != "    hello" {
		t.Errorf("expected '    hello', got '%s'", string(buf[0]))
	}
	if cur.Col != 4 {
		t.Errorf("expected cursor col 4, got %d", cur.Col)
	}
}

func TestAddLineTabWithSelection(t *testing.T) {
	buf := [][]rune{[]rune("line1"), []rune("line2"), []rune("line3")}
	cur := Cursor{0, 0}
	sel := Selection{Active: true, StartRow: 0, StartCol: 0, EndRow: 1, EndCol: 5}
	addLineTab(&buf, &cur, &sel)
	if string(buf[0]) != "    line1" {
		t.Errorf("expected '    line1', got '%s'", string(buf[0]))
	}
	if string(buf[1]) != "    line2" {
		t.Errorf("expected '    line2', got '%s'", string(buf[1]))
	}
	if string(buf[2]) != "line3" {
		t.Errorf("line3 should be unchanged, got '%s'", string(buf[2]))
	}
}

func TestRemoveLineTabFullIndent(t *testing.T) {
	buf := [][]rune{[]rune("    hello")}
	cur := Cursor{0, 4}
	sel := Selection{Active: false}
	removeLineTab(&buf, &cur, &sel)
	if string(buf[0]) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(buf[0]))
	}
	if cur.Col != 0 {
		t.Errorf("expected cursor col 0, got %d", cur.Col)
	}
}

func TestRemoveLineTabPartialIndent(t *testing.T) {
	buf := [][]rune{[]rune("  hello")}
	cur := Cursor{0, 2}
	sel := Selection{Active: false}
	removeLineTab(&buf, &cur, &sel)
	if string(buf[0]) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(buf[0]))
	}
	if cur.Col != 0 {
		t.Errorf("expected cursor col 0, got %d", cur.Col)
	}
}

func TestRemoveLineTabNoIndent(t *testing.T) {
	buf := [][]rune{[]rune("hello")}
	cur := Cursor{0, 0}
	sel := Selection{Active: false}
	removeLineTab(&buf, &cur, &sel)
	if string(buf[0]) != "hello" {
		t.Error("line with no indent should be unchanged")
	}
}

// ── replaceAll ────────────────────────────────────────────────────────────────

func TestReplaceAllNoMatch(t *testing.T) {
	if got := replaceAll("hello world", "xyz", "abc"); got != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", got)
	}
}

func TestReplaceAllSingleMatch(t *testing.T) {
	if got := replaceAll("hello world", "world", "there"); got != "hello there" {
		t.Errorf("expected 'hello there', got '%s'", got)
	}
}

func TestReplaceAllMultipleMatches(t *testing.T) {
	if got := replaceAll("aaa", "a", "b"); got != "bbb" {
		t.Errorf("expected 'bbb', got '%s'", got)
	}
}

func TestReplaceAllLongerReplacement(t *testing.T) {
	if got := replaceAll("foo bar foo", "foo", "foobar"); got != "foobar bar foobar" {
		t.Errorf("expected 'foobar bar foobar', got '%s'", got)
	}
}

func TestReplaceAllShorterReplacement(t *testing.T) {
	if got := replaceAll("hello world", "world", "x"); got != "hello x" {
		t.Errorf("expected 'hello x', got '%s'", got)
	}
}

// ── moveCursor ────────────────────────────────────────────────────────────────

func TestMoveCursorBasic(t *testing.T) {
	buf := [][]rune{[]rune("hello"), []rune("world")}
	cur := Cursor{0, 2}
	rowOffset := 0

	moveCursor(&cur, "down", buf, &rowOffset, 10)
	if cur.Row != 1 || cur.Col != 2 {
		t.Errorf("expected 1,2 after down, got %d,%d", cur.Row, cur.Col)
	}
	moveCursor(&cur, "up", buf, &rowOffset, 10)
	if cur.Row != 0 || cur.Col != 2 {
		t.Errorf("expected 0,2 after up, got %d,%d", cur.Row, cur.Col)
	}
	moveCursor(&cur, "right", buf, &rowOffset, 10)
	if cur.Col != 3 {
		t.Errorf("expected col 3 after right, got %d", cur.Col)
	}
	moveCursor(&cur, "left", buf, &rowOffset, 10)
	if cur.Col != 2 {
		t.Errorf("expected col 2 after left, got %d", cur.Col)
	}
}

func TestMoveCursorBoundaries(t *testing.T) {
	buf := [][]rune{[]rune("hi"), []rune("world")}
	rowOffset := 0

	// Up at row 0 should stay.
	cur := Cursor{0, 1}
	moveCursor(&cur, "up", buf, &rowOffset, 10)
	if cur.Row != 0 {
		t.Errorf("expected to stay at row 0, got %d", cur.Row)
	}

	// Left at 0,0 should stay.
	cur = Cursor{0, 0}
	moveCursor(&cur, "left", buf, &rowOffset, 10)
	if cur.Row != 0 || cur.Col != 0 {
		t.Errorf("expected to stay at 0,0, got %d,%d", cur.Row, cur.Col)
	}

	// Down at last row should stay.
	cur = Cursor{1, 2}
	moveCursor(&cur, "down", buf, &rowOffset, 10)
	if cur.Row != 1 {
		t.Errorf("expected to stay at row 1, got %d", cur.Row)
	}

	// Right at end of last line should stay.
	cur = Cursor{1, 5}
	moveCursor(&cur, "right", buf, &rowOffset, 10)
	if cur.Row != 1 || cur.Col != 5 {
		t.Errorf("expected to stay at 1,5, got %d,%d", cur.Row, cur.Col)
	}
}

func TestMoveCursorLineWrap(t *testing.T) {
	buf := [][]rune{[]rune("hello"), []rune("world")}
	rowOffset := 0

	// Right at end of line wraps to start of next.
	cur := Cursor{0, 5}
	moveCursor(&cur, "right", buf, &rowOffset, 10)
	if cur.Row != 1 || cur.Col != 0 {
		t.Errorf("expected wrap to 1,0, got %d,%d", cur.Row, cur.Col)
	}

	// Left at col 0 wraps to end of previous line.
	cur = Cursor{1, 0}
	moveCursor(&cur, "left", buf, &rowOffset, 10)
	if cur.Row != 0 || cur.Col != 5 {
		t.Errorf("expected wrap to 0,5, got %d,%d", cur.Row, cur.Col)
	}
}

func TestMoveCursorColClampOnRowChange(t *testing.T) {
	// Moving to a shorter line should clamp col.
	buf := [][]rune{[]rune("hello world"), []rune("hi")}
	cur := Cursor{0, 8}
	rowOffset := 0
	moveCursor(&cur, "down", buf, &rowOffset, 10)
	if cur.Col != 2 {
		t.Errorf("expected col clamped to 2, got %d", cur.Col)
	}
}

func TestMoveCursorScrollDown(t *testing.T) {
	buf := [][]rune{
		[]rune("line0"), []rune("line1"), []rune("line2"),
		[]rune("line3"), []rune("line4"),
	}
	cur := Cursor{2, 0}
	rowOffset := 0
	screenRows := 3

	moveCursor(&cur, "down", buf, &rowOffset, screenRows) // row 3 → triggers scroll
	moveCursor(&cur, "down", buf, &rowOffset, screenRows) // row 4 → triggers scroll
	if rowOffset != 2 {
		t.Errorf("expected rowOffset 2 after scrolling down, got %d", rowOffset)
	}
}

func TestMoveCursorScrollUp(t *testing.T) {
	buf := [][]rune{
		[]rune("line0"), []rune("line1"), []rune("line2"), []rune("line3"),
	}
	cur := Cursor{3, 0}
	rowOffset := 2
	screenRows := 3

	moveCursor(&cur, "up", buf, &rowOffset, screenRows) // row 2
	moveCursor(&cur, "up", buf, &rowOffset, screenRows) // row 1 → scroll
	moveCursor(&cur, "up", buf, &rowOffset, screenRows) // row 0 → scroll
	if rowOffset != 0 {
		t.Errorf("expected rowOffset 0 after scrolling up to top, got %d", rowOffset)
	}
}

// ── word movement ─────────────────────────────────────────────────────────────

func TestMoveWordRight(t *testing.T) {
	buf := [][]rune{[]rune("hello world foo")}
	cur := Cursor{0, 0}

	moveWordRight(&cur, buf) // skip "hello", skip " " → land on "world"
	if cur.Col != 6 {
		t.Errorf("expected col 6, got %d", cur.Col)
	}
	moveWordRight(&cur, buf) // skip "world", skip " " → land on "foo"
	if cur.Col != 12 {
		t.Errorf("expected col 12, got %d", cur.Col)
	}
	moveWordRight(&cur, buf) // skip "foo" → end of line
	if cur.Col != 15 {
		t.Errorf("expected col 15 (end of line), got %d", cur.Col)
	}
}

func TestMoveWordRightWrapsLine(t *testing.T) {
	buf := [][]rune{[]rune("hello"), []rune("world")}
	cur := Cursor{0, 5} // already at end of line 0

	moveWordRight(&cur, buf)
	if cur.Row != 1 || cur.Col != 0 {
		t.Errorf("expected wrap to 1,0, got %d,%d", cur.Row, cur.Col)
	}
}

func TestMoveWordLeft(t *testing.T) {
	buf := [][]rune{[]rune("hello world foo")}
	cur := Cursor{0, 15} // past end

	moveWordLeft(&cur, buf) // skip "foo" backwards → land at start of "foo"
	if cur.Col != 12 {
		t.Errorf("expected col 12, got %d", cur.Col)
	}
	moveWordLeft(&cur, buf) // skip " ", skip "world" backwards → start of "world"
	if cur.Col != 6 {
		t.Errorf("expected col 6, got %d", cur.Col)
	}
	moveWordLeft(&cur, buf) // skip " ", skip "hello" backwards → start of "hello"
	if cur.Col != 0 {
		t.Errorf("expected col 0, got %d", cur.Col)
	}
}

func TestMoveWordLeftWrapsLine(t *testing.T) {
	buf := [][]rune{[]rune("hello"), []rune("world")}
	cur := Cursor{1, 0} // at start of line 1

	moveWordLeft(&cur, buf)
	if cur.Row != 0 || cur.Col != 5 {
		t.Errorf("expected wrap to 0,5, got %d,%d", cur.Row, cur.Col)
	}
}

// ── getSelectedText ───────────────────────────────────────────────────────────

func TestGetSelectedTextSingleLine(t *testing.T) {
	buf := [][]rune{[]rune("hello world")}
	sel := Selection{Active: true, StartRow: 0, StartCol: 6, EndRow: 0, EndCol: 11}
	got := getSelectedText(buf, &sel)
	if got != "world" {
		t.Errorf("expected 'world', got '%s'", got)
	}
}

func TestGetSelectedTextMultiLine(t *testing.T) {
	buf := [][]rune{[]rune("hello"), []rune("world"), []rune("foo")}
	sel := Selection{Active: true, StartRow: 0, StartCol: 3, EndRow: 2, EndCol: 3}
	got := getSelectedText(buf, &sel)
	if got != "lo\nworld\nfoo" {
		t.Errorf("expected 'lo\\nworld\\nfoo', got '%s'", got)
	}
}
