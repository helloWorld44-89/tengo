package main

import "fmt"

func runQuickEditor(filePath string) {
    cursor := Cursor{0, 0}

    content, err := openFile(filePath)
    if err != nil { panic(err) }

    buf := toBuffer(content)
    status := "Editing"

    for {
        draw(buf, cursor, filePath, status, &sel)
        key := readKey()
        fmt.Println("KEY:", key)
	
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
            // Step 1: anchor BEFORE movement
            if !sel.Active {
                sel.Active = true
                sel.StartRow = cursor.Row
                sel.StartCol = cursor.Col
            }

            // Move cursor UP
            if cursor.Row > 0 {
                cursor.Row--
            }

            // Update selection end
            sel.EndRow = cursor.Row
            sel.EndCol = cursor.Col



        case "ctrl-down":
            // Step 1: anchor BEFORE movement
            if !sel.Active {
                sel.Active = true
                sel.StartRow = cursor.Row
                sel.StartCol = cursor.Col
            }

            // Move cursor DOWN
            if cursor.Row < len(buf)-1 {
                cursor.Row++
            }

            // Update selection end
            sel.EndRow = cursor.Row
            sel.EndCol = cursor.Col

		case "ctrl-left":
			startSelectionIfNeeded(&sel, cursor)
			cursor.Col--
			updateSelection(&sel, cursor)

		case "ctrl-right":
			startSelectionIfNeeded(&sel, cursor)
			cursor.Col++
			updateSelection(&sel, cursor)

        case "ctrl-c": 
            copySelection(buf, &sel)

        case "ctrl-x": 
            cutSelection(&buf, &cursor, &sel)

        case "ctrl-v": 
            pasteText(&buf, &cursor)


        default:
			sel.Active = false
            if len(key) == 1 && key[0] >= 32 {
                insertRune(&buf, &cursor, rune(key[0]))
            }
        }
    }
}