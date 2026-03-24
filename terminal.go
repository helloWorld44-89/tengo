package main

import (
		 "os"
		 "golang.org/x/term"
		 "syscall")

func enableRaw() func() { // <--This is to put terminal in Raw mode.
	oldState, _ :=term.MakeRaw(int(syscall.Stdin))
	return func() {
		term.Restore(int(syscall.Stdin), oldState)
	}
}

func getTerminalSize() (width, height int) {
    width, height, _ = term.GetSize(int(os.Stdout.Fd()))
    return
}

