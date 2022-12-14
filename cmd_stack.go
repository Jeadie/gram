package main

type UndoCmdFn func(e *Editor) error
type RedoCmdFn func(e *Editor) error

type CommandFn struct {
	u UndoCmdFn
	r RedoCmdFn
}

// CommandHistory holds comamnds run in an editor for undo and redo commands.
type CommandHistory struct {
	undo []CommandFn
	redo []CommandFn
}

func CreateCommandHistory() *CommandHistory {
	return &CommandHistory{
		undo: make([]CommandFn, 0),
		redo: make([]CommandFn, 0),
	}
}

func (cs *CommandHistory) pop() (CommandFn, bool) {
	if len(cs.undo) == 0 {
		return CommandFn{}, false
	}

	c := cs.undo[len(cs.undo)-1]
	cs.undo = cs.undo[:len(cs.undo)-1]
	return c, true
}

// Undo the previous command
func (cs *CommandHistory) Undo(e *Editor) {
	c, exists := cs.pop()
	if exists {
		c.u(e)

		// Add applied undo function to redo.
		cs.redo = append(cs.redo, c)
	}
}

// Redo the previously undone command.
func (cs *CommandHistory) Redo(e *Editor) {
	if len(cs.redo) == 0 {
		return
	}

	c := cs.undo[len(cs.redo)-1]
	c.r(e)

	// Add applied redo function to undo.
	cs.undo = append(cs.undo, c)
}

// AddCmd adds the commands to the Command history.
func (cs *CommandHistory) AddCmd(u UndoCmdFn, r RedoCmdFn) {
	cs.undo = append(cs.undo, CommandFn{u: u, r: r})
}

func (cs *CommandHistory) Depth() uint {
	return uint(len(cs.undo))
}
