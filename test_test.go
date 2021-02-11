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
var dump = func() map[string]bool {
	var m = make(map[string]bool)
	for _, field := range strings.Split(os.Getenv("DUMP"), ",") {
		if field == "all" {
			m["json"] = true
			m["token"] = true
			m["lex"] = true
			m["serialized"] = true
		} else {
			m[field] = true
		}
	}
	return m
}()

func (t scanToken) mk(items ...item) token {
	return token{t, items}
}

func (i ItemType) mk(val string) item {
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

const tcHyphenatedInput = `
config wifi-device wl0
	option type    'broadcom'
	option channel '6'

config wifi-iface wifi0
	option device 'wl0'
	option mode 'ap'
`

const tcComment = `
# heading

# another heading
config foo
	option opt1 1
	# option opt1 2
	option opt2 3 # baa
	option opt3 hello

# a comment block spanning
# multiple lines, surrounded
# by empty lines

# eof
`

const tcInvalid = `
<?xml version="1.0">
<error message="not a UCI file" />
`

const tcIncompletePackage = `
package
`

const tcUnterminatedQuoted = `
config foo "bar
`

const tcUnterminatedUnquoted = `
config foo
	option opt opt\
`

var lexerTests = []struct {
	name, input string
	expected    []item
}{
	{"empty1", tcEmptyInput1, []item{}},
	{"empty2", tcEmptyInput2, []item{}},
	{"simple", tcSimpleInput, []item{
		itemConfig.mk("config"), itemIdent.mk("sectiontype"), itemString.mk("sectionname"),
		ItemOption.mk("option"), itemIdent.mk("optionname"), itemString.mk("optionvalue"),
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
		ItemOption.mk("option"), itemIdent.mk("answer"), itemString.mk("42"),
	}},
	{"unnamed", tcUnnamedInput, []item{
		itemConfig.mk("config"), itemIdent.mk("foo"), itemString.mk("named"),
		ItemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("0"),
		ItemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("0"),
		ItemList.mk("list"), itemIdent.mk("list"), itemString.mk("0"),

		itemConfig.mk("config"), itemIdent.mk("foo"), // unnamed
		ItemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("1"),
		ItemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("1"),
		ItemList.mk("list"), itemIdent.mk("list"), itemString.mk("10"),

		itemConfig.mk("config"), itemIdent.mk("foo"), // unnamed
		ItemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("2"),
		ItemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("1"),
		ItemList.mk("list"), itemIdent.mk("list"), itemString.mk("20"),

		itemConfig.mk("config"), itemIdent.mk("foo"), itemString.mk("named"),
		ItemOption.mk("option"), itemIdent.mk("pos"), itemString.mk("3"),
		ItemOption.mk("option"), itemIdent.mk("unnamed"), itemString.mk("0"),
		ItemList.mk("list"), itemIdent.mk("list"), itemString.mk("30"),
	}},
	{"hyphenated", tcHyphenatedInput, []item{
		itemConfig.mk("config"), itemIdent.mk("wifi-device"), itemString.mk("wl0"),
		ItemOption.mk("option"), itemIdent.mk("type"), itemString.mk("broadcom"),
		ItemOption.mk("option"), itemIdent.mk("channel"), itemString.mk("6"),
		itemConfig.mk("config"), itemIdent.mk("wifi-iface"), itemString.mk("wifi0"),
		ItemOption.mk("option"), itemIdent.mk("device"), itemString.mk("wl0"),
		ItemOption.mk("option"), itemIdent.mk("mode"), itemString.mk("ap"),
	}},
	{"commented", tcComment, []item{
		itemConfig.mk("config"), itemIdent.mk("foo"), // unnamed
		ItemOption.mk("option"), itemIdent.mk("opt1"), itemString.mk("1"),
		ItemOption.mk("option"), itemIdent.mk("opt2"), itemString.mk("3"),
		ItemOption.mk("option"), itemIdent.mk("opt3"), itemString.mk("hello"),
	}},
	{"invalid", tcInvalid, []item{
		itemError.mk("expected keyword (package, config, option, list) or eof"),
	}},
	{"pkg invalid", tcIncompletePackage, []item{
		itemPackage.mk("package"),
		itemError.mk("incomplete package name"),
	}},
	{"unterminated quoted string", tcUnterminatedQuoted, []item{
		itemConfig.mk("config"), itemIdent.mk("foo"), itemError.mk("unterminated quoted string"),
	}},
	{"unterminated unquoted string", tcUnterminatedUnquoted, []item{
		itemConfig.mk("config"), itemIdent.mk("foo"), // unnamed
		ItemOption.mk("option"), itemIdent.mk("opt"), itemError.mk("unterminated unquoted string"),
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
	{"hyphenated", tcHyphenatedInput, []token{
		tokSection.mk(itemIdent.mk("wifi-device"), itemString.mk("wl0")),
		tokOption.mk(itemIdent.mk("type"), itemString.mk("broadcom")),
		tokOption.mk(itemIdent.mk("channel"), itemString.mk("6")),
		tokSection.mk(itemIdent.mk("wifi-iface"), itemString.mk("wifi0")),
		tokOption.mk(itemIdent.mk("device"), itemString.mk("wl0")),
		tokOption.mk(itemIdent.mk("mode"), itemString.mk("ap")),
	}},
	{"commented", tcComment, []token{
		tokSection.mk(itemIdent.mk("foo")),
		tokOption.mk(itemIdent.mk("opt1"), itemString.mk("1")),
		tokOption.mk(itemIdent.mk("opt2"), itemString.mk("3")),
		tokOption.mk(itemIdent.mk("opt3"), itemString.mk("hello")),
	}},
	{"invalid", tcInvalid, []token{
		tokError.mk(itemError.mk("expected keyword (package, config, option, list) or eof")),
	}},
	{"pkg invalid", tcIncompletePackage, []token{
		tokError.mk(itemError.mk("incomplete package name")),
	}},
	{"unterminated quoted string", tcUnterminatedQuoted, []token{
		tokSection.mk(itemIdent.mk("foo")),
		tokError.mk(itemError.mk("unterminated quoted string")),
	}},
	{"unterminated unquoted string", tcUnterminatedUnquoted, []token{
		tokSection.mk(itemIdent.mk("foo")),
		tokError.mk(itemError.mk("unterminated unquoted string")),
	}},
}
