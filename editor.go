package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

type Cmd rune

const (
	UP        Cmd = 1000
	DOWN          = 1001
	LEFT          = 1002
	RIGHT         = 1003
	PAGE_UP       = 1004
	PAGE_DOWN     = 1005
	HOME_KEY      = 1006
	END_KEY       = 1007
	DELETE        = 1008
)

type Editor struct {
	originalTermios *unix.Termios
	wRows, wCols    uint
	cx, cy          uint
	cRow            []byte
	numRows         uint
}

func ConstructEditor() Editor {
	e := Editor{
		cx:      0,
		cy:      0,
		numRows: 0,
	}
	e.GetWindowSize()
	return e
}

func (e *Editor) ShowCursor() {
	fmt.Printf("\x1b[%d;%dH", e.cy+1, e.cx+1)
}

func (e *Editor) HideCursor() {
	fmt.Printf("\x1b[%d;%dL", e.cy+1, e.cx+1)
}

func (e *Editor) GetWindowSize() (uint, uint) {
	if e.wRows != 0 && e.wCols != 0 {
		return e.wRows, e.wCols
	}
	return e.getWindowSize()
}

// getWindowSize returns number of rows, then columns. (0, 0) if error occurs
// TODO: Sometimes unix.IoctlGetWinsize will fail. Implement fallback
//   https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#window-size-the-hard-way
func (e *Editor) getWindowSize() (uint, uint) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdin.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0
	}
	e.wRows = uint(ws.Row)
	e.wCols = uint(ws.Col)
	return e.wRows, e.wCols
}

func (e *Editor) DrawRows() {
	r, _ := e.GetWindowSize()
	fmt.Printf("\x1b[K") // Clear line

	for i := uint(0); i < e.numRows; i++ {
		fmt.Printf(string(e.cRow))
	}
	e.DrawEmptyRows(r - e.numRows)
}

func (e *Editor) DrawEmptyRows(r uint) {
	for i := uint(1); i < r; i++ {
		fmt.Printf("~\r\n")
	}
	fmt.Printf("~\r")
}

func (e *Editor) RefreshScreen() {
	fmt.Printf("\x1b[2J") // Clear the screen
	e.HideCursor()
	fmt.Printf("\x1b[H") // Reposition Cursor

	e.DrawRows()

	e.ShowCursor()
}

func (e *Editor) ReadChar() byte {
	c := make([]byte, 1)
	cs, _ := os.Stdin.Read(c)
	if cs == 0 {
		return 0x00
	}
	return c[0]
}

func (e *Editor) KeyPress(x byte) bool {
	var c Cmd = 0x00
	switch x {
	case Ctrl('q'):
		return true
	case '\x1b':
		c = e.HandleEscapeCode()
		break

	}

	e.HandleMoveCursor(c)
	if !isControlChar(x) {
		fmt.Printf(string(x))
	}
	return false
}

func (e *Editor) DisableRawMode() {
	err := unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TIOCSETA, e.originalTermios)
	if err != nil {
		fmt.Printf(fmt.Errorf(
			"Error on terminal close when disabling raw mode. Error: %w\n", err,
		).Error(),
		)
	}
}

func (e *Editor) HandleMoveCursor(x Cmd) {
	switch x {
	case LEFT:
		if e.cx != 0 {
			e.cx--
		}
		break
	case RIGHT:
		if (e.cx + 1) < e.wCols {
			e.cx++
		}
		break
	case UP:
		if e.cy != 0 {
			e.cy--
		}
		break
	case DOWN:
		if (e.cy + 1) < e.wRows {
			e.cy++
		}
		break
	case PAGE_UP:
		for i := uint(0); i < e.wRows; i++ {
			e.HandleMoveCursor(UP)
		}
		break
	case PAGE_DOWN:
		for i := uint(0); i < e.wRows; i++ {
			e.HandleMoveCursor(DOWN)
		}
		break
	case HOME_KEY:
		e.cx = 0
		break
	case END_KEY:
		e.cx = e.wCols - 1
		break
	}
}

func (e *Editor) HandleEscapeCode() Cmd {
	a := e.ReadChar()
	if a == '\x1b' {
		return '\x1b'
	}
	b := e.ReadChar()
	if b == '\x1b' {
		return '\x1b'
	}

	if a == '[' {
		// Page up/down
		if b >= '0' && b <= '9' {
			c := e.ReadChar()
			if c == '\x1b' {
				return '\x1b'
			}
			if c == '~' {
				switch c {
				case '1':
					return HOME_KEY
				case '3':
					return DELETE
				case '4':
					return END_KEY
				case '5':
					return PAGE_UP
				case '6':
					return PAGE_DOWN
				case '7':
					return HOME_KEY
				case '8':
					return END_KEY
				}
			}
		}

		// Arrow keys
		switch b {
		case 'A':
			return UP
		case 'B':
			return DOWN
		case 'C':
			return RIGHT
		case 'D':
			return LEFT
		}
	}
	if a == 'O' {
		switch b {
		case 'H':
			return HOME_KEY
		case 'F':
			return END_KEY
		}
	}
	return '\x1b'
}

func (e *Editor) Open() {
	e.cRow = []byte("Hello World\000")
	e.numRows = 1
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