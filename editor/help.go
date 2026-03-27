package editor

import (
	"fmt"
	"strings"
)

// Draws and waits for ESC
func ShowPopup(title string, lines []string) {
    rows, cols := getTerminalSize()

    w := cols / 2
    h := rows / 2
    x := (cols - w) / 2
    y := (rows - h) / 2

    // Draw a semi-transparent overlay
    fmt.Print("\x1b[2m")
    for i := range rows {
        MoveCursor(i+1, 1)
        fmt.Print(strings.Repeat(" ", cols))
    }
    fmt.Print("\x1b[22m")

    // Draw popup window
    drawPopupWindow(x, y, w, h, title, lines)

    // Wait for ESC
    for {
        k := readKey()
        if k == "esc" {
            break
        }
    }
}

func drawPopupWindow(x, y, w, h int, title string, lines []string) {
    // Top border
    MoveCursor(y, x)
    fmt.Print("┌" + strings.Repeat("─", w-2) + "┐")

    // Sides + background
    for i := 0; i < h-1; i++ {
        MoveCursor(y+i, x)
        fmt.Print("│" + strings.Repeat(" ", w-2) + "│")
    }

    // Bottom border
    MoveCursor(y+h-1, x)
    fmt.Print("└" + strings.Repeat("─", w-2) + "┘")

    // Title
    MoveCursor(y, x+2)
    if len(title) > w-4 {
        title = title[:w-4]
    }
    fmt.Print(title)

    // Content lines
    for i, line := range lines {
        if i >= h-3 {
            break
        }
        MoveCursor(y+1+i, x+2)
        if len(line) > w-4 {
            line = line[:w-4]
        }
        fmt.Print(line)
    }

    // Footer
    MoveCursor(y+h-2, x+2)
    fmt.Print("[ESC] Close")
}

// Move cursor helper
func MoveCursor(r, c int) {
    fmt.Printf("\x1b[%d;%dH", r, c)
}