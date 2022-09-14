package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "No file was specified. Usage: `gram <FILENAME>`")
		return
	}
	filename := os.Args[1]
	e, err := ConstructEditor(filename)
	if err != nil {
		Exit(e, err)
	}

	// TODO: refactor EnableRawMode() into Editor struct function
	t, err := EnableRawMode()
	e.originalTermios = &t

	if err != nil {
		Exit(e, err)
	}

	for !e.KeyPress() {
		e.RefreshScreen()
	}
	Exit(e, nil)
}

// Exit program. Required after an Editor has been constructed (i.e. ConstructEditor)
func Exit(e Editor, err error) {
	closeErr := e.Close()
	fmt.Println("\x1b[2J")
	fmt.Println("\x1b[H")
	if err != nil {
		fmt.Println(err)
	}
	if closeErr != nil {
		fmt.Println(closeErr)
	}
	if closeErr != nil || err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
