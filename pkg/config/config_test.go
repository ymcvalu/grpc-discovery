package config

import (
	"reflect"
	"testing"
	"time"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		authority string
		succ      bool
		cfg       *Config
	}{
		{
			authority: "test:123456@127.0.0.1:2379,127.0.0.2:2379,127.0.0.3:2379?timeout=5s",
			succ:      true,
			cfg: &Config{
				Username: "test",
				Password: "123456",
				Endpoints: []string{
					"127.0.0.1:2379",
					"127.0.0.2:2379",
					"127.0.0.3:2379",
				},
				Timeout: 5 * time.Second,
			},
		},
		{
			authority: "127.0.0.1:2379,127.0.0.2:2379,127.0.0.3:2379",
			cfg: &Config{
				Endpoints: []string{
					"127.0.0.1:2379",
					"127.0.0.2:2379",
					"127.0.0.3:2379",
				},
			},
		},
	}
	for _, c := range cases {
		cfg, err := Parse(c.authority)
		if err != nil && c.succ {
			t.Errorf("failed to parse %s", c.authority)
		} else if err == nil {
			if !reflect.DeepEqual(cfg, c.cfg) {
				t.Errorf("has error when parsing %s %+v %+v", c.authority, *cfg, *c.cfg)
			}
		}
	}
}
