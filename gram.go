package main

import (
	"fmt"
	"os"
)

func main() {
	// TODO: refactor EnableRawMode() into Editor struct function
	e := ConstructEditor()
	e.Open()
	t, err := EnableRawMode()
	e.originalTermios = &t

	if err != nil {
		Exit(e, err)
	}

	for true {
		cc := e.ReadChar()
		e.RefreshScreen()
		shouldExit := e.KeyPress(cc)
		if shouldExit {
			Exit(e, nil)
		}
	}
}

func Exit(e Editor, err error) {
	e.DisableRawMode()
	fmt.Println("\x1b[2J")
	fmt.Println("\x1b[H")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func Ctrl(b byte) byte {
	return b & 0x1f
}

func isControlChar(x byte) bool {
	return x <= 31 || x == 127
}
