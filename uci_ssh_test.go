package uci_test

import (
	"testing"

	"github.com/digineo/go-uci"
	"github.com/stretchr/testify/assert"
)

const USERNAME = "root"
const PASSWORD = ""
const HOST = "192.168.1.129"

func TestWrapperConnect(t *testing.T) {
	// ToDo: we should ship a test server!
	dut := uci.NewSshTree(USERNAME, PASSWORD, HOST)
	err := dut.Disonnect()
	if err != nil {
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
	dut := uci.NewSshTree(USERNAME, PASSWORD, HOST)

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
	dut := uci.NewSshTree(USERNAME, PASSWORD, HOST)

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
