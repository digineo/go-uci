package uci

import (
	"encoding/json"
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
	return expected
}

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)

	for _, name := range []string{"system", "emptyfile", "emptysection", "luci", "system", "ucitrack"} {
		t.Run(name, func(t *testing.T) {
			r := NewTree("testdata")
			err := r.LoadConfig(name)
			assert.NoError(err)

			actual := r.(*tree).configs[name]

			if os.Getenv("DUMP_JSON") == "1" {
				json.NewEncoder(os.Stderr).Encode(actual)
			}

			expected := loadExpected(t, name)
			assert.EqualValues(expected, actual)
		})
	}
}
