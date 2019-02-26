package uci

import (
	"errors"
	"fmt"
)

// scanner is intertwined with lexer and groups lexemes into token
// (which is a typed list of items).
//
// Refer to the implementation of lexer for hints about the design.
// The scanner is strongly modeled after the same principles, although
// a bit less elegant at times.
type scanner struct {
	lexer  *lexer
	state  scanFn
	last   *item  // last item read from the lexer, but deffered by the state
	curr   []item // accepted items
	tokens chan token
}

func scan(name, input string) *scanner {
	return &scanner{
		lexer:  lex(name, input),
		state:  scanStart,
		curr:   make([]item, 0, 3),
		tokens: make(chan token, 2),
	}
}

func (s *scanner) nextToken() token {
	for {
		select {
		case tok, ok := <-s.tokens:
			if ok {
				return tok
			}
			return s.eof()
		default:
			st := s.state(s)
			if st == nil {
				s.stop()
				return s.eof()
			}
			s.state = st
		}
	}
}

func (s *scanner) eof() token {
	return token{typ: tokEOF}
}

func (s *scanner) stop() {
	if s.tokens == nil {
		return
	}
	s.lexer.stop()
	close(s.tokens)
	s.tokens = nil
}

func (s *scanner) next() item {
	if s.last != nil {
		it := *s.last
		s.last = nil
		return it
	}
	return s.lexer.nextItem()
}

func (s *scanner) peek() item {
	it := s.next()
	s.backup(it)
	return it
}

func (s *scanner) backup(it item) {
	s.last = &it
}

func (s *scanner) accept(it itemType) bool {
	tok := s.next()
	if tok.typ == it {
		s.curr = append(s.curr, tok)
		return true
	}
	s.backup(tok)
	return false
}

func (s *scanner) emit(typ scanToken) {
	s.tokens <- token{typ: typ, items: s.curr}
	s.curr = make([]item, 0, 3)
}

func (s *scanner) errorf(format string, args ...interface{}) scanFn {
	s.tokens <- token{
		typ:   tokError,
		items: []item{item{itemError, fmt.Sprintf(format, args...), 0}},
	}
	return nil
}

// scanStart looks for a "package" or "config" item.
func scanStart(s *scanner) scanFn {
	switch tok := s.next(); tok.typ {
	case itemPackage:
		return scanPackage
	case itemConfig:
		return scanSection
	}
	return s.errorf("expected package or config token")
}

// scanPackage looks for a package name
func scanPackage(s *scanner) scanFn {
	if !s.accept(itemString) {
		return s.errorf("expected string value while parsing package")
	}
	s.emit(tokPackage)
	return scanStart
}

// scanSection looks for a section type and optional a name
func scanSection(s *scanner) scanFn {
	if !s.accept(itemIdent) {
		return s.errorf("expected identifier while parsing config section")
	}
	// the name is optional
	if tok := s.peek(); tok.typ == itemString {
		s.accept(itemString)
	}
	s.emit(tokSection)
	return scanOption
}

// scanOption looks for either an "option" or "list" keyword (with name
// and value), or it falls back to scanStart
func scanOption(s *scanner) scanFn {
	tok := s.next()
	switch tok.typ {
	case itemOption, itemList:
		return scanOptionName
	default:
		s.backup(tok)
		return scanStart
	}
}

// scanOptionName looks for a name of an option
func scanOptionName(s *scanner) scanFn {
	if s.accept(itemIdent) {
		return scanOptionValue
	}
	return s.errorf("expected option name")
}

// scanOptionValue looks for the value associated with an option
func scanOptionValue(s *scanner) scanFn {
	if s.accept(itemString) {
		s.emit(tokOption)
		return scanOption
	}
	return s.errorf("expected option value")
}

func (s *scanner) each(fn func(token) bool) bool {
	for tok := s.nextToken(); tok.typ != tokEOF; tok = s.nextToken() {
		if !fn(tok) {
			s.stop()
			return false
		}
	}
	return true
}

// parse tries to parse a named input string into a config object.
func parse(name, input string) (cfg *config, err error) {
	cfg = newConfig(name)
	var sec *section

	scan(name, input).each(func(tok token) bool {
		switch tok.typ {
		case tokPackage:
			err = errors.New("UCI imports/exports are not yet supported")
			return false
		case tokSection:
			if len(tok.items) == 2 {
				sec = newSection(tok.items[0].val, tok.items[1].val)
			} else {
				sec = newSection(tok.items[0].val, "")
			}
			cfg.Merge(sec)
		case tokOption:
			opt := newOption(tok.items[0].val, []string{tok.items[1].val})
			sec.Merge(opt)
		}
		return true
	})
	return
}
