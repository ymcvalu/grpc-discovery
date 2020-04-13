package config

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

const (
	TimeoutKey = "timeout"
)

type Config struct {
	Endpoints []string
	Username  string
	Password  string
	Timeout   time.Duration
}

var (
	ErrInvalidAuthority = errors.New("invalid authority")
)

// username:password@etcd1,etcd2,etcd3?timeout=3s
func Parse(authority string) (*Config, error) {
	ps := strings.SplitN(authority, "@", 2)
	cfg := &Config{}
	switch len(ps) {
	case 2:
		auth := strings.Split(ps[0], ":")
		if len(auth) != 2 {
			return nil, ErrInvalidAuthority
		}
		cfg.Username, cfg.Password = auth[0], auth[1]
		ps[0] = ps[1]
		fallthrough
	default:
		ps := strings.Split(ps[0], "?")
		if len(ps) > 2 {
			return nil, ErrInvalidAuthority
		}

		cfg.Endpoints = strings.Split(ps[0], ",")
		if len(ps) > 1 {
			vals, err := url.ParseQuery(ps[1])
			if err != nil {
				return nil, ErrInvalidAuthority
			}
			dur := vals.Get(TimeoutKey)
			if len(dur) > 0 {
				dur, err := time.ParseDuration(dur)
				if err != nil {
					return nil, ErrInvalidAuthority
				}
				cfg.Timeout = dur
			}
		}
	}

	return cfg, nil
}
