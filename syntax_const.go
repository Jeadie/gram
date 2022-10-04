package main

var builtinLanguageSyntaxs = []LanguageSyntax{
	{
		Exts:        nil,
		Keywords:    nil,
		StringChars: nil,
		Comment:     "",
		HlStrings:   false,
		HlNumbers:   false,
	}, {
		Exts:        []string{".go"},
		Keywords:    []string{"bool", "uint", "import", "package", "const", "var", "func", "map", "string", "byte", "struct", "int", "any", "error", "type", "continue", "break", "append", "if", "len", "return", "else"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "//",
		HlStrings:   true,
		HlNumbers:   true,
	}, {
		Exts:        []string{".py"},
		Keywords:    []string{"False", "None", "True", "and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else", "except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "#",
		HlStrings:   true,
		HlNumbers:   true,
	},
	{
		Exts:        []string{".sh"},
		Keywords:    []string{"if", "fi", "elif", "case", "esac", "then"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "#",
		HlStrings:   true,
		HlNumbers:   true,
	},
}
