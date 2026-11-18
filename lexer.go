package uci

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const eof = -1

// stateFn represents the state of the scanner as a function that returns
// the next state.
//
// https://talks.golang.org/2011/lex.slide#19, #27 and following
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
//
// https://talks.golang.org/2011/lex.slide#22
type lexer struct {
	name  string    // used only in error reports
	input string    // the string being scanned
	start int       // start position of the current item
	pos   int       // current position in the input
	width int       // width of last rune read from input
	state stateFn   // current state (see *lexer.nextItem())
	items chan item // channel of scanned items
}

// lex starts the lexer
//
// https://talks.golang.org/2011/lex.slide#41
func lex(name, input string) *lexer {
	return &lexer{
		name:  name,
		input: input,
		state: lexKeyword,
		items: make(chan item, 2),
	}
}

// nextItem returns the next item from the input
//
// https://talks.golang.org/2011/lex.slide#41
func (l *lexer) nextItem() item {
	for l.state != nil {
		select {
		case it, ok := <-l.items:
			if ok {
				return it
			}
			return l.eof()

		default:
			s := l.state(l)
			l.state = s
			if s == nil {
				return l.stop()
			}
		}
	}
	return l.eof()
}

// stop closes the item channel, so that nextItem will (eventually)
// return an EOF token.
func (l *lexer) stop() item {
	it := l.eof()
	if l.items == nil {
		return it
	}
	if len(l.items) > 0 {
		it = <-l.items
	}
	close(l.items)
	l.items = nil
	return it
}

// eof directly returns an EOF token.
func (l *lexer) eof() item {
	return item{itemEOF, l.input[l.start:l.pos], l.pos}
}

// emit emits a token
//
// https://talks.golang.org/2011/lex.slide#25
func (l *lexer) emit(t itemType) {
	if l.pos > l.start {
		l.items <- item{t, l.input[l.start:l.pos], l.pos}
		l.start = l.pos
	}
}

// next returns the next rune in the input
//
// https://talks.golang.org/2011/lex.slide#31
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// ignore skips over the pending input before this point.
//
// https://talks.golang.org/2011/lex.slide#32
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune. Can be called only once per call of next.
//
// https://talks.golang.org/2011/lex.slide#32
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// acceptRun consumes a run of runes from the valid set.
//
// https://talks.golang.org/2011/lex.slide#34
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// acceptComment consumes a line comment starting with # until eol or eof.
func (l *lexer) acceptComment() {
	if l.next() == '#' {
		for {
			r := l.next()
			if r == '\n' || r == eof {
				break
			}
		}
	}
	l.backup()
}

// acceptIdent consumes an UCI identifier [-_a-zA-Z0-9].
func (l *lexer) acceptIdent() {
	for {
		r := l.next()
		if r != '-' && r != '_' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			l.backup()
			break
		}
	}
}

// consumeWhitespace consumes (and ignores) space and tab characters.
func (l *lexer) consumeWhitespace() {
	for isSpace(l.peek()) {
		l.next()
	}
	l.ignore()
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// errorf returns an error token and terminates the scan by passing back
// a nil pointer that will be the next state, terminating l.run.
//
// https://talks.golang.org/2011/lex.slide#37
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...), l.pos}
	return nil
}

// rest returns the not-yet ingested part of l.input.
func (l *lexer) rest() string {
	return l.input[l.pos:]
}

func lexKeyword(l *lexer) stateFn {
	l.acceptRun(" \t\n")
	l.ignore()
	switch curr := l.rest(); {
	case strings.HasPrefix(curr, "#"):
		return lexComment
	case strings.HasPrefix(curr, string(kwPackage)):
		return lexPackage
	case strings.HasPrefix(curr, string(kwConfig)):
		return lexConfig
	case strings.HasPrefix(curr, string(kwOption)):
		return lexOption
	case strings.HasPrefix(curr, string(kwList)):
		return lexList
	}
	if l.next() == eof {
		l.emit(itemEOF)
	} else {
		l.backup()
		unexpected := l.rest()
		if len(unexpected) > 10 {
			unexpected = unexpected[:10] + "…"
		}
		l.errorf("expected keyword (package, config, option, list) or eof, got %q", unexpected)
	}
	return nil
}

func lexComment(l *lexer) stateFn {
	l.acceptComment()
	l.ignore()
	return lexKeyword
}

func lexPackage(l *lexer) stateFn {
	l.pos += len(kwPackage)
	l.emit(itemPackage)
	return lexPackageName
}

func lexPackageName(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("incomplete package name")
		case isSpace(r):
			l.ignore()
		case r == '\'' || r == '"':
			l.backup()
			return lexValue
		}
	}
}

func lexConfig(l *lexer) stateFn {
	l.pos += len(kwConfig)
	l.emit(itemConfig)
	l.consumeWhitespace()
	return lexConfigType
}

func lexConfigType(l *lexer) stateFn {
	l.acceptIdent()
	l.emit(itemIdent)
	l.consumeWhitespace()
	return lexOptionalName
}

func lexOptionalName(l *lexer) stateFn {
	switch l.next() {
	case '\n':
		l.ignore()
	case '"', '\'':
		l.backup()
		return lexValue
	default:
		l.acceptIdent()
		l.emit(itemString)
	}
	return lexKeyword
}

func lexOption(l *lexer) stateFn {
	l.pos += len(kwOption)
	l.emit(itemOption)
	l.consumeWhitespace()
	return lexOptionName
}

func lexList(l *lexer) stateFn {
	l.pos += len(kwList)
	l.emit(itemList)
	l.consumeWhitespace()
	return lexOptionName
}

func lexOptionName(l *lexer) stateFn {
	l.acceptIdent()
	l.emit(itemIdent)
	l.consumeWhitespace()
	return lexValue
}

// lexValue scans a string value, following the quoting rules of libuci:
// A value consists of adjacent parts which are concatenated. Single-quoted
// parts are taken literally. In double-quoted and unquoted parts, a
// backslash escapes the next character, and a backslash before a line
// break continues the value on the next line. Example (yields "it's"):
//
//	'it'\''s'
func lexValue(l *lexer) stateFn {
	var sb strings.Builder
	for {
		switch r := l.next(); r {
		case '\'', '"':
			if !l.scanQuoted(&sb, r) {
				return l.errorf("unterminated quoted string")
			}
		case '\\':
			if !l.scanEscaped(&sb) {
				return l.errorf("unterminated unquoted string")
			}
		case eof, ' ', '\t', '\n', '#':
			l.backup()
			if l.pos > l.start {
				l.items <- item{itemString, sb.String(), l.pos}
				l.start = l.pos
			}
			l.consumeWhitespace()
			return lexKeyword
		default:
			sb.WriteRune(r)
		}
	}
}

// scanQuoted consumes a quoted part (after the opening quote q) up to and
// including the closing quote, and writes its content to sb. It returns
// false if the input ends before the closing quote.
func (l *lexer) scanQuoted(sb *strings.Builder, q rune) bool {
	for {
		switch r := l.next(); {
		case r == eof:
			return false
		case r == q:
			return true
		case r == '\\' && q == '"':
			if !l.scanEscaped(sb) {
				return false
			}
		default:
			sb.WriteRune(r)
		}
	}
}

// scanEscaped consumes the character following a backslash and writes it
// to sb. An escaped line break is dropped. It returns false at eof.
func (l *lexer) scanEscaped(sb *strings.Builder) bool {
	switch r := l.next(); r {
	case eof:
		return false
	case '\n':
		return true
	default:
		sb.WriteRune(r)
		return true
	}
}
