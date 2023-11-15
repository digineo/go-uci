package uci

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockTree implements the Tree interface and an instance replaces the
// global defaultTree (see TestMain).
type mockTree struct {
	mock.Mock
}

func (m *mockTree) LoadConfig(name string, forceReload bool) error {
	args := m.Called(name, forceReload)
	return args.Error(0)
}

func (m *mockTree) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockTree) Revert(configs ...string) {
	m.Called(configs)
}

func (m *mockTree) GetSections(config string, secType string) ([]string, error) {
	args := m.Called(config, secType)
	return []string{args.String(0)}, args.Error(1)
}

func (m *mockTree) Get(config, section, option string) ([]string, bool) {
	args := m.Called(config, section, option)
	return []string{args.String(0)}, args.Bool(1)
}

func (m *mockTree) GetLast(config, section, option string) (string, bool) {
	args := m.Called(config, section, option)
	return args.String(0), args.Bool(1)
}

func (m *mockTree) GetBool(config, section, option string) (bool, bool) {
	args := m.Called(config, section, option)
	return args.Bool(0), args.Bool(1)
}

func (m *mockTree) SetType(config, section, option string, typ OptionType, values ...string) error {
	args := m.Called(config, section, option, typ, values)
	return args.Error(0)
}

func (m *mockTree) Del(config, section, option string) error {
	m.Called(config, section, option)
	return nil
}

func (m *mockTree) AddSection(config, section, typ string) error {
	args := m.Called(config, section, typ)
	return args.Error(0)
}

func (m *mockTree) DelSection(config, section string) error {
	m.Called(config, section)
	return nil
}

func TestMain(m *testing.M) {
	defaultTree = &mockTree{}
	os.Exit(m.Run())
}

func TestConvenienceLoadConfig(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("LoadConfig", "foo", true).Return(nil)
	m.On("LoadConfig", "bar", false).Return(io.ErrUnexpectedEOF)
	assert.NoError(LoadConfig("foo", true))
	err := LoadConfig("bar", false)
	assert.Error(err, io.ErrUnexpectedEOF)
	m.AssertExpectations(t)
}

func TestConvenienceCommit(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("Commit").Return(nil)
	assert.NoError(Commit())
	m.AssertExpectations(t)
}

func TestConvenienceRevert(t *testing.T) {
	m := defaultTree.(*mockTree)
	m.On("Revert", []string{"foo", "bar"}).Return()
	Revert("foo", "bar")
	m.AssertExpectations(t)
}

func TestConvenienceGetSections(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("GetSections", "foo", "bar").Return("sec1", nil)
	list, err := GetSections("foo", "bar")
	assert.NoError(err)
	assert.EqualValues([]string{"sec1"}, list)
	m.AssertExpectations(t)
}

func TestConvenienceGet(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("Get", "foo", "bar", "opt").Return("ok", true)
	list, ok := Get("foo", "bar", "opt")
	assert.True(ok)
	assert.EqualValues([]string{"ok"}, list)
	m.AssertExpectations(t)
}

func TestConvenienceGetLast(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("GetLast", "foo", "bar", "opt").Return("ok", true)
	list, ok := GetLast("foo", "bar", "opt")
	assert.True(ok)
	assert.EqualValues("ok", list)
	m.AssertExpectations(t)
}

func TestConvenienceGetBool(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("GetBool", "foo", "bar", "opt").Return(true, true)
	value, ok := GetBool("foo", "bar", "opt")
	assert.True(ok)
	assert.True(value)
	m.AssertExpectations(t)
}

func TestConvenienceDel(t *testing.T) {
	m := defaultTree.(*mockTree)
	m.On("Del", "foo", "bar", "opt").Return()
	err := Del("foo", "bar", "opt")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestConvenienceAddSection(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	addSectionErr := errors.New("invalid")
	m.On("AddSection", "foo", "bar", "system").Return(nil)
	m.On("AddSection", "foo", "bar", "interface").Return(addSectionErr) //nolint:goerr113
	err := AddSection("foo", "bar", "interface")
	assert.Error(err)
	assert.EqualError(err, addSectionErr.Error())
	assert.NoError(AddSection("foo", "bar", "system"))
	m.AssertExpectations(t)
}

func TestConvenienceDelSection(t *testing.T) {
	m := defaultTree.(*mockTree)
	m.On("DelSection", "foo", "bar").Return()
	err := DelSection("foo", "bar")
	assert.NoError(t, err)
	m.AssertExpectations(t)
}
