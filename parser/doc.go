/*
Package parser implements a lexer and parser for UCI config files.

The lexer is heavily inspired by Rob Pike's 2011 GTUG Sydney talk
"Lexical Scanning in Go" (https://talks.golang.org/2011/lex.slide,
https://youtu.be/HxaD_trXwRE), which in turn was a presentation of
an early version of Go's text/template parser. It follows, that this
library borrows code from Go's standard library (BSD-style licensed).

The UCI grammar (for the purpose of this library) is defined as follows:

	uci
		: configDecl*
		| (packageDecl configDecl*)+

	packageDecl
		: "package" value CRLF configDecl* # value:config-name

	configDecl
		: "config" ident value? CRLF optionDecl* # ident:section-type value:section-name

	optionDecl
		: "option" ident value # ident:option-name value:option-value
		| "list" ident value   # ident:option-name value:option-value

	ident
		: [_a-zA-Z0-9]+

	value
		: "'" STRING "'"
		| "\"" STRING "\""
		| ident
*/
package parser
