package main

import (
	"fmt"
	"strings"
)

type Colour string

var AllowedColours = map[string]Colour{"Black": Black, "Default": Default, "Red": Red, "DarkRed": DarkRed, "Green": Green, "DarkGreen": DarkGreen, "DarkYellow": DarkYellow, "Orange": Orange, "Blue": Blue, "DarkBlue": DarkBlue, "Cyan": Cyan, "Magenta": Magenta, "DarkMagenta": DarkMagenta, "DarkCyan": DarkCyan, "LightGray": LightGray, "DarkGray": DarkGray, "White": White}

// Cprintf is Printf with colours. Replace any `%d` with `%Red.d` and it will be printed in Red, or
// wrap any string with `%Red%...%`. E.g. "%Red%HelloWorld%".
func Cprintf(format string, a ...any) (n int, err error) {
	parsed := doCprintfParse(format)
	return fmt.Printf(parsed, a...)
}

// C wraps a string with colour encoding characters. Resets colour at end of string.
func C(s string, c Colour) string {
	return fmt.Sprintf("%s%s%s", c, s, Reset)
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

		// Section can be just a Colour, "%Red"
		v, exists := AllowedColours[s]
		if exists {
			if i+2 >= len(sections) {
				panic("Bad") // TODO: handle errors properly.
			}
			text := sections[i+1]
			result = append(result, C(text, v))    // Coloured content
			result = append(result, sections[i+2]) // next section will be plaintext as its "%" signified end of previous coloured section.
			i += 2
		} else if strings.Index(s, ".") != -1 {
			// Or it can ba a Color with a placeholder, e.g. Red.s
			split := strings.SplitN(s, ".", 2)
			c := split[0]
			v, exists = AllowedColours[c]

			// If start is a colour
			if exists {
				verb, remainingText := extractPrintfVerb(split[1])
				result = append(result, C(verb, v))
				if len(remainingText) > 0 {
					result = append(result, remainingText)
				}
			} else {
				result = append(result, fmt.Sprintf("%%%s", s))
			}

		} else {
			// Error has occurred
			result = append(result, fmt.Sprintf("%%%s", s))
		}
		i++
	}
	return strings.Join(result, "")
}

func extractPrintfVerb(t string) (string, string) {
	// TODO: Implement this better.
	if len(t) == 0 {
		return "", ""
	}
	verb := fmt.Sprintf("%%%s", string(t[0]))
	ext := map[string]uint{"%c": 2, "%d": 2, "%e": 2, "%f": 2, "%i": 2, "%o": 2, "%s": 2, "%u": 2, "%x": 2}

	_, exists := ext[verb]
	if exists {
		if len(t) > 1 {
			return verb, t[1:]
		} else {
			return verb, ""
		}
	}
	return "", ""
}

// ApplyColours per character, onto a string.
func ApplyColours(s string, hl []Colour) string {
	if len(s) == 0 {
		return ""
	} else if len(s) != len(hl) {
		return s
	}

	result := ""

	currI := 0 // Start of current highlight segment
	currC := hl[0]

	for i, c := range hl {
		// Highlighting has changed, apply previous
		if c != currC {
			result += C(s[currI:i], currC)
			currI = i
			currC = c
		}
	}
	// Final un/coloured section
	result += C(s[currI:], currC)

	return result
}
