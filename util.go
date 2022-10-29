package main

import (
	"errors"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"strings"
)

func Ctrl(b byte) byte {
	return b & 0x1f
}

func Min(x, y uint) uint {
	if x < y {
		return x
	} else {
		return y
	}
}

func isControlChar(x byte) bool {
	return x <= 31 || x == 127
}

func RevertTerminalMode(original *unix.Termios) error {
	return unix.IoctlSetTermios(int(os.Stdin.Fd()), ioctlWriteTermios, original)
}

func EnableRawMode() (unix.Termios, error) {
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

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

// GetWindowSize returns number of rows, then columns. (0, 0) if error occurs
// TODO: Sometimes unix.IoctlGetWinsize will fail. Implement fallback
//
//	https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#window-size-the-hard-way
func GetWindowSize() (uint, uint) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdin.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0
	}
	return uint(ws.Row), uint(ws.Col)
}

func Touch(filename string) error {
	return ioutil.WriteFile(filename, []byte{}, 0666)
}

func OpenOrCreate(filename string) ([]Row, error) {
	if !fileExists(filename) {
		err := Touch(filename)
		if err != nil {
			return []Row{}, err
		}
	}

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return []Row{}, err
	}
	file := strings.ReplaceAll(string(raw), "\r", "\n")
	rawRows := strings.Split(file, "\n")
	rows := make([]Row, len(rawRows))

	for i, s := range rawRows {
		rows[i] = ConstructRow(s)
	}

	return rows, nil
}
