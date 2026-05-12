package editor

import "strings"

// Cursor represents the current cursor position in the editor (row and column)
type Cursor struct {
	Row int
	Col int
}

// bufToString converts a rune buffer back to a newline-delimited string.
func bufToString(buf [][]rune) string {
	var b strings.Builder
	for i, line := range buf {
		b.WriteString(string(line))
		if i < len(buf)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// backspace deletes the character before the cursor, or merges lines if at column 0
func backspace(buf *[][]rune, cur *Cursor) {
    // Case 1: at the start of a line → merge upward
    if cur.Col == 0 {
        if cur.Row == 0 {
            return // cannot backspace at very top
        }

        prev := (*buf)[cur.Row-1]
        current := (*buf)[cur.Row]

        newLine := append(prev, current...)
        (*buf)[cur.Row-1] = newLine

        // remove current line
        *buf = append((*buf)[:cur.Row], (*buf)[cur.Row+1:]...)

        cur.Row--
        cur.Col = len(prev)
        return
    }

    // Case 2: normal delete-left
    line := (*buf)[cur.Row]
    (*buf)[cur.Row] = append(line[:cur.Col-1], line[cur.Col:]...)
    cur.Col--
}

// toBuffer converts a string into a 2D slice of runes (one line per slice)
func toBuffer(content string) [][]rune {
	lines :=strings.Split(content, "\n")
	buf := make([][]rune,len(lines))
	for i, line := range lines {
		buf[i]=[]rune(line)
	}
	return buf
}


// removeIndentFromLine removes leading spaces (up to tabSize) from a single line
func removeIndentFromLine(buf *[][]rune, row, tabSize int) int {
    line := (*buf)[row]
    if len(line) == 0 {
        return 0
    }

    removed := 0

    // Remove up to tabSize spaces
    for removed < tabSize && removed < len(line) && line[removed] == ' ' {
        removed++
    }

    // Apply removal
    if removed > 0 {
        (*buf)[row] = line[removed:]
    }

    return removed
}


// removeLineTab outdents the current line or all selected lines by tabSize spaces
func removeLineTab(buf *[][]rune, cur *Cursor, sel *Selection) {
    tabSize := 4

    // ---- NO SELECTION: Outdent only the current line ----
    if !sel.Active {
        removeIndentFromLine(buf, cur.Row, tabSize)
        if cur.Col >= tabSize {
            cur.Col -= tabSize
        } else {
            cur.Col = 0
        }
        return
    }

    // ---- MULTI-LINE SELECTION OUTDENT ----
    sr, _, er, _ := normalizeSelection(sel)

    for row := sr; row <= er; row++ {
        removed := removeIndentFromLine(buf, row, tabSize)

        // Shrink selection columns by however much was removed
        if sel.StartRow == row {
            sel.StartCol -= removed
            if sel.StartCol < 0 { sel.StartCol = 0 }
        }
        if sel.EndRow == row {
            sel.EndCol -= removed
            if sel.EndCol < 0 { sel.EndCol = 0 }
        }
    }

    // Fix cursor position
    if cur.Col >= tabSize {
        cur.Col -= tabSize
    } else {
        cur.Col = 0
    }
}


// addLineTab indents the current line or all selected lines by tabSize spaces
func addLineTab(buf *[][]rune, cur *Cursor, sel *Selection) {
    tabSize:=4

    // No selection: indent only the current line
    if !sel.Active {
        line := (*buf)[cur.Row]
        indent := []rune{' ', ' ', ' ', ' '}
        (*buf)[cur.Row] = append(indent, line...)
        cur.Col += tabSize
        return
    }
    sr, _, er, _ := normalizeSelection(sel)

    for row := sr; row <= er; row++ {
        line := (*buf)[row]

        indent := make([]rune, tabSize)
        for i := range indent {
            indent[i] = ' '
        }

        (*buf)[row] = append(indent, line...)
    }
    cur.Col += tabSize
    sel.StartCol += tabSize
    sel.EndCol += tabSize
}




// insertNewline splits the current line at the cursor position and inserts a new line
func insertNewline(buf *[][]rune, cur *Cursor) {
    row := cur.Row
    col := cur.Col

    // Split the line into two new lines
    left := append([]rune{}, (*buf)[row][:col]...)
    right := append([]rune{}, (*buf)[row][col:]...)

    // Replace current row
    (*buf)[row] = left

    // Insert new line
    newBuf := make([][]rune, 0, len(*buf)+1)
    newBuf = append(newBuf, (*buf)[:row+1]...)
    newBuf = append(newBuf, right)
    newBuf = append(newBuf, (*buf)[row+1:]...)
    *buf = newBuf

    // Move cursor to the start of the new line
    cur.Row++
    cur.Col = 0
}


// insertRune inserts a single rune at the cursor position and advances the cursor
func insertRune(buf *[][]rune, cur *Cursor, r rune) {
    row := cur.Row
    col := cur.Col

    line := (*buf)[row]

    // allocate new slice for safety
    newLine := make([]rune, 0, len(line)+1)

    newLine = append(newLine, line[:col]...)
    newLine = append(newLine, r)
    newLine = append(newLine, line[col:]...)

    (*buf)[row] = newLine
    cur.Col++
}


// startSelectionIfNeeded initializes a selection at the current cursor position
func startSelectionIfNeeded(sel *Selection, cur *Cursor) {
    if !sel.Active {
        sel.Active = true
        sel.StartRow = cur.Row
        sel.StartCol = cur.Col
    }
}


// updateSelection extends the selection to include the current cursor position
func updateSelection(sel *Selection, cur *Cursor) {
    sel.EndRow = cur.Row
    sel.EndCol = cur.Col
}

// clearSelection deactivates the current selection
func clearSelection(sel *Selection) {
    sel.Active = false
}


// normalizeSelection returns the selection bounds in a consistent order (top-left to bottom-right)
func normalizeSelection(sel *Selection) (sr, sc, er, ec int) {
    sr, sc = sel.StartRow, sel.StartCol
    er, ec = sel.EndRow, sel.EndCol

    if sr > er || (sr == er && sc > ec) {
        sr, er = er, sr
        sc, ec = ec, sc
    }
    return
}

// clampSelection ensures selection boundaries are within valid buffer bounds
func clampSelection(sel *Selection, buf [][]rune) {
    // clamp row bounds
    maxRow := len(buf) - 1
    if sel.StartRow < 0 { sel.StartRow = 0 }
    if sel.StartRow > maxRow { sel.StartRow = maxRow }
    if sel.EndRow < 0 { sel.EndRow = 0 }
    if sel.EndRow > maxRow { sel.EndRow = maxRow }

    // clamp columns for each row
    lineLenStart := len(buf[sel.StartRow])
    if sel.StartCol < 0 { sel.StartCol = 0 }
    if sel.StartCol > lineLenStart { sel.StartCol = lineLenStart }

    lineLenEnd := len(buf[sel.EndRow])
    if sel.EndCol < 0 { sel.EndCol = 0 }
    if sel.EndCol > lineLenEnd { sel.EndCol = lineLenEnd }
}
