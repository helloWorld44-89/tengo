package main

import (
	"os"
) 



type Selection struct {
    Active    bool
    StartRow  int
    StartCol  int
    EndRow    int
    EndCol    int
}


var sel Selection



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
    case 3:  return "ctrl-c"
    case 22: return "ctrl-v"
    case 24: return "ctrl-x"
    case 19: return "ctrl-s"
    case 17: return "ctrl-q"
    case 9:  return "tab"
    case 127:return "backspace"
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
    }

    // === SCROLLING (VERTICAL) ===
    if c.Row < *rowOffset {
        *rowOffset = c.Row
    }

    if c.Row >= *rowOffset+screenRows {
        *rowOffset = c.Row - screenRows + 1
    }
}



