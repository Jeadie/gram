package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
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
	BACKSPACE     = 127
	ENTER         = 13
	SEARCH        = 6  // Ctrl-F on Mac OS
	UNDO          = 26 // Ctrl-X on Mac OS
	COPY          = 3  // Ctrl-C on Mac OS
	PASTE         = 22 // Ctrl-V on Mac OS
	DELETE_ROW    = 4  // Ctrl-D on Mac OS
	SAVE          = 19 // Ctrl-S on Mac OS
	SAVE_AND_EXIT = 17 // Ctrl-Q on Mac OS
	EXIT          = 23 // Ctrl-W on Mac OS
)

type Editor struct {
	originalTermios      *unix.Termios
	wRows, wCols         uint  // size of Editor
	cx, cy               uint  // Position in file of cursor
	rows                 []Row // Rows in file
	rowOffset, colOffset uint  // Position in file of top left corner of editor
	filename             string
	charHistory          byteRing
	cmdHistory           *CommandHistory
	syntax               *Syntax
	paste                string
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
		charHistory: *NewbyteRing(10),
		cmdHistory:  CreateCommandHistory(),
		syntax:      CreateSyntax(filename),
		paste:       "",
	}
	e.GetWindowSize()
	return e, nil
}

func (e *Editor) ShowCursor() {
	e.MoveCursor(e.cx, e.cy)
}

func (e *Editor) HideCursor() {
	fmt.Printf("\x1b[%d;%dL", (e.cy-e.rowOffset)+1, (e.cx-e.colOffset)+1)
}

// MoveCursor to document coordinates (x, y).
func (e *Editor) MoveCursor(x, y uint) {
	e.SetScroll()
	fmt.Printf("\x1b[%d;%dH", (y-e.rowOffset)+1, (x-e.colOffset)+1)
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

func (e *Editor) GetRow(y uint) *Row {
	return &e.rows[y]
}

func (e *Editor) GetCurrentRow() *Row {
	return e.GetRow(e.cy)
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

	// TODO: Temporary fix to address unaddressed, overflow error.
	e.DrawEmptyRows(Min(r-nRows, r-1))
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

func (e *Editor) ReadCharBlock() byte {
	c := make([]byte, 1)
	cs, _ := os.Stdin.Read(c)
	for cs == 0 && c[0] == 0x00 {
		cs, _ = os.Stdin.Read(c)
	}
	e.charHistory.Insert(c[0])
	return c[0]
}

func (e *Editor) KeyPress() bool {
	x := e.ReadChar()
	switch x {

	case SAVE_AND_EXIT:
		e.Save()
		return true

	case EXIT:
		return true

	case '\x1b':
		c := e.HandleEscapeCode()
		e.HandleMoveCursor(c)
		e.HandleOtherEscapedCmds(c)
		break
	case BACKSPACE:
		if e.GetRowLength() == 0 {
			e.RemoveCurrentRow()
		} else if e.cx == 0 && e.cy > 0 {
			l := e.GetCurrentRow().RenderLen()
			e.JoinRows(e.cy-1, e.cy)

			// Point cursor where it was prior.
			e.cy--
			e.cx = e.GetCurrentRow().RenderLen() - l

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

	case COPY:
		e.paste = e.RunCopy()
	case PASTE:
		if len(e.paste) > 0 {
			e.GetCurrentRow().AddCharsAt(e.cx, e.paste)
		}
	case DELETE_ROW:
		e.RemoveCurrentRow()

	case SAVE:
		e.Save()
	}

	if !isControlChar(x) {
		e.GetCurrentRow().AddCharAt(e.cx, x)
		e.HandleMoveCursor(RIGHT)
	}
	return false
}

// TODO: This does more than RemoveCurrentRow
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
	Cprintf("%DarkBlue%STATUS BAR --% (%Blue.d, %Blue.d) of (%Magenta.d, %Magenta.d) %v. Row: %d. History: %d. Copy: %s", e.cx, e.cy, x, y, e.charHistory.GetHistory(), r.RenderLen(), h, e.paste)
}

func (e *Editor) Close() error {
	err := RevertTerminalMode(e.originalTermios)
	if err != nil {
		fmt.Println(fmt.Errorf(
			"Error on terminal close when disabling raw mode. Error: %w\n", err,
		).Error())
	}
	e.filename = ""
	return err
}

func (e *Editor) Save() error {
	f, err := os.Create(e.filename)
	if err != nil {
		return err
	}

	noOfRows := len(e.rows)
	for i, row := range e.rows {
		_, err := f.Write(row.Export())
		if err != nil {
			return err
		}
		if i+1 < noOfRows {
			f.Write([]byte{'\n'})
		}
	}
	return nil
}

// HandleOtherEscapedCmds is responsible for handling other ANSI escape keys that don't simply move the cursor
// (i.e. they can move the cursor, but only through e.HandleMoveCursor()).
func (e *Editor) HandleOtherEscapedCmds(c Cmd) {
	switch c {
	case DELETE:
		row := e.GetCurrentRow()
		rowL := uint(len(row.Render()))

		safeToJoinRow := e.cy+1 < e.GetDocumentRows()

		if e.cx+1 < rowL {
			row.RemoveCharAt(e.cx + 1)
		} else if rowL == 1 {
			row.RemoveCharAt(0)
		} else if (e.cx == rowL || rowL == 0) && safeToJoinRow {
			// TODO: need to generalise this for BACKSPACE
			// TODO: Also is this why
			e.JoinRows(e.cy, e.cy+1)
		}
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
		e.cx = r.startI
		e.cy = r.rowI
		e.RefreshScreen()
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

func (e *Editor) RunCopy() string {
	sx, sy := e.cx, e.cy
	ex, ey := e.GetCopyEndCoordinates()
	return e.GetStringBetween(sx, sy, ex, ey)
}

func (e *Editor) GetCopyEndCoordinates() (uint, uint) {
	x := e.ReadCharBlock()
	cmd := e.HandleEscapeCode()

	// TODO: Use copy-specific equaivalent of HandleEscapeCode to avoid side effects/ changes. And below
	if x != '\x1b' || cmd == DELETE { // DELETE is only none-move Cmd from HandleEscapeCode
		return e.cx, e.cy
	}
	for true {
		e.HandleMoveCursor(cmd)
		e.ShowCursor()

		x = e.ReadCharBlock()
		cmd = e.HandleEscapeCode()

		if x != '\x1b' || cmd == DELETE {
			break
		}
	}

	return e.cx, e.cy
}

func (e *Editor) GetStringBetween(sx uint, sy uint, ex uint, ey uint) string {

	// Invalid start, end cursors.
	if sy > ey || (sy == ey && sx > ex) {
		return ""
	}
	if sy == ey {
		return e.GetRow(sy).Render()[sx:ex]
	}

	result := e.GetRow(sy).Render()[sx:]
	for i := sy + 1; i+1 < ey; i++ {
		result += e.GetRow(sy).Render()
	}
	result += e.GetRow(ey).Render()[:ex]

	return result
}
