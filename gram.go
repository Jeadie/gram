package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

func main() {
	c := make([]byte, 1)
	cs := 1 // Number of characters read. Must be > 0 for initial for-loop check
	var err error

	t, err := EnableRawMode()
	defer DisableRawMode(t)
	if err != nil {
		os.Exit(1)
	}

	for true {
		cs, _ = os.Stdin.Read(c)
		if cs == 0 {
			continue
		}

		cc := c[0]
		EditorRefreshScreen()
		shouldExit := EditorKeyPress(cc)
		if shouldExit {
			return
		}
	}
}

func EditorDrawRows() {
	for i := 0; i < 24; i++ {
		fmt.Println("~\r")
	}
}

func EditorRefreshScreen() {
	fmt.Printf("\x1b[2J") // Clear the screen
	fmt.Printf("\x1b[H")  // Reposition Cursor

	EditorDrawRows()
	fmt.Printf("\x1b[H")
}

func EditorKeyPress(x byte) bool {
	switch x {
	case Ctrl('q'):
		return true
	}

	if !isControlChar(x) {
		fmt.Printf(string(x))
	}
	return false
}

func Ctrl(b byte) byte {
	return b & 0x1f
}

func DisableRawMode(t unix.Termios) {
	err := unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TIOCSETA, &t)
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
	returnT := *termios
	if err != nil {
		return returnT, err
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
		return returnT, err
	}

	return returnT, err
}

func isControlChar(x byte) bool {
	return x <= 31 || x == 127
}
