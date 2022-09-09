package main

import (
	"fmt"
	"os"
)

func main() {
	e := ConstructEditor()
	err := e.Open("editor.go")

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
	saveErr := e.Save("editor_1.out")
	e.DisableRawMode()
	fmt.Println("\x1b[2J")
	fmt.Println("\x1b[H")
	if err != nil {
		fmt.Println(err)
	}
	if saveErr != nil {
		fmt.Println(saveErr)
	}
	if saveErr != nil || err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
