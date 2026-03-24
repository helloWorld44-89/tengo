package main
import "os"



type Selection struct {
    Active    bool
    StartRow  int
    StartCol  int
    EndRow    int
    EndCol    int
}


var sel Selection


func readKey() string {
    buf := make([]byte, 8)
    n, _ := os.Stdin.Read(buf)

    // CTRL keys
    if buf[0] == 17 { return "ctrl-q" } // Ctrl+Q
    if buf[0] == 19 { return "ctrl-s" } // Ctrl+S

    // ENTER
    if buf[0] == '\r' {
        return "enter"
    }
	
  	// TAB
    if buf[0] == 9 {
        return "tab"
    }
    // Ctrl+[
        if buf[0] == 27 && n==1{
        return "ctrl-["
    }

	// Ctrl+]
        if buf[0] == 29 && n==1{
        return "ctrl-]"
    }

    // BACKSPACE
    if buf[0] == 127 {
        return "backspace"
    }

	//ctrl + arrow for rapid navigation
	if buf[0] == 27 && n == 6 && buf[1] == '[' && buf[2] == '1' && buf[3] == ';' && buf[4] == '5' {
		switch buf[5] {
		case 'A':
			return "ctrl-up"
		case 'B':
			return "ctrl-down"
		case 'C':
			return "ctrl-right"
		case 'D':
			return "ctrl-left"
		}
	}

    // ESC or ARROW KEYS
    if buf[0] == 27 {
        if n == 1 {
            return "esc"
        }
        if buf[1] == 91 {
            switch buf[2] {
            case 'A':
                return "up"
            case 'B':
                return "down"
            case 'C':
                return "right"
            case 'D':
                return "left"
            }
        }
    }
	
	if buf[0] == 19 { // CTRL-S
		return "ctrl-s"
	}

    // Normal character (letters, numbers, symbols)
    if n == 1 {
        return string(buf[0])
    }

    return ""
}




func moveCursor(c *Cursor, key string, buf [][]rune) {
    switch key {
    case "up":
        if c.Row > 0 {
            c.Row--
            // clamp
            if c.Col > len(buf[c.Row]) {
                c.Col = len(buf[c.Row])
            }
        }

    case "down":
        if c.Row < len(buf)-1 {
            c.Row++
            // clamp
            if c.Col > len(buf[c.Row]) {
                c.Col = len(buf[c.Row])
            }
        }

    case "left":
        if c.Col > 0 {
            c.Col--
        } else if c.Row > 0 {
            c.Row--
            c.Col = len(buf[c.Row])
        }

    case "right":
        if c.Col < len(buf[c.Row]) {
            c.Col++
        } else if c.Row < len(buf)-1 {
            c.Row++
            c.Col = 0
        }
    }
}



