package uci

import "fmt"

// item represents a lexeme (token)
//
// https://talks.golang.org/2011/lex.slide#8
type item struct {
	typ itemType
	val string
	pos int
}

// itemType defines the kind of lexed item
//
// https://talks.golang.org/2011/lex.slide#9
type itemType int

// these items define the UCI language
const (
	itemError itemType = iota // error occured; item.val is text of error

	itemBOF // begin of file; lexing starts here
	itemEOF // end of file; lexing ends here

	itemPackage // package keyword
	itemConfig  // config keyword
	itemOption  // option keyword
	itemList    // list keyword
	itemIdent   // identifier string
	itemString  // quoted string
)

func (t itemType) String() string {
	switch t {
	case itemError:
		return "Error"
	case itemBOF:
		return "BOF"
	case itemEOF:
		return "EOF"
	case itemPackage:
		return "Package"
	case itemConfig:
		return "Config"
	case itemOption:
		return "Option"
	case itemList:
		return "List"
	case itemIdent:
		return "Ident"
	case itemString:
		return "String"
	}
	return fmt.Sprintf("%%itemType(%d)", int(t))
}

// keyword represents a special marker of the input: each (trimmed,
// non-empty) line of the input must start with a keywords
type keyword string

// these are the recognized keywords.
const (
	kwPackage = keyword("package")
	kwConfig  = keyword("config")
	kwOption  = keyword("option")
	kwList    = keyword("list")
)

// String implements fmt.Stringer interface. Useful for debugging
//
// https://talks.golang.org/2011/lex.slide#11
func (i item) String() string {
	if i.typ != itemError && len(i.val) > 25 {
		return fmt.Sprintf("(%s %.25q... %d)", i.typ, i.val, i.pos)
	}
	return fmt.Sprintf("(%s %q %d)", i.typ, i.val, i.pos)
}
