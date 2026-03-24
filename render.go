package main

import (
	"strings"
	"fmt")


func drawTopBar(filename string, width int) string {
    //title := fmt.Sprintf("  %s — tenGo Quick Edit  ", filename)
	title := "  " +filename+ " — tenGo Quick Edit  "
	space:= (width - len(title))/2
	title = strings.Repeat("*", space) + title
    if len(title) < width {
        title += strings.Repeat("*", space)
    } else if len(title) > width {
        title = title[:width]
    }
    return "\x1b[7m" + title + "\x1b[0m"
}

func drawBottomBar(width int) string {
    shortcuts := "  ^S Save   ^Q Quit   ^[ Del Line Tab   ^] Add Line Tab   Ctrl+Arrow Select"
    if len(shortcuts) < width {
        shortcuts += strings.Repeat(" ", width-len(shortcuts))
    } else if len(shortcuts) > width {
        shortcuts = shortcuts[:width]
    }
    return "\x1b[7m" + shortcuts + "\x1b[0m"
}

func draw(buf [][]rune, cur Cursor, filename string, status string, sel *Selection) {
    width, height := getTerminalSize()

    // clears window and scrollback before opening
    fmt.Print("\x1b[3J")
    fmt.Print("\x1b[2J")
    fmt.Print("\x1b[H")

    // === TOP BAR at row 1 ===
    fmt.Println(drawTopBar(filename, width))

    // === FILE CONTENT ===
    usableRows := height - 3 // top bar + bottom bar + status line
  	
	for i := 0; i < usableRows && i < len(buf); i++ {
		line := buf[i]

		// If no selection or selection is inactive
		if !sel.Active {
			fmt.Println(string(line))
			continue
		}

		startRow := sel.StartRow
		endRow := sel.EndRow
		startCol := sel.StartCol
		endCol := sel.EndCol

		if startRow > endRow || (startRow == endRow && startCol > endCol) {
			// normalize (swap)
			startRow, endRow = endRow, startRow
			startCol, endCol = endCol, startCol
		}

		// Case 1: Row not in selection
		if i < startRow || i > endRow {
			fmt.Println(string(line))
			continue
		}

		// Case 2: Full line highlight (middle lines)
		if i > startRow && i < endRow {
			fmt.Print("\x1b[7m")              // reverse
			fmt.Print(string(line))
			fmt.Print("\x1b[0m\n")
			continue
		}

		// Case 3: First or last line: highlight only part
		left := line[:startCol]
		mid := line[startCol:endCol]
		right := line[endCol:]

		fmt.Print(string(left))
		fmt.Print("\x1b[7m")                 // highlight mid
		fmt.Print(string(mid))
		fmt.Print("\x1b[0m")                 // back to normal
		fmt.Print(string(right))
		fmt.Print("\n")
	}

    // === BOTTOM BAR at row height-1 ===
    fmt.Printf("\x1b[%d;1H", height-1)
    fmt.Print(drawBottomBar(width))

    // === STATUS LINE at bottom row ===
    fmt.Printf("\x1b[%d;1H", height)
    fmt.Print(status)

    // === CURSOR POSITION (offset by +1 because of top bar) ===
    fmt.Printf("\x1b[%d;%dH", cur.Row+2, cur.Col+1)

	


}