package uci

import (
	"os"
	"strings"
)

// test helper and common test cases for lexer/parser
//
// XXX: This file is named test_test.go, because `go test` ignores
// files with prefix "_", including "_test.go"... I'm open for less
// stupid names.

// control via DUMP env var, which details should be printed out. Use
// something like
//
//	DUMP="lex,token" go test -v ./...
var dumpJSON, dumpToken, dumpLexemes bool

func init() {
	for _, field := range strings.Split(os.Getenv("DUMP"), ",") {
		dumpJSON = dumpJSON || field == "all" || field == "json"
		dumpToken = dumpToken || field == "all" || field == "token"
		dumpLexemes = dumpLexemes || field == "all" || field == "lex"
	}
}

func (t scanToken) mk(items ...item) token {
	return token{t, items}
}

func (i itemType) mk(val string) item {
	return item{i, val, -1}
}

const tcEmptyInput1 = ""
const tcEmptyInput2 = "  \n\t\n\n \n "
const tcSimpleInput = `config sectiontype 'sectionname'
	option optionname 'optionvalue'
`
const tcExportInput = `package "pkgname"
config empty
config squoted 'sqname'
config dquoted "dqname"
config multiline 'line1\
	line2'
`
const tcUnquotedInput = "config foo bar\noption answer 42\n"

var lexerTests = []struct {
	name, input string
	expected    []item
}{
	{"empty1", tcEmptyInput1, []item{}},
	{"empty2", tcEmptyInput2, []item{}},
	{"simple", tcSimpleInput, []item{
		itemConfig.mk("config"), itemIdent.mk("sectiontype"), itemString.mk("sectionname"),
		itemOption.mk("option"), itemIdent.mk("optionname"), itemString.mk("optionvalue"),
	}},
	{"export", tcExportInput, []item{
		itemPackage.mk("package"), itemString.mk("pkgname"),
		itemConfig.mk("config"), itemIdent.mk("empty"),
		itemConfig.mk("config"), itemIdent.mk("squoted"), itemString.mk("sqname"),
		itemConfig.mk("config"), itemIdent.mk("dquoted"), itemString.mk("dqname"),
		itemConfig.mk("config"), itemIdent.mk("multiline"), itemString.mk("line1\\\n\tline2"),
	}},
	{"unquoted", tcUnquotedInput, []item{
		itemConfig.mk("config"), itemIdent.mk("foo"), itemString.mk("bar"),
		itemOption.mk("option"), itemIdent.mk("answer"), itemString.mk("42"),
	}},
}

var parserTests = []struct {
	name, input string
	expected    []token
}{
	{"empty1", "", []token{}},
	{"empty2", "  \n\t\n\n \n ", []token{}},
	{"simple", tcSimpleInput, []token{
		tokSection.mk(itemIdent.mk("sectiontype"), itemString.mk("sectionname")),
		tokOption.mk(itemIdent.mk("optionname"), itemString.mk("optionvalue")),
	}},
	{"export", tcExportInput, []token{
		tokPackage.mk(itemString.mk("pkgname")),
		tokSection.mk(itemIdent.mk("empty")),
		tokSection.mk(itemIdent.mk("squoted"), itemString.mk("sqname")),
		tokSection.mk(itemIdent.mk("dquoted"), itemString.mk("dqname")),
		tokSection.mk(itemIdent.mk("multiline"), itemString.mk("line1\\\n\tline2")),
	}},
	{"unquoted", tcUnquotedInput, []token{
		tokSection.mk(itemIdent.mk("foo"), itemString.mk("bar")),
		tokOption.mk(itemIdent.mk("answer"), itemString.mk("42")),
	}},
}
