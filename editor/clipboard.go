package editor

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/tiagomelo/go-clipboard/clipboard"
)

var internalClipboard string
var isPasting bool
var inBracketedPaste bool
var pasteBuffer strings.Builder

// --- Copy string to system clipboard if possible ---
func copyToClipboard(text string) {
    c := clipboard.New()
    err := c.CopyText(text)
    if err != nil {
        panic(err)
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
        out, err := exec.Command(
            "powershell.exe",
            "-NoProfile",
            "-STA",
            "-Command",
            "Get-Clipboard").Output()

        if err == nil {
            return string(out)
        }

    }

    return internalClipboard
}


func getSelectedText(buf [][]rune, sel *Selection) string {
    sr, sc, er, ec := normalizeSelection(sel)

    // Single-line selection
    if sr == er {
        return string(buf[sr][sc:ec])
    }

    var out strings.Builder

    // First line: from sc → end
    out.WriteString(string(buf[sr][sc:]))
    out.WriteByte('\n')

    // Middle lines: full lines
    for i := sr + 1; i < er; i++ {
        out.WriteString(string(buf[i]))
        out.WriteByte('\n')
    }

    // Last line: from start → ec
    out.WriteString(string(buf[er][:ec]))

    return out.String()
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
    if !sel.Active {
        return
    }
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
 
    newBuf := make([][]rune, 0, len(*buf)-(er-sr))
    newBuf = append(newBuf, (*buf)[:sr]...)
    newBuf = append(newBuf, newLine)
    newBuf = append(newBuf, (*buf)[er+1:]...)
    *buf = newBuf


    cur.Row, cur.Col = sr, sc
    sel.Active = false
}

func pasteText(buf *[][]rune, cur *Cursor) {
    text := pasteFromClipboard()    
    // text = strings.ReplaceAll(text, "\r\n", "\n")
    // text = strings.ReplaceAll(text, "\r", "\n")
    isPasting = true
    text = strings.TrimRight(text, "\n\r")
    for _, ch := range text {
        switch ch {
        case '\r':
            continue
        case '\n':
            insertNewline(buf, cur)
        default:
            insertRune(buf, cur, ch)
        }
    }
    isPasting = false
}