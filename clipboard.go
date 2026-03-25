package main

import (
    "bytes"
    "os/exec"
    "runtime"
)

var internalClipboard string

// --- Copy string to system clipboard if possible ---
func copyToClipboard(text string) {
    internalClipboard = text // always store internally

    switch runtime.GOOS {
    case "darwin": // macOS
        cmd := exec.Command("pbcopy")
        cmd.Stdin = bytes.NewBufferString(text)
        _ = cmd.Run()

    case "linux":
        // try xclip
        cmd := exec.Command("xclip", "-selection", "clipboard")
        cmd.Stdin = bytes.NewBufferString(text)
        if cmd.Run() == nil {
            return
        }
        // try xsel
        cmd = exec.Command("xsel", "--clipboard", "--input")
        cmd.Stdin = bytes.NewBufferString(text)
        _ = cmd.Run()

    case "windows":
        cmd := exec.Command("clip")
        cmd.Stdin = bytes.NewBufferString(text)
        _ = cmd.Run()
    }
}

// --- Paste string from system clipboard (fallback to internal) ---
func pasteFromClipboard() string {

    switch runtime.GOOS {
    case "darwin":
        out, err := exec.Command("pbpaste").Output()
        if err == nil {
            return string(out)
        }

    case "linux":
        out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
        if err == nil {
            return string(out)
        }
        out, err = exec.Command("xsel", "--clipboard", "--output").Output()
        if err == nil {
            return string(out)
        }

    case "windows":
        out, err := exec.Command("powershell", "-command", "Get-Clipboard").Output()
        if err == nil {
            return string(out)
        }
    }

    return internalClipboard
}


func getSelectedText(buf [][]rune, sel *Selection) string {
    
    sr, sc, er, ec := normalizeSelection(sel)

    // Multi-line → treat as whole-line copy
    if sr != er {
        sc = 0
        // ec handled per-line
    }


    wholeLine := (sc == ec)

    if sr == er {
        if wholeLine {
            return string(buf[sr]) + "\n"
        }
        return string(buf[sr][sc:ec])
    }

    var out string

    if wholeLine {
        for i := sr; i <= er; i++ {
            out += string(buf[i]) + "\n"
        }
        return out
    }

    out += string(buf[sr][sc:]) + "\n"

    for i := sr + 1; i < er; i++ {
        out += string(buf[i]) + "\n"
    }

    out += string(buf[er][:ec])

    return out
}


func copySelection(buf [][]rune, sel *Selection) {
    if !sel.Active {
        return
    }
    text := getSelectedText(buf, sel)
    copyToClipboard(text)
}


func cutSelection(buf *[][]rune, cur *Cursor, sel *Selection) {
    if !sel.Active {
        return
    }
    text := getSelectedText(*buf, sel)
    copyToClipboard(text)
    deleteSelection(buf, cur, sel)
}

func deleteSelection(buf *[][]rune, cur *Cursor, sel *Selection) {
    sr, sc, er, ec := normalizeSelection(sel)

    // simple case: same line
    if sr == er {
        line := (*buf)[sr]
        (*buf)[sr] = append(line[:sc], line[ec:]...)
        cur.Row, cur.Col = sr, sc
        sel.Active = false
        return
    }

    // multi-line delete
    first := (*buf)[sr][:sc]
    last := (*buf)[er][ec:]

    newLine := append(first, last...)

    // replace all lines from sr..er with newLine
    *buf = append((*buf)[:sr], append([][]rune{newLine}, (*buf)[er+1:]...)...)

    cur.Row, cur.Col = sr, sc
    sel.Active = false
}

func pasteText(buf *[][]rune, cur *Cursor) {
    text := pasteFromClipboard()

    i := 0
    for i < len(text) {
        ch := text[i]

        // Handle CRLF (Windows line endings)
        if ch == '\r' {
            if i+1 < len(text) && text[i+1] == '\n' {
                // Treat CRLF as a single newline
                insertNewline(buf, cur)
                i += 2
                continue
            }
            // lone CR (rare)
            insertNewline(buf, cur)
            i++
            continue
        }

        if ch == '\n' {
            insertNewline(buf, cur)
            i++
            continue
        }

        // Normal character
        insertRune(buf, cur, rune(ch))
        i++
    }
}