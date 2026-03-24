package main

import (
    "bytes"
    "os/exec"
    "runtime"
)

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

    // single-line selection
    if sr == er {
        return string(buf[sr][sc:ec])
    }

    // multi-line selection
    var out string

    // first line
    out += string(buf[sr][sc:]) + "\n"

    // middle lines
    for row := sr + 1; row < er; row++ {
        out += string(buf[row]) + "\n"
    }

    // last line
    out += string(buf[er][:ec])

    return out
}