package uci

// defaultTree is a convenient accessor to the UCI default location
// at /etc/config.
var defaultTree = NewTree("/etc/config")

// LoadConfig delegates to the default tree. See Tree for details.
func LoadConfig(name string, forceReload bool) error {
	return defaultTree.LoadConfig(name, forceReload)
}

// Commit delegates to the default tree. See Tree for details.
func Commit() error {
	return defaultTree.Commit()
}

// Revert delegates to the default tree. See Tree for details.
func Revert() {
	defaultTree.Revert()
}

// Get delegates to the default tree. See Tree for details.
func Get(config, section, option string) ([]string, bool) {
	return defaultTree.Get(config, section, option)
}

// Set delegates to the default tree. See Tree for details.
func Set(config, section, option string, values ...string) bool {
	return defaultTree.Set(config, section, option, values...)
}

// Del delegates to the default tree. See Tree for details.
func Del(config, section, option string) {
	defaultTree.Del(config, section, option)
}

// AddSection delegates to the default tree. See Tree for details.
func AddSection(config, section, typ string) error {
	return defaultTree.AddSection(config, section, typ)
}

// DelSection delegates to the default tree. See Tree for details.
func DelSection(config, section string) {
	defaultTree.DelSection(config, section)
}
