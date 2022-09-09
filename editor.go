package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"strings"
)

type Cmd rune

const STATUS_BAR = 1

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
	originalTermios      *unix.Termios
	wRows, wCols         uint  // size of Editor
	cx, cy               uint  // Position in file of cursor
	rows                 []Row // Rows in file
	rowOffset, colOffset uint  // Position in file of top left corner of editor

	previousChar [5]byte
	pCharI       int

	filename string
}

func ConstructEditor() Editor {
	e := Editor{
		cx:           0,
		cy:           0,
		rowOffset:    0,
		colOffset:    0,
		wRows:        0,
		wCols:        0,
		filename:     "",
		rows:         []Row{},
		previousChar: [5]byte{0x00, 0x00, 0x00, 0x00, 0x00},
		pCharI:       0,
	}
	e.GetWindowSize()
	return e
}

func (e *Editor) Open(filename string) error {
	e.filename = filename
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	file := strings.ReplaceAll(string(raw), "\r", "\n")
	rows := strings.Split(file, "\n")
	e.rows = make([]Row, len(rows))

	for i, s := range rows {
		e.rows[i] = ConstructRow(s)
	}
	return nil
}

func (e *Editor) ShowCursor() {
	fmt.Printf("\x1b[%d;%dH", (e.cy-e.rowOffset)+1, (e.cx-e.colOffset)+1)
}

func (e *Editor) HideCursor() {
	fmt.Printf("\x1b[%d;%dL", (e.cy-e.rowOffset)+1, (e.cx-e.colOffset)+1)
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
	fmt.Printf("\x1b[K") // Clear line

	r := e.GetEditorRows()

	// Leave room for status bar
	nRows := e.GetDocumentRows()
	if nRows > e.wRows {
		nRows = e.GetEditorRows()
	}

	for i := uint(0); i < nRows; i++ {
		fmt.Printf("%s\r\n", e.DrawRow(e.rows[i+e.rowOffset]))
	}
	e.DrawEmptyRows(r - nRows)
	e.DrawStatusBar()
}
func (e *Editor) DrawRow(r Row) string {
	l := r.Render()
	lLen := uint(len(l)) - e.colOffset
	if e.colOffset > uint(len(l)) {
		return ""
	} else if lLen > e.wCols {
		return l[e.colOffset : e.colOffset+e.wCols]
	} else {
		return l[e.colOffset:]
	}
}

func (e *Editor) DrawEmptyRows(r uint) {
	for i := uint(1); i < r; i++ {
		fmt.Printf("~\r\n")
	}
	fmt.Printf("~\r")
}

func (e *Editor) RefreshScreen() {
	e.SetScroll()
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

	e.previousChar[e.pCharI] = c[0]
	e.pCharI = (e.pCharI + 1) % 5

	return c[0]
}

func (e *Editor) KeyPress() bool {
	x := e.ReadChar()
	var c Cmd = 0x00
	switch x {
	case Ctrl('q'):
		return true
	case '\x1b':
		c = e.HandleEscapeCode()
		e.HandleMoveCursor(c)
		break

	}

	if !isControlChar(x) {
		e.GetCurrentRow().AddCharAt(e.cx, x)
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

func (e *Editor) GetCurrentRow() *Row {
	return &e.rows[e.cy]
}

func (e *Editor) GetRowLength() uint {
	return e.GetCurrentRow().RenderLen()
}

func (e *Editor) GetHorizontalRightShift() uint {
	if e.GetCurrentRow().GetCharAt(e.cx) == '\t' {
		return 4
	}
	return 1
}

func (e *Editor) GetHorizontalLeftShift() uint {
	if e.GetCurrentRow().GetCharAt(e.cx) == '\t' {
		return 4
	}
	return 1
}

func (e *Editor) HandleMoveCursor(x Cmd) {
	switch x {
	case LEFT:
		if e.cx != 0 {
			e.cx -= 1 // e.GetHorizontalLeftShift()

			// Move left at start of line, go to end of previous line
		} else if e.cy != 0 {
			e.cy--
			e.cx = e.GetRowLength()
		}
		break
	case RIGHT:
		// Move right at EOL, go to start of next line.
		if e.cx+1 >= e.GetRowLength() {
			if e.cy < e.GetDocumentRows() {
				e.cy++
				e.cx = 0
			}
		} else {
			e.cx += 1 // e.GetHorizontalRightShift()
		}
		break
	case UP:
		if e.cy > 0 {
			e.cy--
		}
		break
	case DOWN:
		if (e.cy + 1) < uint(len(e.rows)) {
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
		e.cx = e.GetRowLength() - 1
		if e.cx > e.wCols {
			e.colOffset = e.cx - e.wCols
		}
		break
	}

	rowL := e.GetRowLength()
	if rowL == 0 {
		e.cx = 0
	} else if e.cx >= rowL {
		e.cx = rowL - 1
	}
}

func (e *Editor) GetEditorRows() uint {
	return uint(e.wRows) - STATUS_BAR
}

func (e *Editor) GetDocumentRows() uint {
	return uint(len(e.rows))
}

func (e *Editor) SetScroll() {
	if e.cy < e.rowOffset {
		e.rowOffset = e.cy
	} else if e.cy >= (e.rowOffset + e.GetEditorRows()) {
		e.rowOffset = e.cy - e.GetEditorRows() + 1
	}

	if e.cx < e.colOffset {
		e.colOffset = e.cx
	} else if e.cx >= (e.colOffset + e.wCols) {
		e.colOffset = e.cx - e.wCols + 1
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

			if c == 0x7E {
				switch b {
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
		case 'F':
			return END_KEY
		case 'H':
			return HOME_KEY
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

func (e *Editor) GetCharHistory() [5]byte {
	r := [5]byte{0x00, 0x00, 0x00, 0x00, 0x00}

	for i := 0; i < 5; i++ {
		r[i] = e.previousChar[(e.pCharI+(5-i-1))%5]
	}
	return r
}

func (e *Editor) DrawStatusBar() {
	y, x := e.GetWindowSize()
	fmt.Printf("STATUS BAR -- (%d, %d) of (%d, %d) %v ", e.cx, e.cy, x, y, e.GetCharHistory())
}

func (e *Editor) Close() error {
	saveErr := e.Save()
	e.DisableRawMode()
	e.filename = ""
	return saveErr
}

func (e *Editor) Save() error {
	f, err := os.Create(e.filename)
	if err != nil {
		return err
	}

	for _, row := range e.rows {
		_, err := f.Write(row.Export())
		if err != nil {
			return err
		}
	}
	return nil
}

func Ctrl(b byte) byte {
	return b & 0x1f
}

func isControlChar(x byte) bool {
	return x <= 31 || x == 127
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
