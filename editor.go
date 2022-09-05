package main

import (
	"fmt"
	"golang.org/x/sys/unix"
)

type Editor struct {
	originalTermios unix.Termios
}

func (e Editor) DrawRows() {
	for i := 0; i < 24; i++ {
		fmt.Println("~\r")
	}
}

func (e Editor) RefreshScreen() {
	fmt.Printf("\x1b[2J") // Clear the screen
	fmt.Printf("\x1b[H")  // Reposition Cursor

	e.DrawRows()
	fmt.Printf("\x1b[H")
}

func (e Editor) KeyPress(x byte) bool {
	switch x {
	case Ctrl('q'):
		return true
	}

	if !isControlChar(x) {
		fmt.Printf(string(x))
	}
	return false
}
