package uci

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
)

type SshTree struct {
	host   string
	config *ssh.ClientConfig
	client *ssh.Client
	tree
	sync.Mutex
}

func NewSshTree(config *ssh.ClientConfig, host string) (t *SshTree, err error) {
	t = &SshTree{
		config: config,
		host:   host,
	}
	t.client, err = ssh.Dial("tcp", t.host, t.config)
	return
}

func (t *SshTree) Disonnect() (err error) {
	return t.client.Close()
}

// Overwrite the default loadconfig!
func (t *SshTree) loadConfig(name string) (err error) {
	session, err := t.client.NewSession()
	defer session.Close()
	session.Stderr = os.Stderr
	var body bytes.Buffer
	session.Stdout = &body
	path := filepath.Join(DefaultTreePath, name)
	err = session.Run("cat " + path)
	if err != nil {
		log.Fatalln("Unable to run command: " + err.Error())
	}

	cfg, err := parse(name, body.String())
	if err != nil {
		return err
	}

	if t.configs == nil {
		t.configs = make(map[string]*config)
	}
	t.configs[name] = cfg
	return nil
}

func (t *SshTree) LoadConfig(name string, forceReload bool) error {
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

func (t *SshTree) Commit() error {
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

func (t *SshTree) saveConfig(c *config) (err error) {
	f, err := os.CreateTemp("", "tmpfile-")
	defer f.Close()
	defer os.Remove(f.Name())

	if err != nil {
		return err
	}

	_, err = c.WriteTo(f)
	if err != nil {
		return err
	}

	session, err := t.client.NewSession()
	defer session.Close()
	destinationPath := filepath.Join(DefaultTreePath, c.Name)

	err = scp.CopyPath(f.Name(), destinationPath, session)

	c.tainted = false
	return
}
