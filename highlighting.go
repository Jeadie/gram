package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Colour string

// TODO: use 256 colours from https://developer-book.com/post/definitive-guide-for-colored-text-in-terminal/
const (
	Reset       Colour = "\033[0m"
	Default            = "\033[39m"
	Black              = "\033[30m"
	Red                = "\033[91m"
	DarkRed            = "\033[31m"
	Green              = "\033[92m"
	DarkGreen          = "\033[32m"
	DarkYellow         = "\033[33m"
	Orange             = "\033[38:5:202m" // "\033[93m"
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

type LanguageSyntax struct {
	exts        []string
	keywords    []string
	stringChars []byte
	comment     string
}

var goSyntax = LanguageSyntax{
	exts:        []string{".go"},
	keywords:    []string{"uint", "import", "package", "const", "var", "func", "map", "string", "byte", "struct", "int", "any", "error", "type", "continue", "break", "append", "if", "len", "return", "else"},
	comment:     "//",
	stringChars: []byte{'"', '\'', '`'},
}

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

// AllWordIndices returns all indices of a subword within a larger string.
func AllWordIndices(s, sub string) []int {
	r := make([]int, 0)

	i := strings.Index(s, sub)
	tot := 0

	for i != -1 {
		// TODO: ensure substring is subword before appending (but continue iteration)
		if isWord(s, tot+i, tot+i+len(sub)) {
			r = append(r, tot+i)
		}

		tot += len(sub) + i
		i = strings.Index(s[tot:], sub)
	}
	return r
}

// isWord checks if the word s[start:end] is a standalone word (i.e. not within a word)
func isWord(s string, start, end int) bool {
	if start != 0 {
		c := rune(s[start-1])
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			return false
		}
	}

	if end < len(s) {
		c := rune(s[end])
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			return false
		}
	}

	return true
}

func SimpleGolangSyntax(s string) string {
	hl := make([]Colour, len(s))

	// Keyword highlights
	for _, k := range goSyntax.keywords {
		for _, idx := range AllWordIndices(s, k) {
			for i := 0; i < len(k); i++ {
				hl[idx+i] = Orange
			}
		}
	}

	// Comments
	cIdx := strings.Index(s, goSyntax.comment)
	if cIdx != -1 {
		for i := cIdx; i < len(s); i++ {
			hl[i] = DarkGray
		}
	}

	// TODOs
	toDoIdx := strings.Index(s, goSyntax.comment+" TODO")
	if toDoIdx != -1 {
		for i := toDoIdx + len(goSyntax.comment); i < len(s); i++ {
			hl[i] = DarkYellow
		}
	}

	re, _ := regexp.Compile("[-]?\\d[\\d,]*[\\.]?[\\d{2}]*")
	for _, idx := range re.FindAllIndex([]byte(s), -1) {
		for i := idx[0]; i < idx[1]; i++ {
			hl[i] = Blue
		}
	}

	for _, b := range goSyntax.stringChars {
		re, _ = regexp.Compile(fmt.Sprintf("%s.*%s", string(b), string(b)))
		for _, idx := range re.FindAllIndex([]byte(s), -1) {
			for i := idx[0]; i < idx[1]; i++ {
				hl[i] = DarkGreen
			}
		}
	}

	return ApplyColours(s, hl)
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
			//fmt.Println(currI, i, currC)
			if currC != "" {
				result += C(s[currI:i], currC)
			} else {
				result += s[currI:i]
			}

			currI = i
			currC = c
		}
	}
	if currI+1 != len(s) {
		if currC != "" {
			result += C(s[currI:], currC)
		} else {
			result += s[currI:]
		}
	}

	return result
}

func applyGolangFuncSyntax(s string) string {
	funcs := strings.Index(s, "func ")
	if funcs == -1 || (funcs != 0 && s[funcs] != ' ') {
		return s
	}

	funcColour := C("func ", Orange)
	l := s[:funcs] + funcColour + s[funcs+5:]

	startFunc := strings.Index(l, funcColour)
	endFunc := strings.IndexByte(l, '(')
	if endFunc == -1 {
		return l
	}
	// Highlight from end of "func " to start of bracket (
	return l[:startFunc+len(funcColour)] + C(l[startFunc+len(funcColour):endFunc], Green) + l[endFunc:] // + C(s[startFunc + 5:endFunc], Green)
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
