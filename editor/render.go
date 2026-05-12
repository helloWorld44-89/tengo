package editor

import (
	"fmt"
	"path/filepath"
	"strings"
)

// detectFileType returns a short label for the file type based on extension.
func detectFileType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".yaml", ".yml":
		return "YAML"
	case ".json":
		return "JSON"
	case ".toml":
		return "TOML"
	case ".ini", ".cfg":
		return "INI"
	case ".xml":
		return "XML"
	case ".go":
		return "Go"
	case ".md":
		return "Markdown"
	default:
		return "TXT"
	}
}

// drawTopBar renders the top header bar showing the current file name and editor title
func drawTopBar(filename string, width int) string {
    // Use blue background for top bar
    title := fmt.Sprintf("  📝 %s  |  tenGo Quick Editor  ", filename)
    if len(title) < width {
        title += strings.Repeat(" ", width-len(title))
    } else if len(title) > width {
        title = title[:width]
    }
    // Blue background (44) with white text (37)
    return "\x1b[44;37;1m" + title + "\x1b[0m"
}

// drawBottomBar renders the help/shortcut hint bar at the bottom of the editor
func drawBottomBar(width int) string {
	info := " 💡 Ctrl+H for help • Ctrl+S to save • Ctrl+Q to quit "
	if len(info) < width {
		info += strings.Repeat(" ", width-len(info))
	} else if len(info) > width {
		info = info[:width]
	}
	// Cyan background (46) with black text (30)
	return "\x1b[46;30m" + info + "\x1b[0m"
}

// drawStatusBar renders the status line showing file info, cursor position, file type, and transient messages.
func drawStatusBar(filename string, row, col, totalRows int, modified bool, msg, fileType, selInfo string, width int) string {
	modIndic := ""
	if modified {
		modIndic = " ●"
	}
	base := filepath.Base(filename)
	bar := fmt.Sprintf("  %s%s  •  Line %d/%d  •  Col %d  •  %s  ", base, modIndic, row+1, totalRows, col+1, fileType)
	if selInfo != "" {
		bar += "| " + selInfo + "  "
	} else if msg != "" && msg != "Editing" {
		bar += "| " + msg + "  "
	}
	if len(bar) < width {
		bar += strings.Repeat(" ", width-len(bar))
	} else if len(bar) > width {
		bar = bar[:width]
	}
	return "\x1b[47;30m" + bar + "\x1b[0m"
}

// draw renders the entire editor UI: top bar, content with selection highlighting, status bar, and cursor.
// errorLine is the 1-based line number to mark with a red ! in the gutter (0 = no error).
func draw(buf [][]rune, cur Cursor, filename string, status string, modified bool, rowOffset int, sel *Selection, showLineNumbers bool, errorLine int) {
    width, height := getTerminalSize()

    fmt.Print("\x1b[3J")
    fmt.Print("\x1b[2J")
    fmt.Print("\x1b[H")

    // === TOP BAR ===
    // Use \r\n throughout: MakeRaw clears OPOST/ONLCR so bare \n no longer
    // returns the cursor to column 1, causing a staircase.
    fmt.Printf("%s\r\n", drawTopBar(filename, width))

    usableRows := height - 3 // top bar + bottom bar + status line

    // Gutter layout:
    //   errColW  — 1 char for the "!" marker (shown whenever errorLine > 0)
    //   lineNumW — right-aligned number + 1 space (shown when showLineNumbers)
    errColW := 0
    if errorLine > 0 {
        errColW = 1
    }
    lineNumW := 0
    if showLineNumbers && len(buf) > 0 {
        lineNumW = len(fmt.Sprintf("%d", len(buf))) + 1
    }
    gutterW := errColW + lineNumW

    // File type and selection info for status bar.
    fileType := detectFileType(filename)
    selInfo := ""
    if sel.Active {
        sr2, sc2, er2, ec2 := normalizeSelection(sel)
        numLines := er2 - sr2 + 1
        chars := 0
        if sr2 == er2 {
            chars = ec2 - sc2
        } else {
            if sr2 < len(buf) {
                chars = len(buf[sr2]) - sc2
            }
            for r := sr2 + 1; r < er2; r++ {
                if r < len(buf) {
                    chars += len(buf[r]) + 1
                }
            }
            if er2 < len(buf) {
                chars += ec2
            }
        }
        selInfo = fmt.Sprintf("%dL · %dC selected", numLines, chars)
    }

    // Normalize selection once per frame for content rendering.
    var sr, sc, er, ec int
    if sel.Active {
        sr, sc, er, ec = normalizeSelection(sel)
    }

    // === FILE CONTENT ===
    for screenRow := 0; screenRow < usableRows; screenRow++ {
        fileRow := rowOffset + screenRow
        isCurLine := (fileRow == cur.Row)
        isErrLine := (errorLine > 0 && fileRow+1 == errorLine)

        if fileRow >= len(buf) {
            if gutterW > 0 {
                fmt.Print(strings.Repeat(" ", gutterW))
            }
            if isCurLine {
                fmt.Print("\x1b[48;5;236m\x1b[K\x1b[0m")
            }
            fmt.Print("\r\n")
            continue
        }

        line := buf[fileRow]
        lineLen := len(line)

        // Error marker column.
        if errColW > 0 {
            if isErrLine {
                fmt.Print("\x1b[31;1m!\x1b[0m")
            } else {
                fmt.Print(" ")
            }
        }
        // Line number column.
        if lineNumW > 0 {
            if isErrLine {
                fmt.Printf("\x1b[31m%*d \x1b[0m", lineNumW-1, fileRow+1)
            } else {
                fmt.Printf("\x1b[90m%*d \x1b[0m", lineNumW-1, fileRow+1)
            }
        }

        // Current line background.
        if isCurLine {
            fmt.Print("\x1b[48;5;236m")
        }

        // Render content with optional selection highlighting.
        if !sel.Active || fileRow < sr || fileRow > er {
            fmt.Print(string(line))
            if isCurLine {
                fmt.Print("\x1b[K")
            }
        } else {
            var hlStart, hlEnd int
            switch {
            case sr == er:
                hlStart, hlEnd = sc, ec
            case fileRow == sr:
                hlStart, hlEnd = sc, lineLen
            case fileRow == er:
                hlStart, hlEnd = 0, ec
            default:
                hlStart, hlEnd = 0, lineLen
            }
            if hlStart < 0 {
                hlStart = 0
            }
            if hlEnd > lineLen {
                hlEnd = lineLen
            }
            if hlStart > hlEnd {
                hlStart = hlEnd
            }

            fmt.Print(string(line[:hlStart]))
            fmt.Print("\x1b[7m")
            fmt.Print(string(line[hlStart:hlEnd]))
            fmt.Print("\x1b[27m")
            fmt.Print(string(line[hlEnd:]))
            if isCurLine {
                fmt.Print("\x1b[K")
            }
        }

        fmt.Print("\x1b[0m\r\n")
    }

    // === BOTTOM BAR ===
    fmt.Printf("\x1b[%d;1H", height-1)
    fmt.Print(drawBottomBar(width))

    // === STATUS BAR ===
    fmt.Printf("\x1b[%d;1H", height)
    fmt.Print(drawStatusBar(filename, cur.Row, cur.Col, len(buf), modified, status, fileType, selInfo, width))

    // === CURSOR (offset by gutter) ===
    cursorScreenRow := (cur.Row - rowOffset) + 2
    fmt.Printf("\x1b[%d;%dH", cursorScreenRow, cur.Col+gutterW+1)
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
