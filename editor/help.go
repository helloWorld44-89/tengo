package editor

import (
	"fmt"
	"strings"
)

// ShowPopup displays a centered modal window and waits for user input to close
func ShowPopup(title string, lines []string) {
    cols, rows := getTerminalSize()

    // Calculate popup size based on content and terminal
    w := len(title) + 10
    for _, line := range lines {
        if len(line)+4 > w {
            w = len(line) + 4
        }
    }
    
    // Clamp to terminal constraints
    maxW := cols - 4
    if w > maxW { w = maxW }
    if w < 40 { w = 40 }
    
    h := len(lines) + 5
    maxH := rows - 3
    if h > maxH { h = maxH }
    if h < 10 { h = 10 }
    
    x := (cols - w) / 2
    if x < 1 { x = 1 }
    y := (rows - h) / 2
    if y < 1 { y = 1 }

    // Clear screen and draw semi-transparent background overlay
    fmt.Print("\x1b[?25l") // Hide cursor
    clearAndDrawOverlay(rows, cols)

    // Draw popup window on top
    drawPopupWindow(x, y, w, h, title, lines)

    // Wait for input (ESC, q, or Enter to close)
    for {
        k := readKey()
        if k == "esc" || k == "q" || k == "enter" {
            break
        }
    }
    
    fmt.Print("\x1b[?25h") // Show cursor
}

// clearAndDrawOverlay clears the screen for popup display
func clearAndDrawOverlay(rows, cols int) {
    // Clear screen and move cursor to home
    fmt.Print("\x1b[2J\x1b[H")
}

// drawPopupWindow draws a properly aligned popup with borders and background
func drawPopupWindow(x, y, w, h int, title string, lines []string) {
    // Ensure valid coordinates
    if x < 1 { x = 1 }
    if y < 1 { y = 1 }
    if w < 10 { w = 10 }
    if h < 5 { h = 5 }

    // Draw white background for entire popup area
    for row := 0; row < h; row++ {
        MoveCursor(y+row, x)
        // White background (47) + black text (30) + bold (1)
        fmt.Print("\x1b[47;30;1m" + strings.Repeat(" ", w) + "\x1b[0m")
    }

    // Top border
    MoveCursor(y, x)
    topBorder := "┌" + strings.Repeat("─", w-2) + "┐"
    fmt.Print("\x1b[47;30;1m" + topBorder + "\x1b[0m")

    // Title line
    MoveCursor(y+1, x)
    titleText := title
    if len(titleText) > w-4 {
        titleText = titleText[:w-4]
    }
    padding := w - 4 - len(titleText)
    titleLine := "│ " + titleText + strings.Repeat(" ", padding) + " │"
    fmt.Print("\x1b[47;30;1m" + titleLine + "\x1b[0m")

    // Separator
    MoveCursor(y+2, x)
    separator := "├" + strings.Repeat("─", w-2) + "┤"
    fmt.Print("\x1b[47;30;1m" + separator + "\x1b[0m")

    // Content lines
    maxLines := h - 4
    for i := 0; i < maxLines; i++ {
        MoveCursor(y+3+i, x)
        line := ""
        if i < len(lines) {
            line = lines[i]
            if len(line) > w-4 {
                line = line[:w-4]
            }
        }
        padding := w - 4 - len(line)
        contentLine := "│ " + line + strings.Repeat(" ", padding) + " │"
        fmt.Print("\x1b[47;30;1m" + contentLine + "\x1b[0m")
    }

    // Footer separator
    MoveCursor(y+h-2, x)
    footerSep := "├" + strings.Repeat("─", w-2) + "┤"
    fmt.Print("\x1b[47;30;1m" + footerSep + "\x1b[0m")

    // Bottom bar with instructions
    MoveCursor(y+h-1, x)
    footer := "[ESC/Q/Enter] Close"
    if len(footer) > w-4 {
        footer = footer[:w-4]
    }
    footerPadding := w - 4 - len(footer)
    footerLine := "│ " + footer + strings.Repeat(" ", footerPadding) + " │"
    fmt.Print("\x1b[47;30;1m" + footerLine + "\x1b[0m")

    // Bottom border
    MoveCursor(y+h, x)
    bottomBorder := "└" + strings.Repeat("─", w-2) + "┘"
    fmt.Print("\x1b[47;30;1m" + bottomBorder + "\x1b[0m")
}

// Move cursor helper
func MoveCursor(r, c int) {
    fmt.Printf("\x1b[%d;%dH", r, c)
}

// ShowHelp displays a popup overlay listing all keyboard shortcuts.
func ShowHelp() {
    ShowPopup("Quick Editor Keyboard Shortcuts", []string{
        "Navigation:",
        "  Arrow keys         Move cursor",
        "  Home / End         Start / end of line",
        "  Ctrl+Arrows        Word jump + select",
        "  Alt+Arrows         Fast navigation (4 steps)",
        "",
        "Editing:",
        "  Tab                Insert 4 spaces",
        "  Enter              New line",
        "  Shift+Enter        Insert line below",
        "  Ctrl+Shift+Enter   Insert line above",
        "  Backspace          Delete character",
        "  Ctrl+D             Duplicate line",
        "  Ctrl+/             Toggle comment",
        "  Ctrl+[             Decrease indent",
        "  Ctrl+]             Increase indent",
        "",
        "Undo/Redo:",
        "  Ctrl+Z             Undo",
        "  Ctrl+Y             Redo",
        "",
        "Selection & Search:",
        "  Ctrl+A             Select all",
        "  Ctrl+F             Find",
        "  Ctrl+R             Find & Replace",
        "  Ctrl+C             Copy",
        "  Ctrl+X             Cut",
        "  Ctrl+V             Paste",
        "",
        "File & Validation:",
        "  Ctrl+S             Save (auto-validates)",
        "  Ctrl+E             Validate syntax",
        "  Ctrl+T             Format / pretty-print",
        "  Ctrl+G             Go to line",
        "  Ctrl+N             Toggle line numbers",
        "  Ctrl+Q or Esc      Quit",
        "  Ctrl+H             Show this help",
    })
}