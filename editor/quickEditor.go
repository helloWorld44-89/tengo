package editor

import (
	"fmt"
	"os"
	"github.com/helloWorld44-89/tengo/file"
)

// indexOf searches for a substring in a line of runes
func indexOf(line []rune, term string) int {
	if len(term) == 0 {
		return -1
	}
	return findIndex(string(line), term)
}

// findIndex returns the index of substr in s, or -1 if not found
func findIndex(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// replaceAll replaces all occurrences of old with new in str
func replaceAll(str, old, newStr string) string {
	result := ""
	for len(str) > 0 {
		idx := findIndex(str, old)
		if idx == -1 {
			result += str
			break
		}
		result += str[:idx] + newStr
		str = str[idx+len(old):]
	}
	return result
}

// tabState holds the complete editing state for a single open file.
type tabState struct {
	filePath  string
	buf       [][]rune
	cursor    Cursor
	modified  bool
	errorLine int
	rowOffset int
	sel       Selection
	status    string
	undoStack [][][]rune
	redoStack [][][]rune
}

func newTab(filePath string) tabState {
	content, err := file.OpenFile(filePath)
	if err != nil {
		// New or unreadable file — open with empty buffer.
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: %v\r\n", err)
		}
		content = ""
	}
	return tabState{
		filePath: filePath,
		buf:      toBuffer(content),
		status:   "Editing",
	}
}

// RunQuickEditor opens one or more files for quick editing.
// updateCh, if non-nil, receives a single status message when a newer version
// is available; it is displayed in the status bar on the next redraw.
func RunQuickEditor(filePaths []string, updateCh <-chan string) {
	if len(filePaths) == 0 {
		return
	}

	tabs := make([]tabState, len(filePaths))
	for i, fp := range filePaths {
		tabs[i] = newTab(fp)
	}

	activeTab := 0
	showLineNumbers := false
	wordWrap := false
	inBracketedPaste := false

	// buildDrawArgs assembles the slice-level args draw() needs for the tab bar.
	buildTabArgs := func() ([]string, []bool) {
		names := make([]string, len(tabs))
		mods := make([]bool, len(tabs))
		for i, t := range tabs {
			names[i] = t.filePath
			mods[i] = t.modified
		}
		return names, mods
	}

	for {
		t := &tabs[activeTab]

		// Closures that operate on the active tab's undo stacks.
		pushUndo := func(state [][]rune) {
			t.modified = true
			t.errorLine = 0
			t.redoStack = nil
			cp := make([][]rune, len(state))
			for i := range state {
				cp[i] = make([]rune, len(state[i]))
				copy(cp[i], state[i])
			}
			t.undoStack = append(t.undoStack, cp)
		}
		popUndo := func() ([][]rune, bool) {
			if len(t.undoStack) == 0 {
				return nil, false
			}
			prev := t.undoStack[len(t.undoStack)-1]
			t.undoStack = t.undoStack[:len(t.undoStack)-1]
			t.redoStack = append(t.redoStack, prev)
			return prev, true
		}
		popRedo := func() ([][]rune, bool) {
			if len(t.redoStack) == 0 {
				return nil, false
			}
			next := t.redoStack[len(t.redoStack)-1]
			t.redoStack = t.redoStack[:len(t.redoStack)-1]
			t.undoStack = append(t.undoStack, next)
			return next, true
		}

		width, height := getTerminalSize()
		usableRows := height - 3
		if len(tabs) > 1 {
			usableRows--
		}

		// Ensure cursor is visible in word-wrap mode before drawing.
		if wordWrap {
			gW := gutterWidth(t.buf, showLineNumbers, t.errorLine)
			cWidth := width - gW
			if cWidth < 1 {
				cWidth = 1
			}
			adjustScrollWordWrap(&t.rowOffset, t.cursor, t.buf, usableRows, cWidth)
		}

		// Show update notification in the status bar as soon as it arrives.
		if updateCh != nil {
			select {
			case msg := <-updateCh:
				for i := range tabs {
					if tabs[i].status == "Editing" {
						tabs[i].status = msg
					}
				}
			default:
			}
		}

		tabNames, tabMods := buildTabArgs()
		draw(t.buf, t.cursor, t.filePath, t.status, t.modified, t.rowOffset, &t.sel, showLineNumbers, t.errorLine, tabNames, tabMods, activeTab, wordWrap)

		key := readKey()

		switch key {

		// ── Tab switching ──────────────────────────────────────────────────
		case "ctrl-w":
			if len(tabs) > 1 {
				activeTab = (activeTab + 1) % len(tabs)
			}

		case "alt-1", "alt-2", "alt-3", "alt-4", "alt-5",
			"alt-6", "alt-7", "alt-8", "alt-9":
			n := int(key[4]-'0') - 1
			if n >= 0 && n < len(tabs) {
				activeTab = n
			}

		// ── Quit ───────────────────────────────────────────────────────────
		case "ctrl-q", "esc":
			anyModified := false
			for _, tab := range tabs {
				if tab.modified {
					anyModified = true
					break
				}
			}
			if anyModified {
				prompt := "Unsaved changes — quit? (y/n): "
				if len(tabs) > 1 {
					prompt = "Unsaved changes in open tabs — quit? (y/n): "
				}
				answer, ok := readPrompt(prompt)
				if !ok || (answer != "y" && answer != "Y") {
					break
				}
			}
			return

		// ── File operations ────────────────────────────────────────────────
		case "ctrl-s":
			if err := file.SaveFile(t.filePath, t.buf); err != nil {
				t.status = fmt.Sprintf("Save failed: %v", err)
			} else {
				t.modified = false
				result := ValidateBuffer(t.filePath, t.buf)
				if !result.Valid {
					t.status = "Saved — " + result.Message
					t.errorLine = result.Line
				} else {
					t.status = "Saved"
					t.errorLine = 0
				}
			}

		case "ctrl-e":
			result := ValidateBuffer(t.filePath, t.buf)
			if result.Valid {
				t.status = "✓ " + result.Message
				t.errorLine = 0
			} else {
				t.status = result.Message
				t.errorLine = result.Line
				if result.Line > 0 && result.Line <= len(t.buf) {
					t.cursor.Row = result.Line - 1
					t.cursor.Col = 0
				}
			}

		case "ctrl-t":
			newBuf, err := FormatBuffer(t.filePath, t.buf)
			if err != nil {
				t.status = "Format failed: " + err.Error()
			} else {
				pushUndo(t.buf)
				t.buf = newBuf
				t.status = "Formatted"
				t.errorLine = 0
				if t.cursor.Row >= len(t.buf) {
					t.cursor.Row = len(t.buf) - 1
				}
				if t.cursor.Col > len(t.buf[t.cursor.Row]) {
					t.cursor.Col = len(t.buf[t.cursor.Row])
				}
			}

		case "ctrl-g":
			if input, ok := readPrompt("Go to line: "); ok && input != "" {
				n := 0
				fmt.Sscanf(input, "%d", &n)
				if n >= 1 && n <= len(t.buf) {
					t.cursor.Row = n - 1
					t.cursor.Col = 0
				}
			}

		// ── Undo / Redo ────────────────────────────────────────────────────
		case "ctrl-z":
			if prev, ok := popUndo(); ok {
				t.buf = prev
				clampCursor(&t.cursor, t.buf)
			}

		case "ctrl-y":
			if next, ok := popRedo(); ok {
				t.buf = next
				clampCursor(&t.cursor, t.buf)
			}

		// ── Editing ────────────────────────────────────────────────────────
		case "ctrl-shift-enter":
			pushUndo(t.buf)
			if t.cursor.Row >= 0 && t.cursor.Row < len(t.buf) {
				t.buf = append(t.buf[:t.cursor.Row], append([][]rune{{}}, t.buf[t.cursor.Row:]...)...)
			}

		case "shift-enter":
			pushUndo(t.buf)
			if t.cursor.Row < len(t.buf) {
				t.buf = append(t.buf[:t.cursor.Row+1], append([][]rune{{}}, t.buf[t.cursor.Row+1:]...)...)
				t.cursor.Row++
			}

		case "ctrl-/":
			pushUndo(t.buf)
			if t.cursor.Row >= 0 && t.cursor.Row < len(t.buf) {
				line := t.buf[t.cursor.Row]
				if len(line) >= 2 && line[0] == '/' && line[1] == '/' {
					t.buf[t.cursor.Row] = line[2:]
				} else {
					t.buf[t.cursor.Row] = append([]rune{'/', '/'}, line...)
				}
			}

		case "ctrl-d":
			pushUndo(t.buf)
			if t.cursor.Row >= 0 && t.cursor.Row < len(t.buf) {
				line := make([]rune, len(t.buf[t.cursor.Row]))
				copy(line, t.buf[t.cursor.Row])
				t.buf = append(t.buf[:t.cursor.Row+1], append([][]rune{line}, t.buf[t.cursor.Row+1:]...)...)
				t.cursor.Row++
			}

		case "ctrl-[":
			removeLineTab(&t.buf, &t.cursor, &t.sel)

		case "ctrl-]":
			addLineTab(&t.buf, &t.cursor, &t.sel)

		case "tab":
			deleteSelection(&t.buf, &t.cursor, &t.sel)
			pushUndo(t.buf)
			for i := 0; i < 4; i++ {
				insertRune(&t.buf, &t.cursor, ' ')
			}

		case "enter":
			pushUndo(t.buf)
			deleteSelection(&t.buf, &t.cursor, &t.sel)
			insertNewline(&t.buf, &t.cursor)
			moveCursor(&t.cursor, "", t.buf, &t.rowOffset, usableRows)

		case "backspace":
			pushUndo(t.buf)
			deleteSelection(&t.buf, &t.cursor, &t.sel)
			backspace(&t.buf, &t.cursor)

		case "delete":
			pushUndo(t.buf)
			deleteSelection(&t.buf, &t.cursor, &t.sel)
			deleteForward(&t.buf, &t.cursor)

		// ── Selection ──────────────────────────────────────────────────────
		case "ctrl-a":
			if len(t.buf) == 0 {
				break
			}
			t.sel.Active = true
			t.sel.StartRow, t.sel.StartCol = 0, 0
			t.sel.EndRow = len(t.buf) - 1
			t.sel.EndCol = len(t.buf[t.sel.EndRow])
			t.cursor.Row = t.sel.EndRow
			t.cursor.Col = t.sel.EndCol

		// ── Find / Replace ─────────────────────────────────────────────────
		case "ctrl-f":
			if term, ok := readPrompt("Find: "); ok && term != "" {
				for row, line := range t.buf {
					if idx := indexOf(line, term); idx != -1 {
						t.cursor.Row = row
						t.cursor.Col = idx
						break
					}
				}
			}

		case "ctrl-r":
			searchTerm, ok := readPrompt("Find: ")
			if !ok || searchTerm == "" {
				break
			}
			replaceTerm, ok := readPrompt("Replace with: ")
			if !ok {
				break
			}
			pushUndo(t.buf)
			count := 0
			for row := range t.buf {
				str := string(t.buf[row])
				if findIndex(str, searchTerm) != -1 {
					t.buf[row] = []rune(replaceAll(str, searchTerm, replaceTerm))
					count++
				}
			}
			t.status = fmt.Sprintf("Replaced %d occurrence(s)", count)
			if t.cursor.Row >= len(t.buf) {
				t.cursor.Row = len(t.buf) - 1
			}
			if t.cursor.Row >= 0 && t.cursor.Col > len(t.buf[t.cursor.Row]) {
				t.cursor.Col = len(t.buf[t.cursor.Row])
			}

		// ── Navigation ─────────────────────────────────────────────────────
		case "home":
			t.cursor.Col = 0

		case "end":
			t.cursor.Col = len(t.buf[t.cursor.Row])

		case "up", "down", "left", "right":
			t.sel.Active = false
			moveCursor(&t.cursor, key, t.buf, &t.rowOffset, usableRows)

		case "alt-left", "alt-right", "alt-up", "alt-down":
			moveCursor(&t.cursor, key, t.buf, &t.rowOffset, usableRows)

		case "ctrl-left":
			startSelectionIfNeeded(&t.sel, &t.cursor)
			moveWordLeft(&t.cursor, t.buf)
			updateSelection(&t.sel, &t.cursor)
			clampSelection(&t.sel, t.buf)

		case "ctrl-right":
			startSelectionIfNeeded(&t.sel, &t.cursor)
			moveWordRight(&t.cursor, t.buf)
			updateSelection(&t.sel, &t.cursor)
			clampSelection(&t.sel, t.buf)

		case "ctrl-up":
			startSelectionIfNeeded(&t.sel, &t.cursor)
			if t.cursor.Row > 0 {
				t.cursor.Row--
				if t.cursor.Col > len(t.buf[t.cursor.Row]) {
					t.cursor.Col = len(t.buf[t.cursor.Row])
				}
			}
			updateSelection(&t.sel, &t.cursor)
			clampSelection(&t.sel, t.buf)

		case "ctrl-down":
			startSelectionIfNeeded(&t.sel, &t.cursor)
			if t.cursor.Row < len(t.buf)-1 {
				t.cursor.Row++
				if t.cursor.Col > len(t.buf[t.cursor.Row]) {
					t.cursor.Col = len(t.buf[t.cursor.Row])
				}
			}
			updateSelection(&t.sel, &t.cursor)
			clampSelection(&t.sel, t.buf)

		// ── Clipboard ──────────────────────────────────────────────────────
		case "ctrl-c":
			copySelection(t.buf, &t.sel)

		case "ctrl-x":
			pushUndo(t.buf)
			cutSelection(&t.buf, &t.cursor, &t.sel)

		case "ctrl-v":
			pushUndo(t.buf)
			pasteText(&t.buf, &t.cursor)

		// ── UI toggles ─────────────────────────────────────────────────────
		case "ctrl-h":
			ShowHelp()

		case "ctrl-n":
			showLineNumbers = !showLineNumbers

		case "alt-w":
			wordWrap = !wordWrap

		// ── Bracketed paste ────────────────────────────────────────────────
		case "paste-begin":
			inBracketedPaste = true

		case "paste-end":
			inBracketedPaste = false

		// ── Default / character input ──────────────────────────────────────
		default:
			if inBracketedPaste {
				for _, ch := range key {
					if ch == '\n' {
						insertNewline(&t.buf, &t.cursor)
					} else {
						insertRune(&t.buf, &t.cursor, ch)
					}
				}
				continue
			}
			if len(key) == 1 && key[0] >= 32 {
				deleteSelection(&t.buf, &t.cursor, &t.sel)
				t.sel.Active = false
				r := rune(key[0])
				if !handleAutoClose(&t.buf, &t.cursor, r) {
					insertRune(&t.buf, &t.cursor, r)
				}
			}
		}
	}
}
