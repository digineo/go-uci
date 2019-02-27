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
var dump = make(map[string]bool)

func init() {
	for _, field := range strings.Split(os.Getenv("DUMP"), ",") {
		if field == "all" {
			dump["json"] = true
			dump["token"] = true
			dump["lex"] = true
			dump["serialized"] = true
		} else {
			dump[field] = true
		}
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

const tcUnnamedInput = `
config foo named
	option pos '0'
	option unnamed '0'
	list list 0

config foo
	option pos '1'
	option unnamed '1'
	list list 10

config foo
	option pos '2'
	option unnamed '1'
	list list 20

config foo named
	option pos '3'
	option unnamed '0'
	list list 30
`

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
	{"unnamed", tcUnnamedInput, []item{
		itemConfig.mk("config"), itemIdent.mk("foo"), itemString.mk("named"),
		itemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("0"),
		itemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("0"),
		itemList.mk("list"), itemIdent.mk("list"), itemString.mk("0"),

		itemConfig.mk("config"), itemIdent.mk("foo"), // unnamed
		itemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("1"),
		itemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("1"),
		itemList.mk("list"), itemIdent.mk("list"), itemString.mk("10"),

		itemConfig.mk("config"), itemIdent.mk("foo"), // unnamed
		itemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("2"),
		itemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("1"),
		itemList.mk("list"), itemIdent.mk("list"), itemString.mk("20"),

		itemConfig.mk("config"), itemIdent.mk("foo"), itemString.mk("named"),
		itemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("3"),
		itemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("0"),
		itemList.mk("list"), itemIdent.mk("list"), itemString.mk("30"),
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
	{"unnamed", tcUnnamedInput, []token{
		tokSection.mk(itemIdent.mk("foo"), itemString.mk("named")),
		tokOption.mk(itemIdent.mk("pos"), itemString.mk("0")),
		tokOption.mk(itemIdent.mk("unnamed"), itemString.mk("0")),
		tokList.mk(itemIdent.mk("list"), itemString.mk("0")),

		tokSection.mk(itemIdent.mk("foo")), // unnamed
		tokOption.mk(itemIdent.mk("pos"), itemString.mk("1")),
		tokOption.mk(itemIdent.mk("unnamed"), itemString.mk("1")),
		tokList.mk(itemIdent.mk("list"), itemString.mk("10")),

		tokSection.mk(itemIdent.mk("foo")), // unnamed
		tokOption.mk(itemIdent.mk("pos"), itemString.mk("2")),
		tokOption.mk(itemIdent.mk("unnamed"), itemString.mk("1")),
		tokList.mk(itemIdent.mk("list"), itemString.mk("20")),

		tokSection.mk(itemIdent.mk("foo"), itemString.mk("named")),
		tokOption.mk(itemIdent.mk("pos"), itemString.mk("3")),
		tokOption.mk(itemIdent.mk("unnamed"), itemString.mk("0")),
		tokList.mk(itemIdent.mk("list"), itemString.mk("30")),
	}},
}
