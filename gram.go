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
	if err != nil {
		DisableRawMode(t)
		os.Exit(1)
	}
	defer DisableRawMode(t)

	for err == nil && cs > 0 {
		cs, err = os.Stdin.Read(c)

		if string(c[0]) == "q" {
			return
		}
		fmt.Printf(string(c[0]))
	}
}

func DisableRawMode(t unix.Termios) error {
	fmt.Println("DisableRawMode")
	return unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TIOCSETA, &t)
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

	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.IEXTEN | unix.ISIG
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, termios); err != nil {
		return returnT, err
	}

	return returnT, err
}
