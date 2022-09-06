package main

import (
	"fmt"
	"os"
)

func main() {
	// TODO: refactor EnableRawMode() into Editor struct function
	e := ConstructEditor()
	err := e.Open("editor.go")
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
	e.DisableRawMode()
	fmt.Println("\x1b[2J")
	fmt.Println("\x1b[H")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
