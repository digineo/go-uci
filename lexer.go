package uci

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const eof = -1

// stateFn represents the state of the scanner as a funtion thta returns
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
	items chan item // channel of scanned items
}

// lex starts the lexer
//
// https://talks.golang.org/2011/lex.slide#23, ignoring #41 because of
// note in #39
func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item, 3),
	}
	go l.run()
	return l, l.items
}

// run lexes the input by executing state functions until the state is
// nil.
//
// https://talks.golang.org/2011/lex.slide#24
func (l *lexer) run() {
	for state := lexKeyword; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit a token
//
// https://talks.golang.org/2011/lex.slide#25
func (l *lexer) emit(t itemType) {
	if l.pos > l.start {
		l.items <- item{t, l.input[l.start:l.pos], l.pos}
		l.start = l.pos
	}
}

// emitString emits a string token. it removes the surrounding quotes
func (l *lexer) emitString(t itemType) {
	if l.pos-1 > l.start+1 {
		l.items <- item{t, l.input[l.start+1 : l.pos-1], l.pos}
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

// accept consumes the next rune if it's from the valid set.
//
// https://talks.golang.org/2011/lex.slide#34
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
//
// https://talks.golang.org/2011/lex.slide#34
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) acceptIdent() {
	for {
		r := l.next()
		if !(r == '_' || 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z') {
			l.backup()
			break
		}
	}
}

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

// rest returns the not-yet ingested part of l.input
func (l *lexer) rest() string {
	return l.input[l.pos:]
}

func lexKeyword(l *lexer) stateFn {
	for {
		l.acceptRun(" \t\n")
		l.ignore()
		switch curr := l.rest(); {
		case strings.HasPrefix(curr, string(kwPackage)):
			l.emit(itemPackage)
			return lexPackage
		case strings.HasPrefix(curr, string(kwConfig)):
			l.emit(itemConfig)
			return lexConfig
		case strings.HasPrefix(curr, string(kwOption)):
			l.emit(itemOption)
			return lexOption
		case strings.HasPrefix(curr, string(kwList)):
			l.emit(itemList)
			return lexList
		}
		if l.next() == eof {
			break
		}
	}
	l.emit(itemEOF)
	return nil
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
			return lexQuoted
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
	switch r := l.next(); {
	case r == '\n':
		l.ignore()
		break
	case r == '"' || r == '\'':
		l.backup()
		return lexQuoted
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

func lexValue(l *lexer) stateFn {
	if r := l.peek(); r == '"' || r == '\'' {
		return lexQuoted
	}
	return lexUnquoted
}

// lexQuote scans a quoted string.
func lexQuoted(l *lexer) stateFn {
	q := l.next()
	if q != '"' && q != '\'' {
		return l.errorf("expected quotation")
	}

Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof {
				break // switch
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case q:
			break Loop
		}
	}
	l.emitString(itemString)
	l.consumeWhitespace()
	return lexKeyword
}

func lexUnquoted(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof {
				break
			}
			fallthrough
		case eof:
			return l.errorf("unterminated unquoted string")
		case '\n':
			break Loop
		}
	}
	l.backup()
	l.emit(itemString)
	l.accept("\n")
	l.ignore()
	return lexKeyword
}
