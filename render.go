package main

import (
	"fmt"
	"strings"
)


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
    shortcuts := "  ^S Save   ^Q Quit   ^[ Del Line Tab   ^] Add Line Tab   Ctrl+Arrow Select  Ctrl+C Copy  Ctrl+V Paste  Ctrl+X Cut  "
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
		
		lineLen := len(line)
		if sc < 0 { sc = 0 }
		if ec < 0 { ec = 0 }
		if sc > lineLen { sc = lineLen }
		if ec > lineLen { ec = lineLen }

		if !sel.Active {
			fmt.Println(string(line))
			continue
		}

		// Use normalized selection only
		if sel.Active {
			sr, sc, er, ec := normalizeSelection(sel)

			if i < sr || i > er {
				fmt.Println(string(line))
				continue
			}

			// First or last row
			if i == sr || i == er {
				lineLen := len(line)

				if sc < 0 { sc = 0 }
				if ec < 0 { ec = 0 }
				if sc > lineLen { sc = lineLen }
				if ec > lineLen { ec = lineLen }

				left := line[:sc]
				mid  := line[sc:ec]
				right := line[ec:]
				fmt.Print(string(left))
				fmt.Print("\x1b[7m", string(mid), "\x1b[0m")
				fmt.Print(string(right))
				fmt.Print("\n")
				continue
			}

			// Middle rows
			fmt.Print("\x1b[7m", string(line), "\x1b[0m\n")
			continue
		}

		// NO SELECTION → JUST PRINT
		fmt.Println(string(line))


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
