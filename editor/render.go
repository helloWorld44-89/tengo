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

// gutterWidth computes the combined width of the line-number and error-marker columns.
func gutterWidth(buf [][]rune, showLineNumbers bool, errorLine int) int {
	errColW := 0
	if errorLine > 0 {
		errColW = 1
	}
	lineNumW := 0
	if showLineNumbers && len(buf) > 0 {
		lineNumW = len(fmt.Sprintf("%d", len(buf))) + 1
	}
	return errColW + lineNumW
}

// adjustScrollWordWrap ensures the cursor is visible when word-wrap is active.
// It updates rowOffset so that the cursor's visual row falls within [0, usableRows).
func adjustScrollWordWrap(rowOffset *int, cur Cursor, buf [][]rune, usableRows, contentWidth int) {
	if cur.Row < *rowOffset {
		*rowOffset = cur.Row
		return
	}
	// Count visual rows from rowOffset up to (but not including) the cursor's line.
	visual := 0
	for r := *rowOffset; r < cur.Row && r < len(buf); r++ {
		l := len(buf[r])
		if l == 0 {
			visual++
		} else {
			visual += (l + contentWidth - 1) / contentWidth
		}
	}
	// Include the visual offset within the cursor's own line.
	visual += cur.Col / contentWidth
	// Scroll down until the cursor is within the visible area.
	for visual >= usableRows && *rowOffset < cur.Row {
		l := len(buf[*rowOffset])
		rowVis := 1
		if l > 0 {
			rowVis = (l + contentWidth - 1) / contentWidth
		}
		visual -= rowVis
		*rowOffset++
	}
}

// drawTopBar renders the top header bar showing the current file name and editor title
func drawTopBar(filename string, width int) string {
	title := fmt.Sprintf("  📝 %s  |  tenGo Quick Editor  ", filename)
	if len(title) < width {
		title += strings.Repeat(" ", width-len(title))
	} else if len(title) > width {
		title = title[:width]
	}
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

// drawTabBar renders a tab strip when multiple files are open.
// Each tab shows the base filename; the active tab is highlighted blue.
func drawTabBar(names []string, modified []bool, active, width int) string {
	var b strings.Builder
	visLen := 0

	for i, name := range names {
		base := filepath.Base(name)
		if len(base) > 22 {
			base = base[:19] + "..."
		}
		mod := ""
		if i < len(modified) && modified[i] {
			mod = " ●"
		}
		label := "  " + base + mod + "  "
		visLen += len(label)
		if i == active {
			b.WriteString("\x1b[44;37;1m" + label + "\x1b[0m")
		} else {
			b.WriteString("\x1b[48;5;238;37m" + label + "\x1b[0m")
		}
	}
	if visLen < width {
		b.WriteString("\x1b[48;5;238m" + strings.Repeat(" ", width-visLen) + "\x1b[0m")
	}
	return b.String()
}

// draw renders the entire editor UI: top bar, content with selection highlighting, status bar, and cursor.
// errorLine is the 1-based line number to mark with a red ! in the gutter (0 = no error).
// tabNames/tabModified/activeTab are used to render the tab bar when len(tabNames) > 1.
// wordWrap enables soft word-wrap: lines longer than the content width are displayed across multiple rows.
func draw(buf [][]rune, cur Cursor, filename string, status string, modified bool, rowOffset int, sel *Selection, showLineNumbers bool, errorLine int, tabNames []string, tabModified []bool, activeTab int, wordWrap bool) {
	width, height := getTerminalSize()

	fmt.Print("\x1b[3J")
	fmt.Print("\x1b[2J")
	fmt.Print("\x1b[H")

	fmt.Printf("%s\r\n", drawTopBar(filename, width))

	multiTab := len(tabNames) > 1
	if multiTab {
		fmt.Printf("%s\r\n", drawTabBar(tabNames, tabModified, activeTab, width))
	}

	fixedRows := 3
	if multiTab {
		fixedRows = 4
	}
	usableRows := height - fixedRows
	contentStart := 2
	if multiTab {
		contentStart = 3
	}

	errColW := 0
	if errorLine > 0 {
		errColW = 1
	}
	lineNumW := 0
	if showLineNumbers && len(buf) > 0 {
		lineNumW = len(fmt.Sprintf("%d", len(buf))) + 1
	}
	gutterW := errColW + lineNumW

	contentWidth := width - gutterW
	if contentWidth < 1 {
		contentWidth = 1
	}

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

	var sr, sc, er, ec int
	if sel.Active {
		sr, sc, er, ec = normalizeSelection(sel)
	}

	// === FILE CONTENT ===
	screenRow := 0
	fileRow := rowOffset
	for screenRow < usableRows {
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
			screenRow++
			fileRow++
			continue
		}

		line := buf[fileRow]
		lineLen := len(line)

		numChunks := 1
		if wordWrap && lineLen > contentWidth {
			numChunks = (lineLen + contentWidth - 1) / contentWidth
		}

		for chunk := 0; chunk < numChunks && screenRow < usableRows; chunk++ {
			chunkStart := 0
			chunkEnd := lineLen
			if wordWrap && numChunks > 1 {
				chunkStart = chunk * contentWidth
				chunkEnd = chunkStart + contentWidth
				if chunkEnd > lineLen {
					chunkEnd = lineLen
				}
			}

			// Error marker column.
			if errColW > 0 {
				if chunk == 0 && isErrLine {
					fmt.Print("\x1b[31;1m!\x1b[0m")
				} else {
					fmt.Print(" ")
				}
			}
			// Line number column.
			if lineNumW > 0 {
				if chunk == 0 {
					if isErrLine {
						fmt.Printf("\x1b[31m%*d \x1b[0m", lineNumW-1, fileRow+1)
					} else {
						fmt.Printf("\x1b[90m%*d \x1b[0m", lineNumW-1, fileRow+1)
					}
				} else {
					fmt.Print(strings.Repeat(" ", lineNumW))
				}
			}

			// Current line background.
			if isCurLine {
				fmt.Print("\x1b[48;5;236m")
			}

			chunkRunes := line[chunkStart:chunkEnd]

			if numChunks > 1 {
				// Wrapped line: no syntax highlighting to avoid ANSI offset issues.
				if sel.Active && fileRow >= sr && fileRow <= er {
					var lineHlStart, lineHlEnd int
					switch {
					case sr == er:
						lineHlStart, lineHlEnd = sc, ec
					case fileRow == sr:
						lineHlStart, lineHlEnd = sc, lineLen
					case fileRow == er:
						lineHlStart, lineHlEnd = 0, ec
					default:
						lineHlStart, lineHlEnd = 0, lineLen
					}
					hlStart := lineHlStart - chunkStart
					hlEnd := lineHlEnd - chunkStart
					chunkLen := chunkEnd - chunkStart
					if hlStart < 0 {
						hlStart = 0
					}
					if hlEnd > chunkLen {
						hlEnd = chunkLen
					}
					if hlStart > hlEnd {
						hlStart = hlEnd
					}
					if hlStart < hlEnd {
						fmt.Print(string(chunkRunes[:hlStart]))
						fmt.Print("\x1b[7m")
						fmt.Print(string(chunkRunes[hlStart:hlEnd]))
						fmt.Print("\x1b[27m")
						fmt.Print(string(chunkRunes[hlEnd:]))
					} else {
						fmt.Print(string(chunkRunes))
					}
				} else {
					fmt.Print(string(chunkRunes))
				}
			} else {
				// Non-wrapped: syntax highlighting + selection.
				if !sel.Active || fileRow < sr || fileRow > er {
					fmt.Print(highlightLine(fileType, string(line)))
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
			}

			if isCurLine && numChunks > 1 {
				fmt.Print("\x1b[K")
			}
			fmt.Print("\x1b[0m\r\n")
			screenRow++
		}

		fileRow++
	}

	// === BOTTOM BAR ===
	fmt.Printf("\x1b[%d;1H", height-1)
	fmt.Print(drawBottomBar(width))

	// === STATUS BAR ===
	visCol := 0
	if cur.Row < len(buf) {
		visCol = visualCol(buf[cur.Row], cur.Col, 4)
	}
	fmt.Printf("\x1b[%d;1H", height)
	fmt.Print(drawStatusBar(filename, cur.Row, visCol, len(buf), modified, status, fileType, selInfo, width))

	// === CURSOR ===
	if wordWrap && contentWidth > 0 {
		cursorScreenRow := contentStart
		for r := rowOffset; r < cur.Row && r < len(buf); r++ {
			l := len(buf[r])
			if l == 0 {
				cursorScreenRow++
			} else {
				cursorScreenRow += (l + contentWidth - 1) / contentWidth
			}
		}
		cursorScreenRow += cur.Col / contentWidth
		fmt.Printf("\x1b[%d;%dH", cursorScreenRow, (cur.Col%contentWidth)+gutterW+1)
	} else {
		cursorScreenRow := (cur.Row-rowOffset) + contentStart
		fmt.Printf("\x1b[%d;%dH", cursorScreenRow, cur.Col+gutterW+1)
	}
}

//==========This is for the FULL Editor, not the quick editor.===========
func fulldrawTopBar(filename string, width int) string {
	title := "  " + filename + " | tenGo Quick Edit  "
	space := (width - len(title)) / 2
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
