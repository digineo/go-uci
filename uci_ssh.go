package uci

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type SshTree struct {
	host   string
	config ssh.ClientConfig
	client *ssh.Client
	tree
	sync.Mutex
}

func NewSshTree(username string, password string, host string) (t *SshTree) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// Non-production only
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	t = &SshTree{
		config: *config,
		host:   host,
	}
	var err error
	t.client, err = ssh.Dial("tcp", t.host+":"+"22", &t.config)
	if err != nil {
		panic(err.Error())
	}
	return
}

func (t *SshTree) Disonnect() (err error) {

	err = t.client.Close()
	return err
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
	var body bytes.Buffer
	_, err = c.WriteTo(&body)
	path := filepath.Join(DefaultTreePath, c.Name)
	cmd := "echo '" + body.String() + "' >> " + path

	session, err := t.client.NewSession()
	defer session.Close()
	session.Stderr = os.Stderr
	var reply bytes.Buffer
	session.Stdout = &reply

	err = session.Run(cmd)

	c.tainted = false
	return
}
