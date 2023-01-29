package uci_test

import (
	"testing"
	"time"

	"github.com/digineo/go-uci"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

const USERNAME = "root"
const PASSWORD = ""
const HOST = "192.168.1.129:22"

func getconfig() *ssh.ClientConfig {
	retval := &ssh.ClientConfig{
		User: USERNAME,
		Auth: []ssh.AuthMethod{
			ssh.Password(PASSWORD),
		},
		// Non-production only
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	return retval
}

func TestWrapperConnect(t *testing.T) {
	conf := getconfig()
	// ToDo: we should ship a test server!
	dut, _ := uci.NewSshTree(conf, HOST)
	if err := dut.Disonnect(); err != nil {
		t.Error(err.Error())
	}
}

func TestInterface(t *testing.T) {
	assert := assert.New(t)
	v := &uci.SshTree{}
	var i interface{} = v
	_, ok := i.(uci.Tree)
	assert.Equal(ok, true, "Interface not implemented")
}

func TestSshLoadConfig(t *testing.T) {
	assert := assert.New(t)

	dut, _ := uci.NewSshTree(getconfig(), HOST)

	defer dut.Disonnect()
	err := dut.LoadConfig("system", false)
	if err != nil {
		t.Error(err.Error())
	}
	if values, ok := dut.Get("system", "@system[0]", "hostname"); ok {
		assert.Equal(values[0], "OpenWrt")
	}
}

func TestSshSaveConfig(t *testing.T) {
	assert := assert.New(t)
	dut, _ := uci.NewSshTree(getconfig(), HOST)

	dut.LoadConfig("system", false)
	assert.NoError(dut.AddSection("system", "foo", "foo"))
	assert.True(dut.Set("system", "foo", "bar", "42"))
	_, exists := dut.Get("system", "foo", "bar")
	assert.True(exists)
	err := dut.Commit()
	assert.Nil(err)

	dut.LoadConfig("system", true)
	values, exists := dut.Get("system", "foo", "bar")
	assert.True(exists)
	assert.ElementsMatch(values, []string{
		"42",
	})
	dut.Del("system", "foo", "bar")

	val, _ := dut.Get("system", "foo", "bar")
	assert.Empty(val)
}
