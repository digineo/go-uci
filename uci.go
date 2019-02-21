// Package uci implements a binding to OpenWRT's UCI (Unified Configuration
// Interface) files in pure Go.
//
// The typical use case is reading and modifying UCI config options:
//	import "github.com/digineo/go-uci"
//	uci.Get("network", "lan", "ifname") //=> []string{"eth0.1"}
//	uci.Set("network", "lan", "ipaddr", "192.168.7.1")
//	uci.Commit() // or uci.Revert()
//
// For more details head over to the OpenWRT wiki, or dive into UCI's C
// source code:
//  - https://openwrt.org/docs/guide-user/base-system/uci
//  - https://git.openwrt.org/?p=project/uci.git;a=summary
package uci

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

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
	name    string
	Index   int
	Options map[string]*Option
}

// Name returns a sections name. If it is unnamed, a synthetic name is
// constructed from the type and index (e.g. "@system[0]").
func (s *Section) Name() string {
	if s.name == "" {
		return fmt.Sprintf("@%s[%d]", s.Type, s.Index)
	}
	return s.name
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

func (err *ParseError) Error() string {
	return fmt.Sprintf("cannot parse %s:%d: %s", err.name, err.line, err.msg)
}

func (err *ParseError) withMessage(format string, v ...interface{}) error {
	err.msg = fmt.Sprintf(format, v...)
	return err
}

func loadConfig(name string, r io.Reader) (*Config, error) {
	var (
		cfg  = &Config{Name: name}    // return value
		sec  *Section                 // current section
		perr = ParseError{name: name} // error template
	)

	s := bufio.NewScanner(r)
	for ; s.Scan(); perr.line++ {
		switch line := strings.TrimSpace(s.Text()); {
		case strings.HasPrefix(line, "config"):
			if sec != nil {
				cfg.Sections = append(cfg.Sections, sec)
			}
			typ, name, err := parseSection(line, &perr)
			if err != nil {
				return nil, err
			}
			sec = &Section{
				Type:    typ,
				name:    name,
				Index:   len(cfg.Sections),
				Options: make(map[string]*Option),
			}

		case strings.HasPrefix(line, "option"):
			fallthrough
		case strings.HasPrefix(line, "list"):
			if sec == nil {
				return nil, perr.withMessage("unexpected option without section")
			}
			name, value, err := parseOption(line, &perr)
			if err != nil {
				return nil, err
			}
			if opt, exists := sec.Options[name]; exists {
				opt.Values = append(opt.Values, value)
			} else {
				sec.Options[name] = &Option{
					Name:   name,
					Values: []string{value},
				}
			}
		default:
			continue
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	if sec != nil {
		cfg.Sections = append(cfg.Sections, sec)
	}
	return cfg, nil
}

func parseSection(line string, perr *ParseError) (typ, name string, err error) {
	f := strings.Fields(line)
	l := len(f)
	if l < 2 || l > 3 {
		err = perr.withMessage("expected 2-3 fields, got %v", f)
		return
	}
	if f[0] != "config" {
		err = perr.withMessage(`expected "config", found %q`, f[0])
		return
	}
	if l == 2 {
		return f[1], "", nil
	}
	return f[1], unquote(f[2]), nil
}

func parseOption(line string, perr *ParseError) (name, value string, err error) {
	var havePrefix, haveName bool
	var buf bytes.Buffer
	for _, r := range line {
		if unicode.IsSpace(r) {
			switch {
			case !havePrefix:
				havePrefix = true
				prefix := buf.String()
				if prefix != "option" && prefix != "list" {
					err = perr.withMessage(`expected "option" or "list", found %q`, prefix)
					return
				}
				buf.Reset()
				continue
			case !haveName:
				haveName = true
				name = buf.String()
				buf.Reset()
				continue
			}
		}
		if r == '\'' {
			if !havePrefix {
				err = perr.withMessage(`unexpected ' while parsing prefix`)
				return
			}
			if !haveName {
				err = perr.withMessage(`unexpected ' while parsing name`)
				return
			}
		}
		buf.WriteRune(r)
	}
	value = unquote(buf.String())
	return
}

func unquote(s string) string {
	if l := len(s); l > 2 &&
		(s[0] == '\'' && s[l-1] == '\'' || s[0] == '"' && s[l-1] == '"') {
		return s[1 : l-1]
	}
	return s
}
