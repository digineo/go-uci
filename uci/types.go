package uci

// Config represents a file in UCI. It consists of sections.
type Config struct {
	Name     string
	Sections []*Section
}

// A Section represents a group of options in UCI. It may be named or
// unnamed. In the latter case, its synthetic name is constructed from
// the section type and index (e.g. "@system[0]").
//
// Sections consist of Options.
type Section struct {
	Type    string
	Name    string
	Index   int
	Options map[string]*Option
}

// An Option is the key to one or more values. Multiple values indicate
// a list option.
type Option struct {
	Name   string
	Values []string
}

// A ParseError is emitted when parsing a UCI config file encounters
// unexpected tokens.
type ParseError struct {
	name string
	line int
	msg  string
}
