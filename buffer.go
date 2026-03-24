package main

import "strings"


type Cursor struct {
	Row int
	Col int
}

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

func toBuffer(content string) [][]rune {
	lines :=strings.Split(content, "\n")
	buf := make([][]rune,len(lines))
	for i, line := range lines {
		buf[i]=[]rune(line)
	}
	return buf
}


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




func insertNewline(buf *[][]rune, cur *Cursor) {
    line := (*buf)[cur.Row]

    left := append([]rune{}, line[:cur.Col]...)
    right := append([]rune{}, line[cur.Col:]...)

    (*buf)[cur.Row] = left

    *buf = append((*buf)[:cur.Row+1],
        append([][]rune{right}, (*buf)[cur.Row+1:]...)...)

    cur.Row++
    cur.Col = 0
}

func insertRune(buf *[][]rune, cur *Cursor, r rune) {
    line := (*buf)[cur.Row]
    newLine := append(line[:cur.Col], append([]rune{r}, line[cur.Col:]...)...)
    (*buf)[cur.Row] = newLine
    cur.Col++
}


func startSelectionIfNeeded(sel *Selection, cur Cursor) {
    if !sel.Active {
        sel.Active = true
        sel.StartRow = cur.Row
        sel.StartCol = cur.Col
    }
}


func updateSelection(sel *Selection, cur Cursor) {
    sel.EndRow = cur.Row
    sel.EndCol = cur.Col
}

func clearSelection(sel *Selection) {
    sel.Active = false
}


func normalizeSelection(sel *Selection) (sr, sc, er, ec int) {
    sr, sc = sel.StartRow, sel.StartCol
    er, ec = sel.EndRow, sel.EndCol

    if sr > er || (sr == er && sc > ec) {
        sr, er = er, sr
        sc, ec = ec, sc
    }
    return
}

