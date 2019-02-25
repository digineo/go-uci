// NOTE: the JSON code is mainly used to simplify tests, by generating
// (and manually verifying) an expected test value (testdata/*.json) when
// DUMP_JSON=1 is set, and reading it back when DUMP_JSON has any other
// value.

package uci

import (
	"encoding/json"
	"sort"
)

type jsonConfig struct {
	Name     string         `json:"name"`
	Sections []*jsonSection `json:"sections,omitempty"`
}

type jsonSection struct {
	Name    string        `json:"name,omitempty"`
	Type    string        `json:"type"`
	Options []*jsonOption `json:"options,omitempty"`
}

type jsonOption struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func (c *config) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonConfig{
		Name:     c.name,
		Sections: c.jsonSections(),
	})
}

func (c *config) UnmarshalJSON(data []byte) error {
	jc := jsonConfig{}
	if err := json.Unmarshal(data, &jc); err != nil {
		return err
	}

	cfg := newConfig(jc.Name)
	for _, jsec := range jc.Sections {
		sec := newSection(jsec.Type, jsec.Name)
		for _, jopt := range jsec.Options {
			sec.Add(&option{name: jopt.Name, values: jopt.Values})
		}
		cfg.Add(sec)
	}
	*c = *cfg
	return nil
}

func (c *config) jsonSections() []*jsonSection {
	names := make([]string, 0, len(c.idx))
	for name := range c.idx {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return c.idx[names[i]] < c.idx[names[j]]
	})

	secs := make([]*jsonSection, 0, len(c.idx))
	for _, name := range names {
		sec := c.sec[c.idx[name]]
		secs = append(secs, &jsonSection{
			Name:    sec.name,
			Type:    sec.typ,
			Options: sec.jsonOptions(),
		})
	}
	return secs
}

func (s *section) jsonOptions() []*jsonOption {
	names := make([]string, 0, len(s.idx))
	for name := range s.idx {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return s.idx[names[i]] < s.idx[names[j]]
	})

	opts := make([]*jsonOption, 0, len(s.idx))
	for _, name := range names {
		opt := s.opts[s.idx[name]]
		opts = append(opts, &jsonOption{
			Name:   opt.name,
			Values: opt.values,
		})
	}
	return opts
}
