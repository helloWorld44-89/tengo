package editor

import (
	"fmt"
	"strings"
)


var rowOffset int
var colOffset int


func drawTopBar(filename string, width int) string {
    //title := fmt.Sprintf("  %s — tenGo Quick Edit  ", filename)
	title := "  " +filename+ " | tenGo Quick Edit  "
	space:= (width - len(title))/2
	title = strings.Repeat("-", space) + title
    if len(title) < width {
        title += strings.Repeat("-", space)
    } else if len(title) > width {
        title = title[:width]
    }
    return "\x1b[7m" + title + "\x1b[0m"
}

func drawBottomBar(width int) string {
    shortcuts := " --      ^S Save  ^Q Quit  ^[or] + or - Line Tab   ^+Arrow Select  ^+C Copy  ^+V Paste  ^+X Cut  Alt+Arrow Move+     --"
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

	
	for screenRow := 0; screenRow < usableRows; screenRow++ {
		fileRow := rowOffset + screenRow

		if fileRow >= len(buf) {
			fmt.Print("\n")
			continue
		}

		line := buf[fileRow]
		
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
			sr, sc, er, ec = normalizeSelection(sel)

			if fileRow< sr || fileRow> er {
				fmt.Println(string(line))
				continue
			}

			// First or last row
			if fileRow== sr || fileRow== er {
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
		if fileRow> sr && fileRow< er {
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
    cursorScreenRow := (cur.Row - rowOffset) + 2
	fmt.Printf("\x1b[%d;%dH", cursorScreenRow, cur.Col+1)
}
//==========This is for the FULL Editor, not the quick editor.===========
func fulldrawTopBar(filename string, width int) string {
    //title := fmt.Sprintf("  %s — tenGo Quick Edit  ", filename)
	title := "  " +filename+ " | tenGo Quick Edit  "
	space:= (width - len(title))/2
	title = strings.Repeat("-", space) + title
    if len(title) < width {
        title += strings.Repeat("-", space)
    } else if len(title) > width {
        title = title[:width]
    }
    return "\x1b[7m" + title + "\x1b[0m"
}

func fulldrawBottomBar(width int) string {
    shortcuts := " --      ^S Save  ^Q Quit  ^[or] + or - Line Tab   ^+Arrow Select  ^+C Copy  ^+V Paste  ^+X Cut  Alt+Arrow Move+     --"
    if len(shortcuts) < width {
        shortcuts += strings.Repeat(" ", width-len(shortcuts))
    } else if len(shortcuts) > width {
        shortcuts = shortcuts[:width]
    }
    return "\x1b[7m" + shortcuts + "\x1b[0m"
}




// DrawPopup displays a centered modal window with the given title and lines.
// This blocks UNTIL ESC or q is pressed.
// func DrawPopup(title string, lines []string) {
//     rows, cols := getTerminalSize()

//     // Window size
//     w := cols / 2
//     h := rows / 2

//     // Top-left corner
//     x := (cols - w) / 2
//     y := (rows - h) / 2

//     // Dim the background
//     fmt.Print("\x1b[2m")

//     // Draw background box
//     for i := 0; i < h; i++ {
//        MoveCursor(y+i, x)
//         fmt.Print("\x1b[49m\x1b[37m" + strings.Repeat(" ", w))
//     }

//     // Reset dim for popup
//     fmt.Print("\x1b[22m")

//     // Draw border
//    MoveCursor(y, x)
//     fmt.Print("┌" + strings.Repeat("─", w-2) + "┐")

//     for i := 1; i < h-1; i++ {
//        MoveCursor(y+i, x)
//         fmt.Print("│")
//        MoveCursor(y+i, x+w-1)
//         fmt.Print("│")
//     }

//    MoveCursor(y+h-1, x)
//     fmt.Print("└" + strings.Repeat("─", w-2) + "┘")

//     // Title
//     if len(title) > w-4 {
//         title = title[:w-4]
//     }
//    MoveCursor(y, x+2)
//     fmt.Print(title)

//     // Content
//     for i, line := range lines {
//         if i >= h-2 {
//             break
//         }
//        MoveCursor(y+1+i, x+2)
//         if len(line) > w-4 {
//             line = line[:w-4]
//         }
//         fmt.Print(line)
//     }

//     // Footer / hint
//    MoveCursor(y+h-2, x+2)
//     fmt.Print("[ESC] close")

//     // Input loop (blocks until ESC or q)
//     for {
//         k := readKey()
//         if k == "esc" || k == "q" {
//             break
//         }
//     }

//     // Clear popup (restore screen)
//     redrawAfterPopup()
// }

// // Clears the popup area & redraws the editor
// func redrawAfterPopup() {
//     fmt.Print("\x1b[2J")
//     fmt.Print("\x1b[H")
// }
