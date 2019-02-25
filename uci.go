package uci

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

var defaultTree = NewTree("/etc/config")

// ErrConfigAlreadyLoaded is returned by LoadConfig, if the given config
// name is already present.
type ErrConfigAlreadyLoaded struct {
	name string
}

func (err ErrConfigAlreadyLoaded) Error() string {
	return fmt.Sprintf("%s already loaded", err.name)
}

// Tree defines the base directory for UCI config files. The default value
// on OpenWRT devices point to /etc/config, so that is what the default
// tree uses as well (you can access the default tree with the package level
// functions with the same signature as in this interface).
type Tree interface {
	// LoadConfig reads a config file into memory and returns nil. If the
	// config is already loaded, ErrConfigAlreadyLoaded is returned. Errors
	// reading the config file are returned verbatim.
	//
	// You don't need to explicitly call LoadConfig(): Accessing configs
	// (and their sections) via Get, Set, Add, Delete, DeleteAll will
	// load missing files automatically.
	LoadConfig(name string) error

	// Commit writes all changes back to the system.
	//
	// Note: this is not transaction safe. If, for whatever reason, the
	// writing of any file fails, the succeeding files are left untouched
	// while the preceeding files are not reverted.
	Commit() error

	// Revert undoes any changes. This clears the internal memory and does
	// not access the file system.
	Revert()

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

	// AddSection adds a new config section.
	AddSection(config, section, typ string)

	// DelSection remove a config section and its options.
	DelSection(config, section string)
}

type tree struct {
	dir     string
	configs map[string]*config

	sync.RWMutex
}

var _ Tree = (*tree)(nil)

// NewTree constructs new RootDir pointing to root.
func NewTree(root string) Tree {
	return &tree{dir: root}
}

// LoadConfig delegates to the default tree. See Tree for details.
func LoadConfig(name string) error { return defaultTree.LoadConfig(name) }

// Commit delegates to the default tree. See Tree for details.
func Commit() error { return defaultTree.Commit() }

// Revert delegates to the default tree. See Tree for details.
func Revert() { defaultTree.Revert() }

// Get delegates to the default tree. See Tree for details.
func Get(config, section, option string) ([]string, bool) {
	return defaultTree.Get(config, section, option)
}

// Set delegates to the default tree. See Tree for details.
func Set(config, section, option string, values ...string) bool {
	return defaultTree.Set(config, section, option, values...)
}

// Del delegates to the default tree. See Tree for details.
func Del(config, section, option string) { defaultTree.Del(config, section, option) }

// AddSection delegates to the default tree. See Tree for details.
func AddSection(config, section, typ string) { defaultTree.AddSection(config, section, typ) }

// DelSection delegates to the default tree. See Tree for details.
func DelSection(config, section string) { defaultTree.DelSection(config, section) }

func (t *tree) LoadConfig(name string) error {
	t.RLock()
	var exists bool
	if t.configs != nil {
		_, exists = t.configs[name]
	}
	t.RUnlock()
	if exists {
		return &ErrConfigAlreadyLoaded{name}
	}

	t.Lock()
	defer t.Unlock()

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

	return nil
}

func (t *tree) Revert() {
	t.Lock()
	t.configs = nil
	t.Unlock()
}

func (t *tree) Get(config, section, option string) ([]string, bool) {
	t.RLock()
	defer t.RUnlock()

	return nil, false
}

func (t *tree) Set(config, section, option string, values ...string) bool {
	t.Lock()
	defer t.Unlock()

	return false
}

func (t *tree) Del(config, section, option string) {
	t.Lock()
	defer t.Unlock()
}

func (t *tree) AddSection(config, section, typ string) {
	t.Lock()
	defer t.Unlock()
}

func (t *tree) DelSection(config, section string) {
	t.Lock()
	defer t.Unlock()
}
