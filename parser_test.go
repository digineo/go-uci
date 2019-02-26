package uci

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	for _, tc := range parserTests {
		t.Run(tc.name, func(t *testing.T) {
			testParser(t, tc.name, tc.input, tc.expected)
		})
	}
}

func testParser(t *testing.T, name, input string, expected []token) {
	t.Helper()

	var i int
	ok := scan(name, input).each(func(tok token) bool {
		if dump["token"] {
			fmt.Println(tok)
		}

		if i >= len(expected) {
			t.Errorf("token %d, unexpected item: %s", i, tok)
			return false
		}
		if ex := expected[i]; tok.typ != ex.typ || !equalItemList(tok.items, ex.items) {
			t.Errorf("token %d\nexpected %s\ngot      %s", i, ex, tok)
			return false
		}

		i++
		return true
	})
	if !ok {
		return
	}

	if l := len(expected); i != l {
		t.Errorf("expected to scan %d token, actually scanned %d", l, i)
	}
}

func equalItemList(a, b []item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].typ != b[i].typ || a[i].val != b[i].val {
			return false
		}
	}
	return true
}
