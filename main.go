package main

import (
	"fmt"
	"os"
	"tengo/editor"
)

///Main Run Here
func main() {
	//var filePath = "sample.json"
	
	
    if len(os.Args) < 2 {
        fmt.Println("Usage: editor <filename>")
        return
    }

    filePath := os.Args[1]
    restore := editor.EnableRaw()
    defer restore()
	
	fmt.Print("\x1b[?2004h") // enable bracketed paste
	defer fmt.Print("\x1b[?2004l") // disable on exit

		
	fmt.Print("\x1b[?1049h")
	defer fmt.Print("\x1b[?1049l")

    
    editor.RunQuickEditor(filePath)

}
