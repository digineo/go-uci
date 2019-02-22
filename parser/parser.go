package parser

import (
	"errors"
	"fmt"

	"github.com/digineo/go-uci/uci"
)

// Parse tries to parse a named input string into a Config object.
func Parse(name, input string) (*uci.Config, error) {
	_, ch := lex(name, input)

	cfg := &uci.Config{Name: name}
	var sec *uci.Section
	var opt *uci.Option
	inList := false

	for it := range ch {
		switch it.typ {
		case itemPackage:
			// ignore
		case itemError:
			return nil, errors.New(it.val)
		case itemConfig:
			sec = &uci.Section{
				Index:   len(cfg.Sections),
				Options: make(map[string]*uci.Option),
			}
			cfg.Sections = append(cfg.Sections, sec)
			opt = nil
			inList = false

		case itemOption:
			if sec == nil {
				return nil, fmt.Errorf("missing config declaration, found option")
			}
			opt = &uci.Option{} // cannot append yet, need name first
			inList = false

		case itemList:
			if sec == nil {
				return nil, fmt.Errorf("missing config declaration, found option")
			}
			if opt == nil || opt.Name != "" && !inList {
				opt = &uci.Option{}
				inList = true
			}

		case itemIdent:
			if opt != nil {
				opt.Name = it.val
				sec.Options[it.val] = opt
			} else {
				sec.Type = it.val
			}

		case itemString:
			if opt != nil {
				opt.Values = append(opt.Values, it.val)
			} else {
				sec.Name = it.val
			}
		}
	}

	return cfg, nil
}
