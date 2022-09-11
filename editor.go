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
	BACKSPACE     = 127
)

type Editor struct {
	originalTermios      *unix.Termios
	wRows, wCols         uint  // size of Editor
	cx, cy               uint  // Position in file of cursor
	rows                 []Row // Rows in file
	rowOffset, colOffset uint  // Position in file of top left corner of editor

	charHistory ByteRing

	filename string
}

func ConstructEditor(filename string) (Editor, error) {

	rows, err := OpenOrCreate(filename)
	if err != nil {
		return Editor{}, err
	}
	e := Editor{
		cx:          0,
		cy:          0,
		rowOffset:   0,
		colOffset:   0,
		wRows:       0,
		wCols:       0,
		filename:    filename,
		rows:        rows,
		charHistory: CreateByteRing(5),
	}
	e.GetWindowSize()
	return e, nil
}

func (e *Editor) Touch(filename string) error {
	return ioutil.WriteFile(filename, []byte{}, 0555)
}

func (e *Editor) Open(filename string) error {
	e.filename = filename
	if !fileExists(filename) {
		err := e.Touch(filename)
		if err != nil {
			return err
		}
	}

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
	e.wRows, e.wCols = GetWindowSize()
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

	e.charHistory.Insert(c[0])
	return c[0]
}

func (e *Editor) KeyPress() bool {
	x := e.ReadChar()
	switch x {
	case Ctrl('q'):
		return true
	case '\x1b':
		c := e.HandleEscapeCode()
		e.HandleMoveCursor(c)
		break
	case BACKSPACE:
		e.GetCurrentRow().RemoveCharAt(e.cx)
		e.HandleMoveCursor(LEFT)
	}

	if !isControlChar(x) {
		e.GetCurrentRow().AddCharAt(e.cx, x)
		e.HandleMoveCursor(RIGHT)
	}
	return false
}

func (e *Editor) GetCurrentRow() *Row {
	return &e.rows[e.cy]
}

func (e *Editor) GetRowLength() uint {
	return e.GetCurrentRow().RenderLen()
}

func (e *Editor) HandleMoveCursor(x Cmd) {
	switch x {
	case LEFT:
		if e.cx != 0 {
			e.cx -= 1

			// Move left at start of line, go to end of previous line
		} else if e.cy != 0 {
			e.cy--
			e.cx = e.GetRowLength()
		}
		break
	case RIGHT:
		// Move right at EOL, go to start of next line.
		if (e.cx + 1) > e.GetRowLength() {
			if e.cy < e.GetDocumentRows() {
				e.cy++
				e.cx = 0
			}
		} else {
			e.cx += 1
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
		e.cx = e.GetRowLength()
		if e.cx > e.wCols {
			e.colOffset = e.cx - e.wCols
		}
		break
	}

	if e.cy >= uint(len(e.rows)) {
		e.cy = uint(len(e.rows)) - 1
	}

	rowL := e.GetRowLength()
	if rowL == 0 {
		e.cx = 0
	} else if e.cx > rowL {
		e.cx = rowL
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

func (e *Editor) DrawStatusBar() {
	y, x := e.GetWindowSize()
	r := e.GetCurrentRow()
	fmt.Printf("STATUS BAR -- (%d, %d) of (%d, %d) %v. Row: %d", e.cx, e.cy, x, y, e.charHistory.GetHistory(), r.RenderLen())
}

func (e *Editor) Close() error {
	saveErr := e.Save()
	err := RevertTerminalMode(e.originalTermios)
	if err != nil {
		fmt.Println(fmt.Errorf(
			"Error on terminal close when disabling raw mode. Error: %w\n", err,
		).Error())
	}
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
