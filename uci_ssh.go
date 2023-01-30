package uci

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

var _ Tree = (*SshTree)(nil)

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

	out, err := session.StdoutPipe()
	if err != nil {
		return
	}

	in, err := session.StdinPipe()
	if err != nil {
		return
	}
	defer in.Close()

	// Start the request
	path := filepath.Join(DefaultTreePath, name)
	err = session.Start(fmt.Sprintf("scp -qf %q", path))
	if err != nil {
		return
	}
	err = sendAck(in)
	if err != nil {
		return
	}

	// Get file header
	file, size, err := parseGetHeader(out)
	if !strings.Contains(file, name) {
		return errors.New("Header error")
	}

	// Get data
	sendAck(in)
	buffer := make([]byte, size)
	read, err := out.Read(buffer)
	if read != size {
		return errors.New("some strange read error")
	}
	cfg, err := parse(name, string(buffer))
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

func sendAck(writer io.Writer) (err error) {
	var msg = []byte{0}
	n, err := writer.Write(msg)
	if err != nil {
		return err
	}
	if n < len(msg) {
		return errors.New("failed to write ack buffer")
	}
	return nil
}

func parseGetHeader(reader io.Reader) (file string, size int, err error) {
	buffer := make([]uint8, 1)
	_, err = reader.Read(buffer)
	if err != nil {
		return
	}

	if buffer[0] > 0 {
		bufferedReader := bufio.NewReader(reader)
		response, err := bufferedReader.ReadString('\n')
		parsed := strings.Split(response, " ")
		file = strings.Replace(parsed[2], "\n", "", 1)
		if err != nil {
			return "", 0, err
		}
		size, err = strconv.Atoi(parsed[1])
		if err != nil {
			return "", 0, err
		}
	}
	return
}
