package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

type Cmd rune

const STATUS_BAR = 1

const (
	// Must be higher than 128 to avoid clashing with ASCII
	UP          Cmd = 1000
	DOWN            = 1001
	LEFT            = 1002
	RIGHT           = 1003
	PAGE_UP         = 1004
	PAGE_DOWN       = 1005
	HOME_KEY        = 1006
	END_KEY         = 1007
	DELETE          = 1008
	SHIFT_RIGHT     = 1009
	SHIFT_LEFT      = 1010

	// Specific ANSI mappings
	BACKSPACE = 127
	ENTER     = 13
	SEARCH    = 6  // Ctrl-F on Mac OS
	UNDO      = 26 // Ctrl-X on Mac OS
)

type Editor struct {
	originalTermios      *unix.Termios
	wRows, wCols         uint  // size of Editor
	cx, cy               uint  // Position in file of cursor
	rows                 []Row // Rows in file
	rowOffset, colOffset uint  // Position in file of top left corner of editor
	filename             string
	charHistory          ByteRing
	cmdHistory           *CommandHistory
	syntax               *Syntax
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
		charHistory: CreateByteRing(10),
		cmdHistory:  CreateCommandHistory(),
		syntax:      CreateSyntax(filename),
	}
	e.GetWindowSize()
	return e, nil
}

func (e *Editor) ShowCursor() {
	e.MoveCursor((e.cx-e.colOffset)+1, (e.cy-e.rowOffset)+1)
}

func (e *Editor) HideCursor() {
	fmt.Printf("\x1b[%d;%dL", (e.cy-e.rowOffset)+1, (e.cx-e.colOffset)+1)
}

func (e *Editor) MoveCursor(x, y uint) {
	fmt.Printf("\x1b[%d;%dH", y, x)
}

func (e *Editor) MoveCursorToStatusBar() {
	e.MoveCursor(0, e.wRows-1)
}

func (e *Editor) ClearLine() {
	fmt.Printf("\x1b[K")
}

func (e *Editor) GetWindowSize() (uint, uint) {
	if e.wRows != 0 && e.wCols != 0 {
		return e.wRows, e.wCols
	}
	e.wRows, e.wCols = GetWindowSize()
	return e.wRows, e.wCols
}

func (e *Editor) GetCurrentRow() *Row {
	return &e.rows[e.cy]
}

func (e *Editor) GetRowLength() uint {
	return e.GetCurrentRow().RenderLen()
}

func (e *Editor) GetEditorRows() uint {
	return uint(e.wRows) - STATUS_BAR
}

func (e *Editor) GetDocumentRows() uint {
	return uint(len(e.rows))
}

func (e *Editor) DrawRows() {
	e.ClearLine()

	r := e.GetEditorRows()

	// Leave room for status bar
	nRows := e.GetDocumentRows()
	if nRows > e.wRows {
		nRows = e.GetEditorRows()
	}

	for _, r := range e.rows[e.rowOffset : e.rowOffset+nRows] {
		l := r.RenderWithin(e.colOffset, e.wCols-e.colOffset)
		fmt.Printf("%s\r\n", e.syntax.Highlight(l))
	}
	e.DrawEmptyRows(r - nRows)
	e.DrawStatusBar()
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
		e.HandleOtherEscapedCmds(c)
		break
	case BACKSPACE:
		if e.GetRowLength() == 0 {
			e.RemoveCurrentRow()
		} else {
			e.GetCurrentRow().RemoveCharAt(e.cx)
			e.HandleMoveCursor(LEFT)
		}

	case ENTER:
		e.SplitCurrentRow()

		y := e.cy // So that AddCmd had value of current y, not future e.cy
		e.cmdHistory.AddCmd(
			func(e *Editor) error { e.JoinRows(y, y+1); return nil },
			func(e *Editor) error { return fmt.Errorf("REDO unimplemented") },
		)
		e.HandleMoveCursor(RIGHT) // Jumps to start of next (newly-created) line.
	case SEARCH:
		e.cx, e.cy = e.RunSearch()

	case UNDO:
		e.cmdHistory.Undo(e)
	}

	if !isControlChar(x) {
		e.GetCurrentRow().AddCharAt(e.cx, x)
		e.HandleMoveCursor(RIGHT)
	}
	return false
}

func (e *Editor) RemoveCurrentRow() {
	if (e.cy+1) == e.GetDocumentRows() && e.GetDocumentRows() > 0 {
		// Last row, just remove
		e.rows = e.rows[:]
	} else if e.GetDocumentRows() > 0 {
		e.rows = append(e.rows[:e.cy], e.rows[e.cy+1:]...)
	}
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
			if (e.cy + 1) < e.GetDocumentRows() {
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
		if (e.cy + 1) < e.GetDocumentRows() {
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
	case SHIFT_RIGHT:
		e.cx = e.GetCurrentRow().GetNextWordFrom(e.cx, true)
	case SHIFT_LEFT:
		e.cx = e.GetCurrentRow().GetNextWordFrom(e.cx, false)
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
			if c == 0x3B {
				d := e.ReadChar()
				if d == '\x1b' {
					return '\x1b'
				}
				e := e.ReadChar()
				if e == '\x1b' {
					return '\x1b'
				}
				if d == 0x32 {
					switch e {
					case 0x43:
						return SHIFT_RIGHT
					case 0x44:
						return SHIFT_LEFT
					}
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
	h := e.cmdHistory.Depth()
	Cprintf("%DarkBlue%STATUS BAR --% (%Blue.d, %Blue.d) of (%Magenta.d, %Magenta.d) %v. Row: %d. History: %d", e.cx, e.cy, x, y, e.charHistory.GetHistory(), r.RenderLen(), h)
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

// HandleOtherEscapedCmds is responsible for handling other ANSI escape keys that don't simply move the cursor
// (i.e. they can move the cursor, but only through e.HandleMoveCursor()).
func (e *Editor) HandleOtherEscapedCmds(c Cmd) {
	switch c {
	case DELETE:
		e.GetCurrentRow().RemoveCharAt(e.cx + 1)
	}
}

// SplitCurrentRow based on the current cursor position.
func (e *Editor) SplitCurrentRow() {
	a, b := e.GetCurrentRow().SplitAt(e.cx)
	newRows := []Row{*a, *b}

	if (e.cy + 1) >= e.GetDocumentRows() {

		// Last line in file, must make a new row
		e.rows = append(e.rows[:e.GetDocumentRows()-1], newRows...)
	} else {
		end := make([]Row, e.GetDocumentRows()-e.cy-1)
		copy(end, e.rows[e.cy+1:])
		e.rows = append(e.rows[:e.cy], newRows...)
		e.rows = append(e.rows, end...)
	}

}

// Run Search across file. Return co-ordinate, (x, y) of first result.
// Will read inputs until an enter is pressed, and will be used as search term.
func (e *Editor) RunSearch() (uint, uint) {
	q := make([]byte, 0)
	b := e.ReadChar()

	for b != ENTER {
		e.MoveCursorToStatusBar()
		e.ClearLine()
		fmt.Printf("SEARCH: %s", string(q))

		if !isControlChar(b) {
			q = append(q, b)
		} else if b == BACKSPACE && len(q) > 0 {
			q = q[:len(q)-1]
		}
		b = e.ReadChar()

	}

	// Read search results and let user go through results.
	for r := range SearchRows(e.rows, string(q)) {
		// TODO: handle move cursor on page scrolling.
		e.MoveCursor(r.startI+1, r.rowI+1)

		// Blocking read on input.
		b = e.ReadChar()
		for b == 0x00 {
			b = e.ReadChar()
		}

		if b == SEARCH {
			return r.startI, r.rowI // Exit search mode
		} else if b == ENTER {
			continue // Go to next search term
		} else {
			break // Leave search, back to original cursor.
		}
	}

	// Default back to current position
	e.MoveCursor(e.cx, e.cy)
	return e.cx, e.cy
}

// JoinRows from b into row a. If a or b index out of range, no action applied.
func (e *Editor) JoinRows(a, b uint) {
	if a >= e.GetDocumentRows() || b >= e.GetDocumentRows() {
		return
	}
	e.rows[a].Append(&e.rows[b])

	// Remove row b
	if (b + 1) < e.GetDocumentRows() {
		e.rows = append(e.rows[:b], e.rows[b+1:]...)
	} else {
		e.rows = e.rows[:b]
	}
}
