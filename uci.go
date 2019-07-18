package uci

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Tree defines the base directory for UCI config files. The default value
// on OpenWRT devices point to /etc/config, so that is what the default
// tree uses as well (you can access the default tree with the package level
// functions with the same signature as in this interface).
type Tree interface {
	// LoadConfig reads a config file into memory and returns nil. If the
	// config is already loaded, and forceReload is false, an error of type
	// ErrConfigAlreadyLoaded is returned. Errors reading the config file
	// are returned verbatim.
	//
	// You don't need to explicitly call LoadConfig(): Accessing configs
	// (and their sections) via Get, Set, Add, Delete, DeleteAll will
	// load missing files automatically.
	LoadConfig(name string, forceReload bool) error

	// Commit writes all changes back to the system.
	//
	// Note: this is not transaction safe. If, for whatever reason, the
	// writing of any file fails, the succeeding files are left untouched
	// while the preceding files are not reverted.
	Commit() error

	// Revert undoes changes to the config files given as arguments. If
	// no argument is given, all changes are reverted. This clears the
	// internal memory and does not access the file system.
	Revert(configs ...string)

	// GetSections returns the names of all sections of a certain type
	// in a config, and a boolean indicating whether the config file exists.
	GetSections(config string, secType string) ([]string, bool)

	// Get retrieves (all) values for a fully qualified option, and a
	// boolean indicating whether the config file and the config section
	// within exists.
	Get(config, section, option string) ([]string, bool)

	// Set replaces the fully qualified option with the given values. It
	// returns whether the config file and section exists. For new files
	// and sections, you first need to initialize them with AddSection().
	Set(config, section, option string, values ...string) bool

	// Del removes a fully qualified option.
	Del(config, section, option string)

	// AddSection adds a new config section. If the section already exists,
	// and the types match (existing type and given type), nothing happens.
	// Otherwise an ErrSectionTypeMismatch is returned.
	AddSection(config, section, typ string) error

	// DelSection remove a config section and its options.
	DelSection(config, section string)
}

type tree struct {
	dir     string
	configs map[string]*config

	sync.Mutex
}

var _ Tree = (*tree)(nil)

// NewTree constructs new RootDir pointing to root.
func NewTree(root string) Tree {
	return &tree{dir: root}
}

func (t *tree) LoadConfig(name string, forceReload bool) error {
	t.Lock()
	defer t.Unlock()

	var exists bool
	if t.configs != nil {
		_, exists = t.configs[name]
	}
	if exists && !forceReload {
		return &ErrConfigAlreadyLoaded{name}
	}
	return t.loadConfig(name)
}

// loadConfig actually reads a config file. Its call must be guarded by
// locking the tree's mutex.
func (t *tree) loadConfig(name string) error {
	body, err := ioutil.ReadFile(filepath.Join(t.dir, name))
	if err != nil {
		return err
	}
	cfg, err := parse(name, string(body))
	if err != nil {
		return err
	}

	if t.configs == nil {
		t.configs = make(map[string]*config)
	}
	t.configs[name] = cfg
	return nil
}

func (t *tree) Commit() error {
	t.Lock()
	defer t.Unlock()

	for _, config := range t.configs {
		if !config.tainted {
			continue
		}
		err := t.saveConfig(config)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *tree) Revert(configs ...string) {
	t.Lock()
	if len(configs) == 0 {
		t.configs = nil
	}
	for _, config := range configs {
		delete(t.configs, config)
	}
	t.Unlock()
}

func (t *tree) GetSections(config string, secType string) ([]string, bool) {
	cfg, exists := t.ensureConfigLoaded(config)
	if !exists {
		return nil, false
	}

	names := []string{}
	for _, s := range cfg.Sections {
		if s.Type == secType {
			names = append(names, cfg.sectionName(s))
		}
	}

	return names, true
}

func (t *tree) Get(config, section, option string) ([]string, bool) {
	t.Lock()
	defer t.Unlock()

	if vals, ok := t.lookupValues(config, section, option); ok {
		return vals, true
	}

	if err := t.loadConfig(config); err != nil {
		return nil, false
	}
	return t.lookupValues(config, section, option)
}

func (t *tree) ensureConfigLoaded(config string) (*config, bool) {
	cfg, loaded := t.configs[config]
	if !loaded {
		if err := t.loadConfig(config); err != nil {
			return nil, false
		}
		cfg = t.configs[config]
	}
	return cfg, true
}

func (t *tree) lookupOption(config, section, option string) (*option, bool) {
	cfg, exists := t.configs[config]
	if !exists {
		return nil, false
	}
	sec := cfg.Get(section)
	if sec == nil {
		return nil, true
	}
	return sec.Get(option), true
}

func (t *tree) lookupValues(config, section, option string) ([]string, bool) {
	opt, ok := t.lookupOption(config, section, option)
	if !ok {
		return nil, false
	}
	if opt == nil {
		return nil, true
	}
	return opt.Values, true
}

func (t *tree) Set(config, section, option string, values ...string) bool {
	t.Lock()
	defer t.Unlock()

	cfg, ok := t.ensureConfigLoaded(config)
	if !ok {
		return false
	}
	sec := cfg.Get(section)
	if sec == nil {
		return false
	}

	if opt := sec.Get(option); opt != nil {
		opt.SetValues(values...)
	} else {
		sec.Add(newOption(option, values...))
	}
	cfg.tainted = true
	return true
}

func (t *tree) Del(config, section, option string) {
	t.Lock()
	defer t.Unlock()

	cfg, ok := t.ensureConfigLoaded(config)
	if !ok {
		// we want to delete option, but neither config, nor section,
		// nor config do exist. hence, we've reached our desired state
		return
	}

	sec := cfg.Get(section)
	if sec == nil {
		// same logic applies here
		return
	}

	if sec.Del(option) {
		cfg.tainted = true
	}
}

func (t *tree) AddSection(config, section, typ string) error {
	t.Lock()
	defer t.Unlock()

	cfg, ok := t.ensureConfigLoaded(config)
	if !ok {
		cfg = newConfig(config)
		cfg.tainted = true
		t.configs[config] = cfg
	}
	sec := cfg.Get(section)
	if sec == nil {
		cfg.Add(newSection(typ, section))
		cfg.tainted = true
		return nil
	}
	if sec.Type != typ {
		return ErrSectionTypeMismatch{config, section, sec.Type, typ}
	}
	return nil
}

func (t *tree) DelSection(config, section string) {
	t.Lock()
	defer t.Unlock()

	cfg, ok := t.ensureConfigLoaded(config)
	if !ok {
		return
	}
	cfg.Del(section)
	cfg.tainted = true
}

func (t *tree) saveConfig(c *config) error {
	// We need to create a tempfile in the tree's base directory, since
	// os.Rename fails when that directory and ioutil.Tempdir are on
	// different file systems (os.Rename being not much more than a shim
	// for syscall.Renameat).
	//
	// The full path for f will hence be "$root/.$rnd.$name", which
	// translates to something like "/etc/config/.42.network" on
	// OpenWRT devices.
	//
	// We rely a bit on the fact that UCI ignores dotfiles in /etc/config,
	// so this should not interfere with normal operations when we leave
	// incomplete files behind (for whatever reason).
	f, err := ioutil.TempFile(t.dir, ".*."+c.Name)
	if err != nil {
		return err
	}

	_, err = c.WriteTo(f)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}

	if err = f.Chmod(0644); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()

	if err = os.Rename(f.Name(), filepath.Join(t.dir, c.Name)); err != nil {
		return err
	}

	c.tainted = false
	return nil
}
