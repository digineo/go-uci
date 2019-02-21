package uci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)

	r := NewRootDir("testdata")
	err := r.LoadConfig("system")
	assert.NoError(err)

	assert.EqualValues(&Config{
		Name: "system",
		Sections: []*Section{
			&Section{
				Type:  "system",
				name:  "",
				Index: 0,
				Options: map[string]*Option{
					"timezone":     &Option{"timezone", []string{"UTC"}},
					"ttylogin":     &Option{"ttylogin", []string{"0"}},
					"log_size":     &Option{"log_size", []string{"64"}},
					"urandom_seed": &Option{"urandom_seed", []string{"0"}},
					"hostname":     &Option{"hostname", []string{"testhost"}},
				},
			},
			&Section{
				Type:  "timeserver",
				name:  "ntp",
				Index: 1,
				Options: map[string]*Option{
					"enabled":       &Option{"enabled", []string{"1"}},
					"enable_server": &Option{"enable_server", []string{"0"}},
					"server":        &Option{"server", []string{"0.lede.pool.ntp.org", "1.lede.pool.ntp.org", "2.lede.pool.ntp.org", "3.lede.pool.ntp.org"}},
				},
			},
			&Section{
				Type:  "gpio_switch",
				name:  "poe_passthrough",
				Index: 2,
				Options: map[string]*Option{
					"name":     &Option{"name", []string{"PoE Passthrough"}},
					"gpio_pin": &Option{"gpio_pin", []string{"0"}},
					"value":    &Option{"value", []string{"0"}},
				},
			},
		},
	}, r.(*rootDir).configs["system"])
}
