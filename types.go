package uci

// NOTE: config, section and option types basically are AST nodes for the
// parser.

import "fmt"

// config represents a file in UCI. It consists of sections.
type config struct {
	name string
	sec  []*section
	idx  map[string]int // index into sec
}

// newConfig returns a new config object.
func newConfig(name string) *config {
	return &config{
		name: name,
		sec:  make([]*section, 0, 1),
		idx:  make(map[string]int),
	}
}

// Get fetches a section by name.
//
// TODO: Support for unnamed section notation (@foo[idx]) is pending.
func (c *config) Get(name string) *section {
	i, ok := c.idx[name]
	if !ok {
		return nil
	}
	return c.sec[i]
}

func (c *config) Add(s *section) {
	n := s.name
	if n == "" {
		n = fmt.Sprintf("@%s[%d]", s.typ, c.count(s.typ))
	}

	if i, exists := c.idx[n]; exists {
		c.sec[i] = s
	} else {
		i := len(c.sec)
		c.idx[n] = i
		c.sec = append(c.sec, s)
	}
}

func (c *config) count(typ string) (n int) {
	for _, sec := range c.sec {
		if sec.typ == typ {
			n++
		}
	}
	return
}

// A section represents a group of options in UCI. It may be named or
// unnamed. In the latter case, its synthetic name is constructed from
// the section type and index (e.g. "@system[0]").
type section struct {
	typ  string
	name string
	opts []*option
	idx  map[string]int
}

// newSection returns a new section object.
func newSection(typ, name string) *section {
	return &section{
		typ:  typ,
		name: name,
		opts: make([]*option, 0, 1),
		idx:  make(map[string]int),
	}
}

func (s *section) Add(o *option) {
	n := o.name
	if i, exists := s.idx[n]; exists {
		s.opts[i] = o
	} else {
		i := len(s.opts)
		s.idx[n] = i
		s.opts = append(s.opts, o)
	}
}

// Get fetches an option by name.
func (s *section) Get(name string) *option {
	i, ok := s.idx[name]
	if !ok {
		return nil
	}
	return s.opts[i]
}

// An Option is the key to one or more values. Multiple values indicate
// a list option.
type option struct {
	name   string
	values []string
}

// newOption returns a new option object.
func newOption(name string, values []string) *option {
	return &option{
		name:   name,
		values: values,
	}
}

func (o *option) SetValues(vs ...string) {
	o.values = vs
}

func (o *option) AddValue(v string) {
	o.values = append(o.values, v)
}
