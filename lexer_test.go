package uci

import (
	"fmt"
	"os"
	"testing"
)

const simple = `config sectiontype 'sectionname'
	option optionname 'optionvalue'
`

const complex = `package "pkgname"
config empty
config squoted 'sqname'
config dquoted "dqname"
config multiline 'line1\
	line2'
`

const unquoted = "config foo bar\noption answer 42\n"

func (i itemType) mk(val string) item {
	return item{i, val, 0}
}

func TestLexer(t *testing.T) {
	tt := []struct {
		name, input string
		expected    []item
	}{
		{"empty1", "", []item{}},
		{"empty2", "  \n\t\n\n \n ", []item{}},
		{"simple", simple, []item{
			itemConfig.mk("config"), itemIdent.mk("sectiontype"), itemString.mk("sectionname"),
			itemOption.mk("option"), itemIdent.mk("optionname"), itemString.mk("optionvalue"),
		}},
		{"complex", complex, []item{
			itemPackage.mk("package"), itemString.mk("pkgname"),
			itemConfig.mk("config"), itemIdent.mk("empty"),
			itemConfig.mk("config"), itemIdent.mk("squoted"), itemString.mk("sqname"),
			itemConfig.mk("config"), itemIdent.mk("dquoted"), itemString.mk("dqname"),
			itemConfig.mk("config"), itemIdent.mk("multiline"), itemString.mk("line1\\\n\tline2"),
		}},
		{"unquoted", unquoted, []item{
			itemConfig.mk("config"), itemIdent.mk("foo"), itemString.mk("bar"),
			itemOption.mk("option"), itemIdent.mk("answer"), itemString.mk("42"),
		}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			testLexer(t, tc.name, tc.input, tc.expected)
		})
	}
}

func testLexer(t *testing.T, name, input string, expected []item) {
	t.Helper()

	dump := os.Getenv("DUMP_LEXEMES") == "1"

	if dump {
		defer fmt.Println("")
	}

	_, ch := lex(name, input)
	var i int

	for it := range ch {
		if dump {
			fmt.Print(it, " ")
		}

		if i >= len(expected) {
			t.Errorf("token %d, unexpected item: %s", i, it)
			return
		}
		if ex := expected[i]; it.typ != ex.typ || it.val != ex.val {
			t.Errorf("token %d, expected %s, got %s", i, ex, it)
			return
		}

		i++
	}
}
