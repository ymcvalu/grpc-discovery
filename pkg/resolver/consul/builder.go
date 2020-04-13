package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/ymcvalu/grpc-discovery/pkg/backoff"
	"github.com/ymcvalu/grpc-discovery/pkg/config"
	"google.golang.org/grpc/resolver"
	"time"
)

var (
	DC       = "dc1"
	MaxDelay = 5 * time.Second
)

func init() {
	resolver.Register(&consulBuilder{})
}

// consulBuilder creates a consulResolver that will be used to watch name resolution updates.
type consulBuilder struct{}

func (b *consulBuilder) Scheme() string {
	return "consul"
}

func (b *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	cfg, err := config.Parse(target.Authority)
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(&api.Config{
		Datacenter: DC,
		Address:    cfg.Endpoints[0],
		HttpAuth: &api.HttpBasicAuth{
			Username: cfg.Username,
			Password: cfg.Password,
		},
	})

	if err != nil {
		return nil, err
	}

	r := &consulResolver{
		cc:      cc,
		client:  client,
		dc:      DC,
		key:     target.Endpoint,
		done:    make(chan struct{}),
		backoff: backoff.New(MaxDelay).Backoff,
	}

	go r.watch()

	return r, nil
}
