package main

func runQuickEditor(filePath string) {
    cursor := Cursor{0, 0}

    content, err := openFile(filePath)
    if err != nil { panic(err) }

    buf := toBuffer(content)
    status := "Editing"

    for {
        draw(buf, cursor, filePath, status, &sel)
        key := readKey()
	
        switch key {
        case "ctrl-q", "esc":
            return
        case "up", "down", "left", "right":
			sel.Active = false
            moveCursor(&cursor, key, buf)
        case "tab":
			sel.Active = false
            for i := 0; i < 4; i++ { insertRune(&buf, &cursor, ' ') }
        case "enter":
			sel.Active = false
            insertNewline(&buf, &cursor)
        case "backspace":
			sel.Active = false
            backspace(&buf, &cursor)
        case "ctrl-s":
            saveFile(filePath, buf)
        case "ctrl-[":
            removeLineTab(&buf, &cursor, &sel)
		case "ctrl-]":
			addLineTab(&buf, &cursor, &sel)
		
		case "ctrl-up":
			startSelectionIfNeeded(&sel, cursor)
			cursor.Row--
			updateSelection(&sel, cursor)

		case "ctrl-down":
			startSelectionIfNeeded(&sel, cursor)
			cursor.Row++
			updateSelection(&sel, cursor)

		case "ctrl-left":
			startSelectionIfNeeded(&sel, cursor)
			cursor.Col--
			updateSelection(&sel, cursor)

		case "ctrl-right":
			startSelectionIfNeeded(&sel, cursor)
			cursor.Col++
			updateSelection(&sel, cursor)

        default:
			sel.Active = false
            if len(key) == 1 && key[0] >= 32 {
                insertRune(&buf, &cursor, rune(key[0]))
            }
        }
    }
}