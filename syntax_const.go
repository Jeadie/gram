package main

var builtinLanguageSyntaxs = []LanguageSyntax{
	{
		Exts:        nil,
		Keywords:    nil,
		StringChars: nil,
		Comment:     "",
		HlStrings:   false,
		HlNumbers:   false,
	},
	{
		Exts:        []string{".java"},
		Keywords:    []string{"abstract", "assert", "boolean", "break", "byte", "case", "catch", "char", "class", "const", "continue", "default", "do", "double", "else", "enum", "extends", "final", "finally", "float", "for", "if", "goto", "implements", "import", "instanceof", "int", "interface", "long", "native", "new", "package", "private", "protected", "public", "return", "short", "static", "strictfp", "super", "switch", "synchronized", "this", "throw", "throws", "transient", "try", "void", "volatile", "while", "_", "exports", "module", "non-sealed", "open", "opens", "permits", "provides", "record", "requires", "sealed", "to", "transitive", "uses", "var", "with", "yield"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "//",
		HlStrings:   true,
		HlNumbers:   true,
	},
	{
		Exts:        []string{".kt"},
		Keywords:    []string{"as", "as?", "break", "class", "continue", "do", "else", "false", "for", "fun", "if", "in", "!in", "interface", "is", "!is", "null", "object", "package", "return", "super", "this", "throw", "true", "try", "typealias", "typeof", "val", "var", "when", "while", "by", "catch", "constructor", "delegate", "dynamic", "field", "file", "finally", "get", "import", "init", "param", "property", "receiver", "set", "setparam", "where", "actual", "abstract", "annotation", "companion", "const", "crossinline", "data", "enum", "expect", "external", "final", "infix", "inline", "inner", "internal", "lateinit", "noinline", "open", "operator", "out", "override", "private", "protected", "public", "reified", "sealed", "suspend", "tailrec", "vararg", "field", "it"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "//",
		HlStrings:   true,
		HlNumbers:   true,
	},
	{
		Exts:        []string{".js", ".jsx"},
		Keywords:    []string{"break", "case", "catch", "class", "const", "continue", "debugger", "default", "delete", "do", "else", "export", "extends", "finally", "for", "function", "if", "import", "in", "instanceof", "new", "return", "super", "switch", "this", "throw", "try", "typeof", "var", "void", "while", "with", "yield", "let", "static", "enum", "await", "implements", "interface", "package", "private", "protected", "public", "null", "true", "false"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "//",
		HlStrings:   true,
		HlNumbers:   true,
	},
	{
		Exts:        []string{".go"},
		Keywords:    []string{"bool", "uint", "import", "package", "const", "var", "func", "map", "string", "byte", "struct", "int", "any", "error", "type", "continue", "break", "append", "if", "len", "return", "else"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "//",
		HlStrings:   true,
		HlNumbers:   true,
	},
	{
		Exts:        []string{".py"},
		Keywords:    []string{"False", "None", "True", "and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else", "except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "#",
		HlStrings:   true,
		HlNumbers:   true,
	},
	{
		Exts:      []string{".xit"},
		Keywords:  []string{"[ ]", "[x]", "[@]", "[~]"},
		Comment:   "#", // These are technically Tags. Current Gram does not support the semantics of xit entirely.
		HlStrings: false,
		HlNumbers: true,
	},
	{
		Exts:        []string{".sh"},
		Keywords:    []string{"if", "fi", "elif", "case", "esac", "then"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "#",
		HlStrings:   true,
		HlNumbers:   true,
	},
	{
		Exts:        []string{".ts", ".tsx"},
		Keywords:    []string{"break", "case", "catch", "class", "const", "continue", "debugger", "default", "delete", "do", "else", "enum", "export", "extends", "false", "finally", "for", "function", "If", "import", "in", "istanceOf", "new", "null", "return", "super", "switch", "this", "throw", "true", "try", "typeOf", "var", "void", "while", "with"},
		StringChars: []string{"'", "\"", "`"},
		Comment:     "//",
		HlStrings:   true,
		HlNumbers:   true,
	},
}