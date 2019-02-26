package uci

import "fmt"

// NOTE: config, section and option types basically are AST nodes for the
// parser. The JSON struct tags are mainly for development and testing
// purposes: We'er generating JSON dumps of the tree when running tests
// with DUMP="json". After a manual comparison with the corresponding UCI
// file in testdata/, we can use the dumps to read them back as test case
// expectations.

// config represents a file in UCI. It consists of sections.
type config struct {
	Name     string     `json:"name"`
	Sections []*section `json:"sections,omitempty"`
}

// newConfig returns a new config object.
func newConfig(name string) *config {
	return &config{
		Name:     name,
		Sections: make([]*section, 0, 1),
	}
}

// Get fetches a section by name.
//
// TODO: Support for unnamed section notation (@foo[idx]) is pending.
func (c *config) Get(name string) *section {
	for _, sec := range c.Sections {
		if sec.Name == name {
			return sec
		}
	}
	return nil
}

func (c *config) Add(s *section) {
	c.Sections = append(c.Sections, s)
}

func (c *config) Merge(s *section) {
	var sec *section
	for i := range c.Sections {
		sname := c.sectionName(s)
		cname := c.sectionName(c.Sections[i])

		if sname == cname {
			sec = c.Sections[i]
			break
		}
	}

	if sec == nil {
		c.Add(s)
		return
	}
	for _, o := range s.Options {
		sec.Merge(o)
	}
}

func (c *config) Del(name string) {
	var i int
	for i = 0; i < len(c.Sections); i++ {
		if c.Sections[i].Name == name {
			break
		}
	}
	if i < len(c.Sections) {
		c.Sections = append(c.Sections[:i], c.Sections[i+1:]...)
	}
}

func (c *config) sectionName(s *section) string {
	if s.Name != "" {
		return s.Name
	}
	return fmt.Sprintf("@%s[%d]", s.Type, c.count(s.Type))
}

func (c *config) count(typ string) (n int) {
	for _, sec := range c.Sections {
		if sec.Type == typ {
			n++
		}
	}
	return
}

// A section represents a group of options in UCI. It may be named or
// unnamed. In the latter case, its synthetic name is constructed from
// the section type and index (e.g. "@system[0]").
type section struct {
	Name    string    `json:"name,omitempty"`
	Type    string    `json:"type"`
	Options []*option `json:"options,omitempty"`
}

// newSection returns a new section object.
func newSection(typ, name string) *section {
	return &section{
		Type:    typ,
		Name:    name,
		Options: make([]*option, 0, 1),
	}
}

func (s *section) Add(o *option) {
	s.Options = append(s.Options, o)
}

func (s *section) Merge(o *option) {
	for _, opt := range s.Options {
		if opt.Name == o.Name {
			opt.MergeValues(o.Values...)
			return
		}
	}
	s.Options = append(s.Options, o)
}

func (s *section) Del(name string) {
	var i int
	for i = 0; i < len(s.Options); i++ {
		if s.Options[i].Name == name {
			break
		}
	}
	s.Options = append(s.Options[:i], s.Options[i+1:]...)
}

// Get fetches an option by name.
func (s *section) Get(name string) *option {
	for _, opt := range s.Options {
		if opt.Name == name {
			return opt
		}
	}
	return nil
}

// An Option is the key to one or more values. Multiple values indicate
// a list option.
type option struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// newOption returns a new option object.
func newOption(name string, values []string) *option {
	return &option{
		Name:   name,
		Values: values,
	}
}

func (o *option) SetValues(vs ...string) {
	o.Values = vs
}

func (o *option) AddValue(v string) {
	o.Values = append(o.Values, v)
}

func (o *option) MergeValues(vs ...string) {
	have := make(map[string]struct{})
	for _, v := range o.Values {
		have[v] = struct{}{}
	}

	for _, v := range vs {
		if _, exists := have[v]; exists {
			continue
		}
		o.AddValue(v)
	}
}
