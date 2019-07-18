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

func (m *mockTree) GetSections(config string, secType string) ([]string, bool) {
	args := m.Called(config, secType)
	return []string{args.String(0)}, args.Bool(1)
}

func (m *mockTree) Get(config, section, option string) ([]string, bool) {
	args := m.Called(config, section, option)
	return []string{args.String(0)}, args.Bool(1)
}

func (m *mockTree) Set(config, section, option string, values ...string) bool {
	args := m.Called(config, section, option, values)
	return args.Bool(0)
}

func (m *mockTree) Del(config, section, option string) {
	m.Called(config, section, option)
}

func (m *mockTree) AddSection(config, section, typ string) error {
	args := m.Called(config, section, typ)
	return args.Error(0)
}

func (m *mockTree) DelSection(config, section string) {
	m.Called(config, section)
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
	assert.Error(LoadConfig("bar", false))
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
	m.On("GetSections", "foo", "bar").Return("sec1", true)
	list, ok := GetSections("foo", "bar")
	assert.True(ok)
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

func TestConvenienceSet(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("Set", "foo", "bar", "valid", []string{"42"}).Return(true)
	m.On("Set", "foo", "bar", "invalid", []string(nil)).Return(false)
	assert.False(Set("foo", "bar", "invalid"))
	assert.True(Set("foo", "bar", "valid", "42"))
	m.AssertExpectations(t)
}

func TestConvenienceDel(t *testing.T) {
	m := defaultTree.(*mockTree)
	m.On("Del", "foo", "bar", "opt").Return()
	Del("foo", "bar", "opt")
	m.AssertExpectations(t)
}

func TestConvenienceAddSection(t *testing.T) {
	assert := assert.New(t)
	m := defaultTree.(*mockTree)
	m.On("AddSection", "foo", "bar", "system").Return(nil)
	m.On("AddSection", "foo", "bar", "interface").Return(errors.New("invalid"))
	assert.Error(AddSection("foo", "bar", "interface"))
	assert.NoError(AddSection("foo", "bar", "system"))
	m.AssertExpectations(t)
}

func TestConvenienceDelSection(t *testing.T) {
	m := defaultTree.(*mockTree)
	m.On("DelSection", "foo", "bar").Return()
	DelSection("foo", "bar")
	m.AssertExpectations(t)
}
