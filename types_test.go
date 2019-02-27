package uci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmangleSectionName(t *testing.T) {
	tt := map[string]struct {
		typ string
		idx int
		err string
	}{
		// simple test cases
		"":            {err: "implausible section selector: must be at least 5 characters long"},
		"aa[0]":       {err: "invalid syntax: section selector must start with @ sign"},
		"@@[0]":       {err: "invalid syntax: multiple @ signs found"},
		"@@@@@@@@@@@": {err: "invalid syntax: multiple @ signs found"},
		"@[[0]":       {err: "invalid syntax: multiple open brackets found"},
		"@][0]":       {err: "invalid syntax: multiple closed brackets found"},
		"@aa0]":       {err: "invalid syntax: section selector must have format '@type[index]'"},
		"@a[b]":       {err: `strconv.Atoi: parsing "b": invalid syntax`},

		// valid test cases
		"@a[0]":    {typ: "a", idx: 0},
		"@a[4223]": {typ: "a", idx: 4223},
		"@a[-1]":   {typ: "a", idx: -1},

		// longer types/indices
		"@abcdEFGHijkl[-255]": {typ: "abcdEFGHijkl", idx: -255},
		"@abcdEFGHijkl[0xff]": {err: `strconv.Atoi: parsing "0xff": invalid syntax`},
	}

	for input, tc := range tt {
		t.Run(input, func(t *testing.T) {
			assert := assert.New(t)
			typ, idx, err := unmangleSectionName(input)

			if tc.err != "" {
				assert.EqualError(err, tc.err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.idx, idx)
				assert.Equal(tc.typ, typ)
			}
		})
	}
}
