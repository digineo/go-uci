package uci

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	for _, tc := range lexerTests {
		t.Run(tc.name, func(t *testing.T) {
			testLexer(t, tc.name, tc.input, tc.expected)
		})
	}
}

func testLexer(t *testing.T, name, input string, expected []item) {
	t.Helper()

	if dumpLexemes {
		defer fmt.Println("")
	}

	l := lex(name, input)
	var i int
	for it := l.nextItem(); it.typ != itemEOF; it = l.nextItem() {
		if dumpLexemes {
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
	if l := len(expected); i != l {
		t.Errorf("expected to lex %d items, actually lexed %d", l, i)
	}
}
