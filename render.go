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

    // === TOP BAR ===
    fmt.Println(drawTopBar(filename, width))

    usableRows := height - 3 // top bar + bottom bar + status line

    // STEP 2: Normalize selection ONCE here
    var sr, sc, er, ec int
    if sel.Active {
        sr, sc, er, ec = normalizeSelection(sel)
    }
	


    // === FILE CONTENT ===

	for i := 0; i < usableRows && i < len(buf); i++ {
		line := buf[i]

		if !sel.Active {
			fmt.Println(string(line))
			continue
		}

		// Use normalized selection only
		if i < sr || i > er {
			fmt.Println(string(line))
			continue
		}

		// Middle lines (full highlight)
		if i > sr && i < er {
			fmt.Print("\x1b[7m", string(line), "\x1b[0m\n")
			continue
		}

		// First or last selected line (column highlight)
		left  := line[:sc]
		mid   := line[sc:ec]
		right := line[ec:]

		fmt.Print(string(left))
		fmt.Print("\x1b[7m", string(mid), "\x1b[0m")
		fmt.Print(string(right))
		fmt.Print("\n")
	}

    // === BOTTOM BAR ===
    fmt.Printf("\x1b[%d;1H", height-1)
    fmt.Print(drawBottomBar(width))

    // === STATUS BAR ===
    fmt.Printf("\x1b[%d;1H", height)
    fmt.Print(status)

    // === CURSOR ===
    fmt.Printf("\x1b[%d;%dH", cur.Row+2, cur.Col+1)
}
