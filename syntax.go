package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

type Syntax struct {
	l     LanguageSyntax
	cache LRUCache
	c     ColourScheme
}

type LanguageSyntax struct {
	Exts        []string `json:"extensions"`
	Keywords    []string `json:"Keywords"`
	StringChars []string `json:"stringCharacters"`
	Comment     string   `json:"commentCharacter"`
	HlStrings   bool     `json:"highlightStrings"`
	HlNumbers   bool     `json:"highlightNumbers"`
}

var defaultSyntax = LanguageSyntax{
	Exts:        []string{""},
	Keywords:    []string{},
	Comment:     "#",
	StringChars: []string{"\""},
	HlNumbers:   false,
	HlStrings:   false,
}

func CreateSyntax(filename string) *Syntax {
	return &Syntax{
		l:     GetLanguageSyntax(filename),
		cache: *CreateCache(100),
		c:     GetColourScheme(),
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

func LoadSyntaxesFromFile(file string) []LanguageSyntax {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return []LanguageSyntax{}
	}

	var syntaxes []LanguageSyntax
	err = json.Unmarshal(bytes, &syntaxes)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}
	return syntaxes
}

// GetLanguageSyntax of file based on filename.
func GetLanguageSyntax(filename string) LanguageSyntax {
    syntaxes := append(LoadSyntaxesFromFile("syntax.json"), builtinLanguageSyntaxs...)
	for _, syntax := range syntaxes {
		if FileHasExtension(filename, syntax.Exts) {
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
	for _, k := range s.l.Keywords {
		for _, idx := range AllWordIndices(x, k) {
			for i := 0; i < len(k); i++ {
				hl[idx+i] = s.c.Keyword
			}
		}
	}

	// Comments
	if len(s.l.Comment) > 0 {

		cIdx := strings.Index(x, s.l.Comment)
		if cIdx != -1 {
			for i := cIdx; i < len(x); i++ {
				hl[i] = s.c.Comments
			}
		}
	}

	// TODOs
	toDoIdx := strings.Index(x, s.l.Comment+" TODO")
	if toDoIdx != -1 {
		for i := toDoIdx + len(s.l.Comment); i < len(x); i++ {
			hl[i] = s.c.Todos
		}
	}

	// Numbers
	if s.l.HlNumbers {
		HighlightRegex(x, "[-]?\\d[\\d,]*[\\.]?[\\d{2}]*", &hl, s.c.Numbers)
	}

	// Strings
	if s.l.HlStrings {
		for _, b := range s.l.StringChars {
			HighlightString(x, string(b), &hl, s.c.Strings)
		}
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
