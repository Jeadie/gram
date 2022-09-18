package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Syntax struct {
	l     LanguageSyntax
	cache LRUCache
}

func CreateSyntax(filename string) *Syntax {
	return &Syntax{
		l:     goSyntax,
		cache: *CreateCache(100),
	}
}

func (s *Syntax) Highlight(x string) string {
	return SimpleGolangSyntax(x)
}

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

	// Numbers
	HighlightRegex(s, "[-]?\\d[\\d,]*[\\.]?[\\d{2}]*", &hl, Blue)

	// Strings
	for _, b := range goSyntax.stringChars {
		HighlightRegex(s, fmt.Sprintf("%s.*%s", string(b), string(b)), &hl, DarkGreen)
	}

	return ApplyColours(s, hl)
}

func HighlightRegex(s, regex string, hl *[]Colour, c Colour) {
	re, _ := regexp.Compile(regex)
	for _, idx := range re.FindAllIndex([]byte(s), -1) {
		for i := idx[0]; i < idx[1]; i++ {
			(*hl)[i] = c
		}
	}
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
