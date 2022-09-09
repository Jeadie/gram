package main

import (
	"fmt"
	"os"
)

func main() {
	filename := os.Args[1]
	e := ConstructEditor()
	err := e.Open(filename)

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
