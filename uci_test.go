package uci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
		t.Fatalf("error decoding json: %v", err)
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
	assert := assert.New(t)

	for _, name := range []string{"system", "emptyfile", "emptysection", "luci", "ucitrack"} {
		name := name
		t.Run(name, func(t *testing.T) {
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

func TestWriteConfig(t *testing.T) {
	assert := assert.New(t)

	tt := []string{"system", "emptyfile", "emptysection", "luci", "ucitrack"}
	for i := range tt {
		name := tt[i]
		t.Run(name, func(t *testing.T) {
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

	names, exists := r.GetSections("system", "system")
	assert.True(exists)
	assert.ElementsMatch(names, []string{"@system[1]"})

	names, exists = r.GetSections("system", "timeserver")
	assert.True(exists)
	assert.ElementsMatch(names, []string{"ntp"})
}

func TestSingleDelete(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	val, _ := r.Get("system", "ntp", "enable_server")
	assert.NotEmpty(val)

	r.Del("system", "ntp", "enable_server")

	val, _ = r.Get("system", "ntp", "enable_server")
	assert.Empty(val)
}

func TestListDelete(t *testing.T) {
	assert := assert.New(t)

	r := NewTree("testdata")

	val, _ := r.Get("system", "ntp", "server")
	assert.NotEmpty(val)

	r.Del("system", "ntp", "server")

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
