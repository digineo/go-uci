package uci

import (
	"errors"
	"fmt"
)

// parse tries to parse a named input string into a Config object.
func parse(name, input string) (*config, error) {
	_, ch := lex(name, input)

	cfg := newConfig(name)
	var sec *section
	var opt *option
	inList := false

	for it := range ch {
		switch it.typ {
		case itemPackage:
			// ignore
		case itemError:
			return nil, errors.New(it.val)
		case itemConfig:
			if sec != nil {
				cfg.Add(sec)
			}

			sec = newSection("", "")
			opt = nil
			inList = false

		case itemOption:
			if sec == nil {
				return nil, fmt.Errorf("missing config declaration, found option")
			}
			if opt != nil {
				sec.Add(opt)
			}
			opt = newOption("", nil)
			inList = false

		case itemList:
			if sec == nil {
				return nil, fmt.Errorf("missing config declaration, found option")
			}
			if opt == nil || opt.name != "" && !inList {
				opt = newOption("", nil)
				inList = true
			}

		case itemIdent:
			if opt != nil {
				opt.name = it.val
				sec.Add(opt)
			} else {
				sec.typ = it.val
			}

		case itemString:
			if opt != nil {
				opt.AddValue(it.val)
			} else {
				sec.name = it.val
				cfg.Add(sec)
			}
		}
	}

	return cfg, nil
}
