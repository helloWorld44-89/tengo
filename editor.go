package main

func runQuickEditor(filePath string) {
	cursor := Cursor{0, 0}

	content, err := openFile(filePath)
	if err != nil {
		panic(err)
	}

	buf := toBuffer(content)
	status := "Editing"

	for {
		if !isPasting {
			draw(buf, cursor, filePath, status, &sel)
		}

		width, height := getTerminalSize()
		usableRows := height - 3
		width = width - colOffset

		key := readKey()

		switch key {

		case "ctrl-q", "esc":
			return

		case "up", "down", "left", "right":
			sel.Active = false
			moveCursor(&cursor, key, buf, &rowOffset, usableRows)

		case "tab":
			deleteSelection(&buf, &cursor, &sel)
			for i := 0; i < 4; i++ {
				insertRune(&buf, &cursor, ' ')
			}

		case "enter":
			deleteSelection(&buf, &cursor, &sel)
			insertNewline(&buf, &cursor)
			moveCursor(&cursor, "", buf, &rowOffset, usableRows)

		case "backspace":
			deleteSelection(&buf, &cursor, &sel)
			backspace(&buf, &cursor)

		case "ctrl-s":
			saveFile(filePath, buf)

		case "ctrl-[":
			removeLineTab(&buf, &cursor, &sel)

		case "ctrl-]":
			addLineTab(&buf, &cursor, &sel)

		// -------- CTRL + ARROWS --------
		case "ctrl-left":
			startSelectionIfNeeded(&sel, &cursor)
			if cursor.Col > 0 {
				cursor.Col--
			} else if cursor.Row > 0 {
				cursor.Row--
				cursor.Col = len(buf[cursor.Row])
			}
			updateSelection(&sel, &cursor)
			clampSelection(&sel, buf)

		case "ctrl-right":
			startSelectionIfNeeded(&sel, &cursor)
			lineLen := len(buf[cursor.Row])
			if cursor.Col < lineLen {
				cursor.Col++
			} else if cursor.Row < len(buf)-1 {
				cursor.Row++
				cursor.Col = 0
			}
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

		case "ctrl-c":
			copySelection(buf, &sel)

		case "ctrl-x":
			cutSelection(&buf, &cursor, &sel)

		case "paste-begin":
			inBracketedPaste = true

		case "paste-end":
			inBracketedPaste = false

		case "ctrl-v":
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