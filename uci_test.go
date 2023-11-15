package uci

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func loadExpected(t *testing.T, name string) *config {
	t.Helper()

	f, err := os.Open(filepath.Join("testdata", name+".json"))
	if err != nil {
		t.Fatalf("cannot open %s.json: %v", name, err)
	}
	defer f.Close()

	expected := &config{}
	err = json.NewDecoder(f).Decode(&expected)
	if err != nil {
		t.Fatalf("errstr decoding json: %v", err)
	}

	// The JSON dump does not contain empty slices (they're marked with
	// "omitempty"), but the decoder creates them anyway. To get the tests
	// to pass, we need to eliminate nil slices (sections of config and
	// options of section) manually.
	if expected.Sections == nil {
		expected.Sections = []*section{}
	}
	for _, sec := range expected.Sections {
		if sec.Options == nil {
			sec.Options = []*option{}
		}
	}
	return expected
}

func TestLoadConfig(t *testing.T) {
	tt := []string{"system", "emptyfile", "emptysection", "luci", "ucitrack"}
	for i := range tt {
		name := tt[i]
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			r := NewTree("testdata")
			err := r.LoadConfig(name, false)
			assert.NoError(err)

			actual := r.(*tree).configs[name]

			if dump["json"] {
				assert.NoError(json.NewEncoder(os.Stderr).Encode(actual))
			}

			expected := loadExpected(t, name)
			assert.EqualValues(expected, actual)
		})
	}
}

func TestLoadConfig_nonExistent(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")
	err := r.LoadConfig("nonexistent", false)

	// os.IsNotExist fails on 1.16, https://golang.org/issue/44349
	assert.True(errors.Is(err, os.ErrNotExist))
}

func TestLoadConfig_forceReload(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")

	err := r.LoadConfig("system", false)
	assert.NoError(err)

	err = r.LoadConfig("system", false)
	assert.True(IsConfigAlreadyLoaded(err))

	err = r.LoadConfig("system", true)
	assert.NoError(err)
}

func TestLoadConfig_invalidFile(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")
	err := r.LoadConfig("invalid", false)
	assert.Error(err)

	var parseErr *ParseError
	assert.True(errors.As(err, &parseErr))
}

func TestWriteConfig(t *testing.T) {
	tt := []string{"system", "emptyfile", "emptysection", "luci", "ucitrack"}
	for i := range tt {
		name := tt[i]
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			r := NewTree("testdata")
			err := r.LoadConfig(name, false)
			assert.NoError(err)

			actual := r.(*tree).configs[name]
			var buf bytes.Buffer
			_, err = actual.WriteTo(&buf)
			assert.NoError(err)

			if dump["serialized"] {
				fmt.Fprint(os.Stderr, buf.String())
			}

			// TODO: validate content of buf
		})
	}
}

func TestGetSections(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	names, err := r.GetSections("system", "system")
	assert.NoError(err)
	assert.ElementsMatch(names, []string{"@system[0]"})

	names, err = r.GetSections("system", "timeserver")
	assert.NoError(err)
	assert.ElementsMatch(names, []string{"ntp"})

	names, err = r.GetSections("nonexistent", "foo")
	assert.True(errors.Is(err, os.ErrNotExist)) // fails as the underlying file fails to load.
	assert.Nil(names)

	names, err = r.GetSections("anonymous", "anon1")
	assert.NoError(err)
	assert.ElementsMatch(names, []string{"@anon1[0]"})

	names, err = r.GetSections("anonymous", "anon2")
	assert.NoError(err)
	assert.ElementsMatch(names, []string{"@anon2[0]", "@anon2[1]"})

	names, err = r.GetSections("anonymous", "anon3")
	assert.NoError(err)
	assert.ElementsMatch(names, []string{"@anon3[0]", "@anon3[1]", "@anon3[2]"})
}

func TestAddSection(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")

	assert.NoError(r.AddSection("nonexistent", "a", "section"))

	assert.NoError(r.AddSection("system", "foo", "foo"))
	assert.NoError(r.Set("system", "foo", "bar", "42"))
	values, exists := r.Get("system", "foo", "bar")
	assert.True(exists)
	assert.ElementsMatch(values, []string{"42"})

	assert.Error(r.AddSection("system", "foo", "notfoo"))
	assert.NoError(r.AddSection("system", "foo", "foo"))

	assert.NoError(r.AddSection("nonexistent", "a", "section"))
	assert.NoError(r.Set("nonexistent", "a", "section", "value"))
	values, exists = r.Get("nonexistent", "a", "section")
	assert.True(exists)
	assert.ElementsMatch(values, []string{"value"})
}

func TestDelSection(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")

	names, err := r.GetSections("system", "timeserver")
	assert.NoError(err)
	assert.ElementsMatch(names, []string{"ntp"})
	err = r.DelSection("system", "ntp")
	assert.NoError(err)

	names, err = r.GetSections("system", "timeserver")
	assert.NoError(err)
	assert.Len(names, 0)

	_, err = r.GetSections("nonexistent", "foo")
	assert.Error(err) // Todo: specify error type
	err = r.DelSection("nonexistent", "@foo[0]")
	assert.Error(err) // Todo: specify error type

	_, err = r.GetSections("nonexistent", "foo")
	assert.Error(err) // Todo: specify error type
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	values, exists := r.Get("system", "ntp", "server")
	assert.True(exists)
	assert.ElementsMatch(values, []string{
		"0.lede.pool.ntp.org",
		"1.lede.pool.ntp.org",
		"2.lede.pool.ntp.org",
		"3.lede.pool.ntp.org",
	})

	values, exists = r.Get("system", "@system[0]", "timezone")
	assert.True(exists)
	assert.ElementsMatch(values, []string{"UTC"})

	values, exists = r.Get("system", "nonexistent", "foo")
	assert.False(exists)
	assert.Nil(values)
}

func TestDel(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")

	values, _ := r.Get("system", "ntp", "enabled")
	assert.ElementsMatch(values, []string{"1"})
	err := r.Del("system", "ntp", "enabled")
	assert.NoError(err)
	values, _ = r.Get("system", "ntp", "enabled")
	assert.ElementsMatch(values, []string{})

	_, exists := r.Get("system", "nonexistent", "foo")
	assert.False(exists)
	err = r.Del("system", "nonexistent", "foo")
	assert.Error(err)
	eerr := &ErrSectionNotFound{}
	assert.True(errors.As(err, eerr)) // no such section
	_, exists = r.Get("system", "nonexistent", "foo")
	assert.False(exists)

	_, exists = r.Get("nonexistent", "foo", "bar")
	assert.False(exists)
	err = r.Del("nonexistent", "foo", "bar")
	assert.Error(err)
	assert.True(errors.Is(err, os.ErrNotExist)) // fails as the underlying file fails to load.
	_, exists = r.Get("nonexistent", "foo", "bar")
	assert.False(exists)

	// without prior loading
	err = r.Del("nonexistent2", "foo2", "bar2")
	assert.Error(err)
	assert.True(errors.Is(err, os.ErrNotExist)) // fails as the underlying file fails to load.
	_, exists = r.Get("nonexistent2", "foo2", "bar2")
	assert.False(exists)
}

func TestSet(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")

	assert.NoError(r.Set("system", "ntp", "enabled", "0"))
	values, exists := r.Get("system", "ntp", "enabled")
	assert.True(exists)
	assert.ElementsMatch(values, []string{"0"})

	values, exists = r.Get("system", "@system[0]", "hostname")
	assert.True(exists)
	assert.ElementsMatch(values, []string{"testhost"})

	assert.NoError(r.Set("system", "@system[0]", "hosttest"))

	assert.Error(r.Set("system", "nonexistent", "foo", "bar")) // Todo: specify error type
	values, exists = r.Get("system", "nonexistent", "foo")
	assert.False(exists)
	assert.Nil(values)

	assert.Error(r.Set("nonexistent", "foo", "bar", "42"))
	values, exists = r.Get("nonexistent", "foo", "bar")
	assert.False(exists)
	assert.Nil(values)
}

func TestListDelete(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	val, _ := r.Get("system", "ntp", "server")
	assert.NotEmpty(val)

	err := r.Del("system", "ntp", "server")
	assert.NoError(err)

	val, _ = r.Get("system", "ntp", "server")
	assert.Empty(val)
}

func TestGetLast_Success(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	val, ok := r.GetLast("system", "ntp", "server")
	assert.True(ok)

	assert.Equal(val, "3.lede.pool.ntp.org")
}

func TestGetLast_Failure(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	_, ok := r.GetLast("system", "ntp", "port")
	assert.False(ok)
}

func TestGetBool_False(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	val, ok := r.GetBool("wireless", "guest_radio0", "disabled")
	assert.True(ok)

	assert.False(val)
}

func TestGetBool_True(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	val, ok := r.GetBool("wireless", "guest_radio1", "disabled")
	assert.True(ok)

	assert.True(val)
}

func TestGetBool_Other(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	_, ok := r.GetBool("wireless", "guest_radio0", "mode")
	assert.False(ok)
}

func TestRevert(t *testing.T) {
	assert := assert.New(t)
	r := NewTree("testdata")
	tree := r.(*tree)

	assert.NoError(r.LoadConfig("system", false))
	assert.Len(tree.configs, 1)

	// revert all
	r.Revert()
	assert.Len(tree.configs, 0)

	assert.NoError(r.LoadConfig("system", false))
	assert.Len(tree.configs, 1)

	// taint tree
	assert.NoError(r.Set("system", "ntp", "foo", "42"))
	assert.True(tree.configs["system"].tainted)
	r.Revert("system")
	assert.Len(tree.configs, 0)
}

func TestCommit(t *testing.T) {
	origNewTmpFile := newTmpFile
	m := &mockTempFile{}
	newTmpFile = func(_, _ string) (tmpFile, error) { return m, nil }
	defer func() { newTmpFile = origNewTmpFile }()

	assert := assert.New(t)
	r := NewTree("testdata")

	// untainted save
	assert.NoError(r.Commit())

	// taint the tree
	assert.NoError(r.AddSection("cfgname", "secname", "sectype"))
	assert.NoError(r.Set("cfgname", "secname", "optname", "optvalue"))
	const content = "\nconfig sectype 'secname'\n\toption optname 'optvalue'\n\n"

	// try saving, but let it fail at different points
	reset := func(onwrite, onchmod, onsync, onrename error) {
		m.Buffer.Reset()
		m.ExpectedCalls = nil
		m.On("Close").Return(nil)
		m.On("Remove").Return(nil)
		m.On("Write", mock.AnythingOfType("[]uint8")).Return(onwrite)
		m.On("Chmod", os.FileMode(0644)).Return(onchmod)
		m.On("Sync").Return(onsync)
		m.On("Rename", "testdata/cfgname").Return(onrename)
	}

	reset(errors.New("fail write"), nil, nil, nil) //nolint:goerr113
	assert.EqualError(r.Commit(), "fail write")
	assert.Equal(0, m.Buffer.Len())

	reset(nil, errors.New("fail chmod"), nil, nil) //nolint:goerr113
	assert.EqualError(r.Commit(), "save: failed to set permissions: fail chmod")
	assert.EqualValues(content, m.Buffer.String())

	reset(nil, nil, errors.New("fail sync"), nil) //nolint:goerr113
	assert.EqualError(r.Commit(), "save: failed to sync: fail sync")

	reset(nil, nil, nil, errors.New("fail rename")) //nolint:goerr113
	assert.EqualError(r.Commit(), "save: failed to replace existing config: fail rename")

	reset(nil, nil, nil, nil)
	assert.NoError(r.Commit())
}

type mockTempFile struct {
	mock.Mock
	bytes.Buffer
}

func (m *mockTempFile) Write(p []byte) (int, error) {
	args := m.Called(p)
	if err := args.Error(0); err != nil {
		return 0, err // nolint:wrapcheck
	}
	n, _ := m.Buffer.Write(p)
	return n, nil
}

func (m *mockTempFile) Chmod(mode os.FileMode) error {
	args := m.Called(mode)
	return args.Error(0)
}

func (m *mockTempFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockTempFile) Remove() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockTempFile) Rename(newpath string) error {
	args := m.Called(newpath)
	return args.Error(0)
}

func (m *mockTempFile) Sync() error {
	args := m.Called()
	return args.Error(0)
}
