package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"strings"
)

type Editor struct {
	originalTermios *unix.Termios
	wRows, wCols    uint
	cx, cy          uint
}

func ConstructEditor() Editor {
	e := Editor{
		cx: 0,
		cy: 0,
	}
	e.GetWindowSize()
	return e
}

func (e Editor) ShowCursor() {
	fmt.Printf("\x1b[%d;%dH", e.cy+1, e.cx+1)
}

func (e Editor) HideCursor() {
	fmt.Printf("\x1b[%d;%dL", e.cy+1, e.cx+1)
}

func (e Editor) GetWindowSize() (uint, uint) {
	if e.wRows != 0 && e.wCols != 0 {
		return e.wRows, e.wCols
	}
	return e.getWindowSize()
}

// getWindowSize returns number of rows, then columns. (0, 0) if error occurs
// TODO: Sometimes unix.IoctlGetWinsize will fail. Implement fallback
//   https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#window-size-the-hard-way
func (e Editor) getWindowSize() (uint, uint) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdin.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0
	}
	e.wRows = uint(ws.Row)
	e.wCols = uint(ws.Col)
	return e.wRows, e.wCols
}

func (e Editor) DrawRows() {
	r, c := e.GetWindowSize()
	fmt.Printf("\x1b[K") // Clear line

	for i := uint(1); i < r; i++ {
		if i == (r / 3) {
			welcomeMsg := "Gram editor -- version 0.0.1"
			pad := strings.Repeat(" ", (int(c)-len(welcomeMsg))/2)
			fmt.Printf("%sGram editor -- version 0.0.1%s", pad, pad)
		} else {
			fmt.Printf("~\r\n")
		}
	}
	fmt.Printf("~\r")
}

func (e Editor) RefreshScreen() {
	fmt.Printf("\x1b[2J") // Clear the screen
	fmt.Printf("\x1b[H")  // Reposition Cursor
	e.HideCursor()

	e.DrawRows()

	e.ShowCursor()
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

func (e Editor) DisableRawMode() {
	err := unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TIOCSETA, e.originalTermios)
	if err != nil {
		fmt.Printf(fmt.Errorf(
			"Error on terminal close when disabling raw mode. Error: %w\n", err,
		).Error(),
		)
	}
}

func EnableRawMode() (unix.Termios, error) {
	// TODO: these are only for Mac OS, and not other linux
	const ioctlReadTermios = unix.TIOCGETA
	const ioctlWriteTermios = unix.TIOCSETA

	fd := int(os.Stdin.Fd())

	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	returnTermios := *termios

	if err != nil {
		return returnTermios, err
	}

	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.IXON
	termios.Iflag &^= unix.IXON  // Ctrl-S and Ctrl-Q
	termios.Iflag &^= unix.ICRNL // Ctrl-M

	termios.Oflag &^= unix.OPOST // #Output Processing

	termios.Lflag &^= unix.ECHO | unix.ECHONL // Echo
	termios.Lflag &^= unix.ICANON             // Canonical Mode
	termios.Lflag &^= unix.ISIG               // Ctrl-C and Ctrl-Z
	termios.Lflag &^= unix.IEXTEN             // ctrl-V, ctrl-O (on macOS

	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8

	termios.Cc[unix.VMIN] = 0  // Number of bytes to let Read() return
	termios.Cc[unix.VTIME] = 1 // Maximum wait time for Read(). Measured in tenths of a second

	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, termios); err != nil {
		return returnTermios, err
	}

	return returnTermios, err
}
