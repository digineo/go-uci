package uci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemTypeString(t *testing.T) {
	assert := assert.New(t)
	names := []string{
		"Error",
		"BOF",
		"EOF",
		"Package",
		"Config",
		"Option",
		"List",
		"Ident",
		"String",
		"%itemType(9)",
	}

	for i, expected := range names {
		subject := ItemType(i)
		assert.Equal(expected, subject.String())
	}
}

func TestScanTokenString(t *testing.T) {
	assert := assert.New(t)
	names := []string{
		"error",
		"eof",
		"package",
		"config",
		"option",
		"list",
		"%scanToken(6)",
	}

	for i, expected := range names {
		subject := scanToken(i)
		assert.Equal(expected, subject.String())
	}
}

func TestItemString(t *testing.T) {
	assert := assert.New(t)

	tt := []struct {
		subject  item
		expected string
	}{
		{
			item{itemString, "foo", -1},
			`(String "foo")`,
		}, {
			item{itemString, "foo 0123456789 bar 0123456789", -1},
			`(String "foo 0123456789 bar 012345"...)`,
		}, {
			item{itemError, "foo 0123456789 bar 0123456789", -1},
			`(Error "foo 0123456789 bar 0123456789")`,
		},

		{
			item{itemString, "foo", 42},
			`(String "foo" 42)`,
		}, {
			item{itemString, "foo 0123456789 bar 0123456789", 42},
			`(String "foo 0123456789 bar 012345"... 42)`,
		}, {
			item{itemError, "foo 0123456789 bar 0123456789", 42},
			`(Error "foo 0123456789 bar 0123456789" 42)`,
		},
	}

	for i := range tt {
		tc := tt[i]
		assert.Equal(tc.expected, tc.subject.String())
	}
}

func TestTokenString(t *testing.T) {
	assert := assert.New(t)

	tt := []struct {
		subject  token
		expected string
	}{
		{
			token{tokList, nil},
			`list[]`,
		}, {
			token{tokPackage, []item{
				{itemIdent, "foo", -1},
			}},
			`package[(Ident "foo")]`,
		}, {
			token{tokPackage, []item{
				{itemIdent, "foo", 42},
			}},
			`package[(Ident "foo" 42)]`,
		}, {
			token{tokPackage, []item{
				{itemIdent, "foo", 42},
				{itemIdent, "bar", -1},
			}},
			`package[(Ident "foo" 42) (Ident "bar")]`,
		},
	}
	for i := range tt {
		tc := tt[i]
		assert.Equal(tc.expected, tc.subject.String())
	}
}
