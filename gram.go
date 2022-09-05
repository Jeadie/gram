package main

import (
	"fmt"
	"os"
)

func main() {
	c := make([]byte, 1)
	cs := 1 // Number of characters read. Must be > 0 for initial for-loop check

	// TODO: refactor EnableRawMode() into Editor struct function
	e := ConstructEditor()
	t, err := EnableRawMode()
	e.originalTermios = &t

	if err != nil {
		Exit(e, err)
	}

	for true {
		cs, _ = os.Stdin.Read(c)
		if cs == 0 {
			continue
		}

		cc := c[0]
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
