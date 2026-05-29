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
		"@a[b]":       {err: `invalid syntax: index must be numeric: strconv.Atoi: parsing "b": invalid syntax`},

		// valid test cases
		"@a[0]":    {typ: "a", idx: 0},
		"@a[4223]": {typ: "a", idx: 4223},
		"@a[-1]":   {typ: "a", idx: -1},

		// longer types/indices
		"@abcdEFGHijkl[-255]": {typ: "abcdEFGHijkl", idx: -255},
		"@abcdEFGHijkl[0xff]": {err: `invalid syntax: index must be numeric: strconv.Atoi: parsing "0xff": invalid syntax`},
	}

	for input := range tt {
		input, tc := input, tt[input]
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

func TestConfigDel_Named(t *testing.T) {
	assert := assert.New(t)
	c := newConfig("test")
	// newSection signature: newSection(typ, name)
	c.Sections = append(c.Sections, newSection("peer", "alice"), newSection("peer", "bob"))

	c.Del("alice")

	assert.Len(c.Sections, 1)
	assert.Equal("bob", c.Sections[0].Name)
}

func TestConfigDel_Anonymous(t *testing.T) {
	// Del must remove an anonymous section when addressed via "@type[index]".
	assert := assert.New(t)
	c := newConfig("test")
	c.Sections = append(c.Sections,
		newSection("peer", ""),
		newSection("peer", ""),
		newSection("peer", ""),
	)

	c.Del("@peer[1]")

	assert.Len(c.Sections, 2, "one section should be gone")
	// The remaining sections must still be reachable as @peer[0] and @peer[1].
	assert.NotNil(c.Get("@peer[0]"))
	assert.NotNil(c.Get("@peer[1]"))
	// Three sections → only two must remain.
	assert.Nil(c.Get("@peer[2]"))
}

func TestConfigDel_AnonymousFirst(t *testing.T) {
	// Deleting the first anonymous section must shift the indices of the
	// remaining ones, not leave gaps.
	assert := assert.New(t)
	c := newConfig("test")
	s0 := newSection("peer", "")
	s1 := newSection("peer", "")
	s2 := newSection("peer", "")
	// Give each section a distinguishing option so we can tell them apart.
	s0.Options = []*option{newOption("id", TypeOption, "0")}
	s1.Options = []*option{newOption("id", TypeOption, "1")}
	s2.Options = []*option{newOption("id", TypeOption, "2")}
	c.Sections = append(c.Sections, s0, s1, s2)

	c.Del("@peer[0]")

	assert.Len(c.Sections, 2)
	// What was @peer[1] is now @peer[0].
	got := c.Get("@peer[0]")
	assert.NotNil(got)
	assert.Equal("1", got.Get("id").Values[0])
}

func TestConfigDel_AnonymousLast(t *testing.T) {
	assert := assert.New(t)
	c := newConfig("test")
	c.Sections = append(c.Sections,
		newSection("peer", ""),
		newSection("peer", ""),
	)

	c.Del("@peer[-1]") // negative index: last element

	assert.Len(c.Sections, 1)
	assert.NotNil(c.Get("@peer[0]"))
	assert.Nil(c.Get("@peer[1]"))
}

func TestConfigDel_AnonymousOnly(t *testing.T) {
	// Removing the sole anonymous section must leave an empty Sections slice.
	assert := assert.New(t)
	c := newConfig("test")
	c.Sections = append(c.Sections, newSection("peer", ""))

	c.Del("@peer[0]")

	assert.Len(c.Sections, 0)
}

func TestConfigDel_NotFound(t *testing.T) {
	// Deleting a non-existent name must be a silent no-op.
	assert := assert.New(t)
	c := newConfig("test")
	c.Sections = append(c.Sections, newSection("peer", "alice"))

	c.Del("bob")       // named, not present
	c.Del("@peer[5]")  // anonymous, index out of bounds
	c.Del("@other[0]") // anonymous, type not present

	assert.Len(c.Sections, 1)
}

func TestConfigDel_MixedNamedAndAnonymous(t *testing.T) {
	// The @type[index] selector counts ALL sections of a given type, including
	// named ones. So with [named, anon, anon]:
	//   @peer[0] == "named"
	//   @peer[1] == first anonymous
	//   @peer[2] == second anonymous
	// Deleting @peer[1] removes the first anonymous section; "named" and the
	// second anonymous section must survive.
	assert := assert.New(t)
	c := newConfig("test")
	c.Sections = append(c.Sections,
		newSection("peer", "named"),
		newSection("peer", ""),
		newSection("peer", ""),
	)

	c.Del("@peer[1]") // first anonymous section

	assert.Len(c.Sections, 2)
	assert.NotNil(c.Get("named"))
	// The second anonymous section is now the only one and sits at @peer[1]
	// (index 0 is still occupied by "named").
	assert.NotNil(c.Get("@peer[1]"))
}

func TestConfigGet(t *testing.T) { //nolint:funlen
	config, err := parse("unnamed", tcUnnamedInput)
	assert.NoError(t, err)

	cases := []*section{
		// for fun, tcUnnamedInput starts with a named section. for extra
		// fun, tcUnnamedInput extends the named section at the end.
		{"named", "foo", []*option{
			newOption("pos", TypeOption, "3"), // gets overwritten by last section
			newOption("unnamed", TypeOption, "0"),
			newOption("list", TypeList, "0", "30"), // gets merged with last section
		}},

		// the @foo[0] selector only compares type (foo) and index (0)
		{"@foo[0]", "foo", []*option{ // alias for "named"
			newOption("pos", TypeOption, "3"),
			newOption("unnamed", TypeOption, "0"),
			newOption("list", TypeList, "0", "30"),
		}},
		{"@foo[1]", "foo", []*option{
			newOption("pos", TypeOption, "1"),
			newOption("unnamed", TypeOption, "1"),
			newOption("list", TypeOption, "10"),
		}},
		{"@foo[2]", "foo", []*option{
			newOption("pos", TypeOption, "2"),
			newOption("unnamed", TypeOption, "1"),
			newOption("list", TypeList, "20"),
		}},

		// negative indices count from the end
		{"@foo[-3]", "foo", []*option{ // alias for "@foo[0]" == "named"
			newOption("pos", TypeOption, "3"),
			newOption("unnamed", TypeOption, "0"),
			newOption("list", TypeList, "0", "30"),
		}},
		{"@foo[-2]", "foo", []*option{ // alias for "@foo[1]"
			newOption("pos", TypeOption, "1"),
			newOption("unnamed", TypeOption, "1"),
			newOption("list", TypeList, "10"),
		}},
		{"@foo[-1]", "foo", []*option{ // alias for "@foo[2]"
			newOption("pos", TypeOption, "2"),
			newOption("unnamed", TypeOption, "1"),
			newOption("list", TypeList, "20"),
		}},
	}

	for i := range cases {
		s := cases[i]
		for j := range s.Options {
			o := s.Options[j]
			t.Run("unnamed."+s.Name+"."+o.Name, func(t *testing.T) {
				sec := config.Get(s.Name)
				if !assert.NotNil(t, sec) {
					return
				}

				opt := sec.Get(o.Name)
				if !assert.NotNil(t, opt) {
					return
				}

				assert.EqualValues(t, o.Values, opt.Values)
			})
		}
	}
}
