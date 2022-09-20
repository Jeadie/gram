package main

import (
	"regexp"
	"strings"
	"unicode"
)

type Syntax struct {
	l     LanguageSyntax
	cache LRUCache
}

type LanguageSyntax struct {
	exts        []string
	keywords    []string
	stringChars []byte
	comment     string
}

var pythonSyntax = LanguageSyntax{
	exts:        []string{".py"},
	keywords:    []string{"False", "None", "True", "and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else", "except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield"},
	comment:     "#",
	stringChars: []byte{'"', '\''},
}

var goSyntax = LanguageSyntax{
	exts:        []string{".go"},
	keywords:    []string{"uint", "import", "package", "const", "var", "func", "map", "string", "byte", "struct", "int", "any", "error", "type", "continue", "break", "append", "if", "len", "return", "else"},
	comment:     "//",
	stringChars: []byte{'"', '\'', '`'},
}
var defaultSyntax = LanguageSyntax{
	exts:        []string{""},
	keywords:    []string{},
	comment:     "#",
	stringChars: []byte{'"'},
}

var syntaxs = []LanguageSyntax{goSyntax, pythonSyntax}

func CreateSyntax(filename string) *Syntax {
	return &Syntax{
		l:     GetLanguageSyntax(filename),
		cache: *CreateCache(100),
	}
}

// Highlight string according to a given highlighting syntax
func (s *Syntax) Highlight(x string) string {
	v, exists := s.cache.Get(x)
	if exists {
		return v
	}

	// Cache miss
	v = s.ApplySyntax(x)
	s.cache.Set(x, v)
	return v
}

// GetLanguageSyntax of file based on filename.
func GetLanguageSyntax(filename string) LanguageSyntax {
	for _, syntax := range syntaxs {
		if FileHasExtension(filename, syntax.exts) {
			return syntax
		}
	}
	return defaultSyntax
}

func FileHasExtension(filename string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
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

func (s *Syntax) ApplySyntax(x string) string {
	hl := make([]Colour, len(x))

	// Keyword highlights
	for _, k := range s.l.keywords {
		for _, idx := range AllWordIndices(x, k) {
			for i := 0; i < len(k); i++ {
				hl[idx+i] = Orange
			}
		}
	}

	// Comments
	cIdx := strings.Index(x, s.l.comment)
	if cIdx != -1 {
		for i := cIdx; i < len(x); i++ {
			hl[i] = DarkGray
		}
	}

	// TODOs
	toDoIdx := strings.Index(x, s.l.comment+" TODO")
	if toDoIdx != -1 {
		for i := toDoIdx + len(s.l.comment); i < len(x); i++ {
			hl[i] = DarkYellow
		}
	}

	// Numbers
	HighlightRegex(x, "[-]?\\d[\\d,]*[\\.]?[\\d{2}]*", &hl, Blue)

	// Strings
	// TODO: Fix Multiple strings in same line interpolating incorrectly
	for _, b := range s.l.stringChars {
		HighlightString(x, string(b), &hl, DarkGreen)
	}

	return ApplyColours(x, hl)
}

func HighlightString(s, char string, hl *[]Colour, c Colour) {
	re, _ := regexp.Compile(char)
	results := re.FindAllIndex([]byte(s), -1)

	//  Every second result to avoid highlighting between strings.
	for j := 0; j+1 < len(results); j += 2 {
		idx := results[j]
		idx2 := results[j+1]
		for i := idx[0]; i <= idx2[0]; i++ {
			(*hl)[i] = c
		}
	}
}

func HighlightRegex(s, regex string, hl *[]Colour, c Colour) {
	re, _ := regexp.Compile(regex)
	for _, idx := range re.FindAllIndex([]byte(s), -1) {
		for i := idx[0]; i < idx[1]; i++ {
			(*hl)[i] = c
		}
	}
}
