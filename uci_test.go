package uci

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/digineo/go-uci/uci"
	"github.com/stretchr/testify/assert"
)

func opt(name string, vs ...string) *uci.Option {
	return &uci.Option{Name: name, Values: vs}
}

func loadExpected(name string) *uci.Config {
	f, err := os.Open(filepath.Join("testdata", name+".json"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	expected := uci.Config{}
	err = json.NewDecoder(f).Decode(&expected)
	if err != nil {
		panic(err)
	}
	return &expected
}

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)

	for _, name := range []string{"emptyfile", "emptysection", "luci", "system", "ucitrack"} {
		t.Run(name, func(t *testing.T) {
			r := NewTree("testdata")
			err := r.LoadConfig(name)
			assert.NoError(err)

			actual := r.(*tree).configs[name]

			if os.Getenv("DUMP_JSON") == "1" {
				json.NewEncoder(os.Stderr).Encode(actual)
			}

			expected := loadExpected(name)
			assert.EqualValues(expected, actual)
		})
	}
}
