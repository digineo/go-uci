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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/digineo/go-uci/parser"
	"github.com/digineo/go-uci/uci"
)

var defaultRoot = NewRootDir("/etc/config")

// ErrConfigAlreadyLoaded is returned by LoadConfig, if the given config
// name is already present.
type ErrConfigAlreadyLoaded struct {
	name string
}

func (err ErrConfigAlreadyLoaded) Error() string {
	return fmt.Sprintf("%s already loaded", err.name)
}

// RootDir defines the base directory for UCI config files. The default
// on an OpenWRT device points to `/etc/config`.
type RootDir interface {
	// LoadConfig reads a config file into memory and returns nil. If the
	// config is already loaded, ErrConfigAlreadyLoaded is returned. Errors
	// reading the config file are returned verbatim.
	//
	// You don't need to explicitly call LoadConfig(): Accessing configs
	// (and their sections) via Get, Set, Add, Delete, DeleteAll will
	// load missing files automatically.
	LoadConfig(name string) error

	// Commit writes all changes back to the system.

	// Note: this is not transaction safe. If, for whatever reason, the
	// writing of any file fails, the succeeding files are left untouched
	// while the preceeding files are not reverted.
	Commit() error

	// Revert undoes any changes. This clears the internal memory and does
	// not access the file system.
	Revert()
}

type rootDir struct {
	root    string
	configs map[string]*uci.Config

	sync.RWMutex
}

var _ RootDir = (*rootDir)(nil)

// NewRootDir constructs new RootDir pointing to root.
func NewRootDir(root string) RootDir {
	return &rootDir{root: root}
}

// LoadConfig delegates to the default root. See RootDir for details.
func LoadConfig(name string) error { return defaultRoot.LoadConfig(name) }

// Commit delegates to the default root. See RootDir for details.
func Commit() error { return defaultRoot.Commit() }

// Revert delegates to the default root. See RootDir for details.
func Revert() { defaultRoot.Revert() }

func (root *rootDir) LoadConfig(name string) error {
	root.RLock()
	var exists bool
	if root.configs != nil {
		_, exists = root.configs[name]
	}
	root.RUnlock()
	if exists {
		return &ErrConfigAlreadyLoaded{name}
	}

	root.Lock()
	defer root.Unlock()

	body, err := ioutil.ReadFile(filepath.Join(root.root, name))
	if err != nil {
		return err
	}
	cfg, err := parser.Parse(name, string(body))
	if err != nil {
		return err
	}

	if root.configs == nil {
		root.configs = make(map[string]*uci.Config)
	}
	root.configs[name] = cfg
	return nil
}

func (root *rootDir) Commit() error {
	root.Lock()
	defer root.Unlock()

	return nil
}

func (root *rootDir) Revert() {
	root.Lock()
	root.configs = nil
	root.Unlock()
}
