package editor

import (
    "fmt"
    "os"
    "tengo/file"
)

// indexOf searches for a substring in a line of runes
func indexOf(line []rune, term string) int {
	if len(term) == 0 {
		return -1
	}
	str := string(line)
	return findIndex(str, term)
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
func replaceAll(str, old, new string) string {
	result := ""
	for len(str) > 0 {
		idx := findIndex(str, old)
		if idx == -1 {
			result += str
			break
		}
		result += str[:idx] + new
		str = str[idx+len(old):]
	}
	return result
}

// RunQuickEditor opens a file for quick editing with keyboard-driven navigation
// It supports undo/redo, find, selection, clipboard operations, and advanced shortcuts
func RunQuickEditor(filePath string) {
	modified := false
	errorLine := 0

	undoStack := make([][][]rune, 0)
	redoStack := make([][][]rune, 0)
	pushUndo := func(state [][]rune) {
		modified = true
		errorLine = 0
		redoStack = nil
		copyBuf := make([][]rune, len(state))
		for i := range state {
			copyBuf[i] = make([]rune, len(state[i]))
			copy(copyBuf[i], state[i])
		}
		undoStack = append(undoStack, copyBuf)
	}
	popUndo := func() ([][]rune, bool) {
		// move to redo stack
		if len(undoStack) == 0 {
			return nil, false
		}
		prev := undoStack[len(undoStack)-1]
		undoStack = undoStack[:len(undoStack)-1]
		redoStack = append(redoStack, prev)
		return prev, true
	}
	popRedo := func() ([][]rune, bool) {
		if len(redoStack) == 0 {
			return nil, false
		}
		next := redoStack[len(redoStack)-1]
		redoStack = redoStack[:len(redoStack)-1]
		undoStack = append(undoStack, next)
		return next, true
	}

	cursor := Cursor{0, 0}

	content, err := file.OpenFile(filePath)
	if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to open file: %v\n", err)
        return
	}

	buf := toBuffer(content)
	status := "Editing"

	rowOffset := 0
	sel := Selection{}
	inBracketedPaste := false
	showLineNumbers := false

	for {
		draw(buf, cursor, filePath, status, modified, rowOffset, &sel, showLineNumbers, errorLine)

		_, height := getTerminalSize()
		usableRows := height - 3

		key := readKey()

		switch key {
		case "ctrl-shift-enter":
			pushUndo(buf)
			if cursor.Row >= 0 && cursor.Row < len(buf) {
				newLine := []rune{}
				buf = append(buf[:cursor.Row], append([][]rune{newLine}, buf[cursor.Row:]...)...)
			}

		case "shift-enter":
			pushUndo(buf)
			if cursor.Row < len(buf) {
				newLine := []rune{}
				buf = append(buf[:cursor.Row+1], append([][]rune{newLine}, buf[cursor.Row+1:]...)...)
				cursor.Row++
			}

		case "ctrl-/":
			pushUndo(buf)
			if cursor.Row >= 0 && cursor.Row < len(buf) {
				line := buf[cursor.Row]
				if len(line) >= 2 && line[0] == '/' && line[1] == '/' {
					buf[cursor.Row] = line[2:]
				} else {
					buf[cursor.Row] = append([]rune{'/', '/'}, line...)
				}
			}

		case "ctrl-d":
			pushUndo(buf)
			if cursor.Row >= 0 && cursor.Row < len(buf) {
				line := make([]rune, len(buf[cursor.Row]))
				copy(line, buf[cursor.Row])
				buf = append(buf[:cursor.Row+1], append([][]rune{line}, buf[cursor.Row+1:]...)...)
				cursor.Row++
			}

		case "ctrl-a":
			if len(buf) == 0 {
				break
			}
			sel.Active = true
			sel.StartRow = 0
			sel.StartCol = 0
			sel.EndRow = len(buf) - 1
			sel.EndCol = len(buf[len(buf)-1])
			cursor.Row = len(buf) - 1
			cursor.Col = len(buf[len(buf)-1])

		case "ctrl-f":
			if term, ok := readPrompt("Find: "); ok && term != "" {
				for row, line := range buf {
					if idx := indexOf(line, term); idx != -1 {
						cursor.Row = row
						cursor.Col = idx
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

			pushUndo(buf)
			replacementCount := 0
			for row := range buf {
				str := string(buf[row])
				if findIndex(str, searchTerm) != -1 {
					buf[row] = []rune(replaceAll(str, searchTerm, replaceTerm))
					replacementCount++
				}
			}
			status = fmt.Sprintf("Replaced %d occurrences", replacementCount)

			// Clamp cursor — replacements may have shortened the current line.
			if cursor.Row >= len(buf) {
				cursor.Row = len(buf) - 1
			}
			if cursor.Row >= 0 && cursor.Col > len(buf[cursor.Row]) {
				cursor.Col = len(buf[cursor.Row])
			}

		case "ctrl-y":
			if next, ok := popRedo(); ok {
				buf = next
			}

		case "ctrl-z":
			if prev, ok := popUndo(); ok {
				buf = prev
			}


		case "ctrl-q", "esc":
			if modified {
				answer, ok := readPrompt("Unsaved changes — quit? (y/n): ")
				if !ok || (answer != "y" && answer != "Y") {
					break
				}
			}
			return

		case "home":
			cursor.Col = 0

		case "end":
			cursor.Col = len(buf[cursor.Row])

		case "up", "down", "left", "right":
			sel.Active = false
			moveCursor(&cursor, key, buf, &rowOffset, usableRows)

		case "tab":
			deleteSelection(&buf, &cursor, &sel)
			pushUndo(buf)
			for i := 0; i < 4; i++ {
				insertRune(&buf, &cursor, ' ')
			}

		case "enter":
			pushUndo(buf)
			deleteSelection(&buf, &cursor, &sel)
			insertNewline(&buf, &cursor)
			moveCursor(&cursor, "", buf, &rowOffset, usableRows)

		case "backspace":
			pushUndo(buf)
			deleteSelection(&buf, &cursor, &sel)
			backspace(&buf, &cursor)

		case "ctrl-s":
			if err := file.SaveFile(filePath, buf); err != nil {
				status = fmt.Sprintf("Save failed: %v", err)
			} else {
				modified = false
				result := ValidateBuffer(filePath, buf)
				if !result.Valid {
					status = "Saved — " + result.Message
					errorLine = result.Line
				} else {
					status = "Saved"
					errorLine = 0
				}
			}

		case "ctrl-e":
			result := ValidateBuffer(filePath, buf)
			if result.Valid {
				status = "✓ " + result.Message
				errorLine = 0
			} else {
				status = result.Message
				errorLine = result.Line
				if result.Line > 0 && result.Line <= len(buf) {
					cursor.Row = result.Line - 1
					cursor.Col = 0
				}
			}

		case "ctrl-t":
			newBuf, err := FormatBuffer(filePath, buf)
			if err != nil {
				status = "Format failed: " + err.Error()
			} else {
				pushUndo(buf)
				buf = newBuf
				status = "Formatted"
				errorLine = 0
				if cursor.Row >= len(buf) {
					cursor.Row = len(buf) - 1
				}
				if cursor.Col > len(buf[cursor.Row]) {
					cursor.Col = len(buf[cursor.Row])
				}
			}

		case "ctrl-g":
			if input, ok := readPrompt("Go to line: "); ok && input != "" {
				n := 0
				fmt.Sscanf(input, "%d", &n)
				if n >= 1 && n <= len(buf) {
					cursor.Row = n - 1
					cursor.Col = 0
				}
			}

		case "ctrl-[":
			removeLineTab(&buf, &cursor, &sel)

		case "ctrl-]":
			addLineTab(&buf, &cursor, &sel)
		
		
		case "ctrl-h":
			ShowHelp()

		case "ctrl-n":
			showLineNumbers = !showLineNumbers


		// -------- CTRL + ARROWS (word-level movement + selection) --------
		case "ctrl-left":
			startSelectionIfNeeded(&sel, &cursor)
			moveWordLeft(&cursor, buf)
			updateSelection(&sel, &cursor)
			clampSelection(&sel, buf)

		case "ctrl-right":
			startSelectionIfNeeded(&sel, &cursor)
			moveWordRight(&cursor, buf)
			updateSelection(&sel, &cursor)
			clampSelection(&sel, buf)

		case "ctrl-up":
			startSelectionIfNeeded(&sel, &cursor)
			if cursor.Row > 0 {
				cursor.Row--
				if cursor.Col > len(buf[cursor.Row]) {
					cursor.Col = len(buf[cursor.Row])
				}
			}
			updateSelection(&sel, &cursor)
			clampSelection(&sel, buf)

		case "ctrl-down":
			startSelectionIfNeeded(&sel, &cursor)
			if cursor.Row < len(buf)-1 {
				cursor.Row++
				if cursor.Col > len(buf[cursor.Row]) {
					cursor.Col = len(buf[cursor.Row])
				}
			}
			updateSelection(&sel, &cursor)
			clampSelection(&sel, buf)

		//=======Alt + Arrows for fast Navigation=======
		case "alt-left":
			moveCursor(&cursor, key, buf, &rowOffset, usableRows)
		case "alt-right":
			moveCursor(&cursor, key, buf, &rowOffset, usableRows)
		case "alt-up":
			moveCursor(&cursor, key, buf, &rowOffset, usableRows)
		case "alt-down":
			moveCursor(&cursor, key, buf, &rowOffset, usableRows)

		// -------- CLIPBOARD --------

		case "ctrl-c":
			copySelection(buf, &sel)

		case "ctrl-x":
			pushUndo(buf)
			cutSelection(&buf, &cursor, &sel)

		case "paste-begin":
			inBracketedPaste = true

		case "paste-end":
			inBracketedPaste = false

		case "ctrl-v":
			pushUndo(buf)
			pasteText(&buf, &cursor)

		// -------- SINGLE DEFAULT --------
		default:

			// If inside bracketed paste, treat bytes as literal input
			if inBracketedPaste {
				for _, b := range key {
					if b == '\n' {
						insertNewline(&buf, &cursor)
					} else {
						insertRune(&buf, &cursor, rune(b))
					}
				}
				continue
			}

			// Normal character input
			if len(key) == 1 && key[0] >= 32 {
				deleteSelection(&buf, &cursor, &sel)
				sel.Active = false
				insertRune(&buf, &cursor, rune(key[0]))
			}
		}
	}
}