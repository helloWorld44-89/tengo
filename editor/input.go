package editor

import (
	"fmt"
	"os"
	"strings"
)



type Selection struct {
    Active    bool
    StartRow  int
    StartCol  int
    EndRow    int
    EndCol    int
}


func readEscSequence(first byte) string {
    seq := []byte{first}
    buf := make([]byte, 1)

    // Keep reading until we hit a letter or "~"
    for {
        n, _ := os.Stdin.Read(buf)
        if n == 0 {
            break
        }
        seq = append(seq, buf[0])

        if (buf[0] >= 'A' && buf[0] <= 'Z') ||
           (buf[0] >= 'a' && buf[0] <= 'z') ||
            buf[0] == '~' {
            break
        }
    }

    return string(seq)
}



func readKey() string {
    buf := make([]byte, 1)

    // Read first byte
    n, err := os.Stdin.Read(buf)
    if err != nil || n == 0 {
        return ""
    }

    b := buf[0]

    // ============================================
    // 1. Single-byte controls
    // ============================================
    switch b {
    case 1:  return "ctrl-a"
    case 3:  return "ctrl-c"
    case 4:  return "ctrl-d"
    case 6:  return "ctrl-f"
    case 8:  return "ctrl-h"
    case 9:  return "tab"
    case 14: return "ctrl-n"
    case 17: return "ctrl-q"
    case 18: return "ctrl-r"
    case 19: return "ctrl-s"
    case 22: return "ctrl-v"
    case 24: return "ctrl-x"
    case 23: return "ctrl-w"
    case 25: return "ctrl-y"
    case 26: return "ctrl-z"
    case 29: return "ctrl-]"
    case 31: return "ctrl-/"
    case 127: return "backspace"
    case '\r': return "enter"
    }

    // ============================================
    // 2. Printable characters
    // ============================================
    if b != 27 { // not ESC
        return string([]byte{b})
    }

    // ============================================
    // 3. ESC SEQUENCE — read the full sequence
    // ============================================
    seq := []byte{27}

    // Read until final byte of an escape sequence
    for {
        n, err := os.Stdin.Read(buf)
        if err != nil || n == 0 {
            break
        }
        seq = append(seq, buf[0])

        c := buf[0]

        // Final bytes of ESC sequences: letter or '~'
        if (c >= 'A' && c <= 'Z') ||
           (c >= 'a' && c <= 'z') ||
            c == '~' {
            break
        }
    }

    s := string(seq)

    // ============================================
    // 4. Bracketed paste mode
    // ============================================
    if s == "\x1b[200~" {
        return "paste-begin"
    }
    if s == "\x1b[201~" {
        return "paste-end"
    }

    // ============================================
    // 5. Arrow keys
    // ============================================
    switch s {
    case "\x1b[A": return "up"
    case "\x1b[B": return "down"
    case "\x1b[C": return "right"
    case "\x1b[D": return "left"
    }
    // ============================================
    // 6. ALT + Arrow
    // ============================================
    

    if strings.HasPrefix(s, "\x1b[1;3") {
        switch s[len(s)-1] {
        case 'A':
            return "alt-up"
        case 'B':
            return "alt-down"
        case 'C':
            return "alt-right"
        case 'D':
            return "alt-left"
        }
    }


    

    // ============================================
    // 6. Ctrl + Arrow
    // ============================================
    switch s {
    case "\x1b[1;5A": return "ctrl-up"
    case "\x1b[1;5B": return "ctrl-down"
    case "\x1b[1;5C": return "ctrl-right"
    case "\x1b[1;5D": return "ctrl-left"
    }

    // ============================================
    // 7. Home / End
    // ============================================
    switch s {
    case "\x1b[H", "\x1b[1~":
        return "home"
    case "\x1b[F", "\x1b[4~":
        return "end"
    }

    // ============================================
    // 8. Page Up / Page Down
    // ============================================
    switch s {
    case "\x1b[5~": return "page-up"
    case "\x1b[6~": return "page-down"
    }

    // ============================================
    // 9. ALT + key
    // ============================================
    if len(s) == 2 && s[0] == 27 {
        return "alt-" + string(s[1])
    }

    // Fallback
    return s
}


func moveCursor(c *Cursor, key string, buf [][]rune, rowOffset *int, screenRows int) {

    switch key {

    case "up":
        if c.Row > 0 {
            c.Row--
            if c.Col > len(buf[c.Row]) {
                c.Col = len(buf[c.Row])
            }
        }

    case "down":
        if c.Row < len(buf)-1 {
            c.Row++
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
    
    case "alt-left":
        // Move Left by 4 spaces
        for range 4 {
            if c.Col > 0 {
                c.Col--
            }  else {
                break
            }
        }

    
    case "alt-right":
        // Move Right by 4 spaces
        for range 4 {
            if c.Col < len(buf[c.Row]) {
                c.Col++
            }  else {
                break
            }
        }
    case "alt-up":
        // Move Up by 4 lines
        for range 4 {
            if c.Row > 0 {
                c.Row--
                if c.Col > len(buf[c.Row]) {
                    c.Col = len(buf[c.Row])
                }
            } else {
                break
            }
        }
    case "alt-down":
        // Move Down by 4 lines
        for range 4 {
            if c.Row < len(buf)-1 {
                c.Row++
                if c.Col > len(buf[c.Row]) {
                    c.Col = len(buf[c.Row])
                }
            } else {
                break
            }
        }


    }

    // === SCROLLING (VERTICAL) ===
    if c.Row < *rowOffset {
        *rowOffset = c.Row
    }

    if c.Row >= *rowOffset+screenRows {
        *rowOffset = c.Row - screenRows + 1
    }
}

func isWordChar(r rune) bool {
    return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// moveWordRight jumps to the start of the next word on the current line,
// or to the start of the next line if already at or past the end.
func moveWordRight(cur *Cursor, buf [][]rune) {
    line := buf[cur.Row]
    col := cur.Col

    if col >= len(line) {
        if cur.Row < len(buf)-1 {
            cur.Row++
            cur.Col = 0
        }
        return
    }

    // Skip the current run of word characters.
    for col < len(line) && isWordChar(line[col]) {
        col++
    }
    // Skip any trailing non-word characters (punctuation, spaces).
    for col < len(line) && !isWordChar(line[col]) {
        col++
    }
    cur.Col = col
}

// moveWordLeft jumps to the start of the current or previous word.
func moveWordLeft(cur *Cursor, buf [][]rune) {
    if cur.Col == 0 {
        if cur.Row > 0 {
            cur.Row--
            cur.Col = len(buf[cur.Row])
        }
        return
    }

    line := buf[cur.Row]
    col := cur.Col - 1

    // Skip non-word characters to the left (spaces, punctuation).
    for col > 0 && !isWordChar(line[col]) {
        col--
    }
    // Skip word characters to the left to reach the start of the word.
    for col > 0 && isWordChar(line[col-1]) {
        col--
    }
    cur.Col = col
}

// readPrompt renders a prompt in the status bar and collects typed input via
// readKey so it works correctly in raw terminal mode.
// Returns (input, true) on Enter, or ("", false) if the user pressed Esc/ctrl-q.
func readPrompt(prompt string) (string, bool) {
    width, height := getTerminalSize()
    var result []rune

    for {
        // Render prompt + typed text in the bottom status bar row.
        fmt.Printf("\x1b[%d;1H", height)
        display := prompt + string(result)
        if len(display) < width {
            display += strings.Repeat(" ", width-len(display))
        } else if len(display) > width {
            display = display[:width]
        }
        fmt.Printf("\x1b[47;30m%s\x1b[0m", display)

        // Position the cursor right after the typed text.
        col := len(prompt) + len(result) + 1
        if col > width {
            col = width
        }
        fmt.Printf("\x1b[%d;%dH", height, col)

        key := readKey()
        switch key {
        case "enter":
            return string(result), true
        case "esc", "ctrl-q":
            return "", false
        case "backspace":
            if len(result) > 0 {
                result = result[:len(result)-1]
            }
        default:
            if len(key) == 1 && key[0] >= 32 {
                result = append(result, rune(key[0]))
            }
        }
    }
}

