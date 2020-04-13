package etcdv3

import (
	"github.com/ymcvalu/grpc-discovery/pkg/backoff"
	"github.com/ymcvalu/grpc-discovery/pkg/config"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"path"
	"time"
)

var (
	Prefix   = "/grpc-discovery"
	MaxDelay = 5 * time.Second
)

func init() {
	resolver.Register(&etcdBuilder{})
}

type etcdBuilder struct{}

func (b *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	key := path.Join(Prefix, target.Endpoint)
	r := &etcdResolver{
		cc:      cc,
		key:     key,
		done:    make(chan struct{}),
		backoff: backoff.New(MaxDelay).Backoff,
	}

	// etcd://user:passwd@127.0.0.1:2379,127.0.0.2:2379,127.0.0.3:2379?timeout=3/dev/app
	cfg, err := config.Parse(target.Authority)
	if err != nil {
		return nil, err
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: cfg.Timeout,
	})
	if err != nil {
		return nil, err
	}

	r.client = client

	go r.watch()

	return r, nil
}

func (b *etcdBuilder) Scheme() string {
	return "etcd"
}
