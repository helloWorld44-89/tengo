package main

import (
		 "os"
		 "fmt"
		 "golang.org/x/term"
		 "syscall")

func enableRaw() func() {
    oldState, _ := term.MakeRaw(int(syscall.Stdin))

    // DISABLE BRACKETED PASTE MODE
    fmt.Print("\x1b[?2004l")

    return func() {
        // Restore bracketed paste
        fmt.Print("\x1b[?2004h")
        term.Restore(int(syscall.Stdin), oldState)
    }
}

func getTerminalSize() (width, height int) {
    width, height, _ = term.GetSize(int(os.Stdout.Fd()))
    return
}

