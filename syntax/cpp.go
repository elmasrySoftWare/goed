package syntax

// c++, c
var SyntaxCpp = syntax{
	Extensions: []string{".cpp", ".c", ".h"},
	Patterns: []SyntaxPattern{
		NewSyntaxPattern(`/*`, `*/`, ``, true, StyleComment),
		NewSyntaxPattern(`//`, ``, ``, false, StyleComment),
		NewSyntaxPattern(`"`, `"`, `\`, false, StyleString),
		NewSyntaxPattern(`'`, `'`, `\`, false, StyleString),
	},
	Keywords1: []string{
		"asm",
		"auto",
		"bool",
		"break",
		"case",
		"catch",
		"char",
		"const",
		"const_cast",
		"continue",
		"default",
		"delete",
		"do",
		"double",
		"dynamic_cast",
		"else",
		"explicit",
		"extern",
		"false",
		"float",
		"for",
		"friend",
		"goto",
		"if",
		"inline",
		"int",
		"long",
		"mutable",
		"namespace",
		"new",
		"operator",
		"private",
		"protected",
		"public",
		"register",
		"reinterpret_cast",
		"return",
		"short",
		"signed",
		"sizeof",
		"static",
		"static_cast",
		"switch",
		"template",
		"this",
		"throw",
		"true",
		"try",
		"typeid",
		"typename",
		"union",
		"unsigned",
		"virtual",
		"void",
		"volatile",
		"wchar_t",
		"while",
	},
	Keywords2: []string{
		"class", "enum", "export", "struct", "typedef", "using",
	},
	Keywords3: []string{
		"#define",
		"#elif",
		"#endif",
		"#error",
		"#if",
		"#ifdef",
		"#ifndef",
		"#include",
		"#line",
		"#warning",
		"#undef",
	},
	Symbols1: []string{ // ~ assignment
		">>=", "<<=", "++", "+=", "-=", "*=", "/=", "%=",
		"|=", "&=", "^=", "--", "=",
	},
	Symbols2: []string{ // ~ comparators
		"&&", "||", ">=", "<=", "!=", "==", ">", "<", "!", "?", "?:",
	},
	Symbols3: []string{ // others
		"+", "-", "*", "/", "%", "|", "&", "^", "<<", ">>",
	},
	Separators1: []string{
		"(", ")", "[", "]", "{", "}", "<", ">",
	},
	Separators2: []string{
		",", ".", ";", ":", "->", "->*", ".*", "::",
	},
}
