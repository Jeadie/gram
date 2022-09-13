package main

import (
	"fmt"
	"strings"
)

type Colour string

const (
	Reset       Colour = "\033[0m"
	Default            = "\033[39m"
	Black              = "\033[30m"
	Red                = "\033[91m"
	DarkRed            = "\033[31m"
	Green              = "\033[92m"
	DarkGreen          = "\033[32m"
	DarkYellow         = "\033[33m"
	Orange             = "\033[93m"
	Blue               = "\033[94m"
	DarkBlue           = "\033[34m"
	Cyan               = "\033[96m"
	Magenta            = "\033[95m"
	DarkMagenta        = "\033[35m"
	DarkCyan           = "\033[36m"
	LightGray          = "\033[37m"
	DarkGray           = "\033[90m"
	White              = "\033[97m"
)

var AllowedColours = map[string]Colour{"Black": Black, "Default": Default, "Red": Red, "DarkRed": DarkRed, "Green": Green, "DarkGreen": DarkGreen, "DarkYellow": DarkYellow, "Orange": Orange, "Blue": Blue, "DarkBlue": DarkBlue, "Cyan": Cyan, "Magenta": Magenta, "DarkMagenta": DarkMagenta, "DarkCyan": DarkCyan, "LightGray": LightGray, "DarkGray": DarkGray, "White": White}

// Cprintf is Printf with colours. Replace any `%d` with `%Red.d` and it will be printed in Red, or
// wrap any string with `%Red%...%`. E.g. "%Red%HelloWorld%".
func Cprintf(format string, a ...any) (n int, err error) {
	parsed := doCprintfParse(format)
	return fmt.Printf(parsed, a...)
}

func doCprintfParse(x string) string {
	sections := strings.Split(x, "%")
	result := make([]string, 0)
	if len(sections) <= 1 {
		return x
	}

	// If we don't start with a %, add it as raw.
	if len(sections[0]) > 0 {
		result = append(result, sections[0])
	}

	sections = sections[1:] // Either zeroth section has been added or is ''

	i := 0
	for i < len(sections) {

		s := sections[i] // Allows us to increment i within loop

		// Section can be just a Colour
		v, exists := AllowedColours[s]
		if exists {
			if i+2 >= len(sections) {
				panic("Bad")
			}
			text := sections[i+1]
			result = append(result, C(text, v))
			i += 2
		} else if strings.Index(s, ".") != -1 {
			// Or it can ba a Color with a placeholder, e.g. Red.s
			split := strings.SplitN(s, ".", 2)
			c := split[0]
			placeholder := split[1]
			v, exists = AllowedColours[c]
			if exists {
				result = append(result, C("%"+placeholder, v))
			}

		} else {
			// Error has occurred
			panic("bad")
		}
		i++
	}
	return strings.Join(result, "")
}

// C wraps a string with colour encoding characters. Resets colour at end of string.
func C(s string, c Colour) string {
	return fmt.Sprintf("%s%s%s", c, s, Reset)
}
