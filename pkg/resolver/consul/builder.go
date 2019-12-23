package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/ymcvalu/grpc-discovery/pkg/backoff"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
	"google.golang.org/grpc/resolver"
	"time"
)

func WithMdConvert(mc instance.MetadataConvert) Option {
	return func(opts *Options) {
		opts.mdConvert = mc
	}
}

func WithDataCenter(dc string) Option {
	return func(opts *Options) {
		opts.dc = dc
	}
}

func WithMaxDelay(maxDelay time.Duration) Option {
	return func(opts *Options) {
		opts.maxDelay = maxDelay
	}
}

type Option func(opts *Options)

type Options struct {
	mdConvert instance.MetadataConvert
	dc        string
	addr      string
	maxDelay  time.Duration
}

// consulBuilder creates a consulResolver that will be used to watch name resolution updates.
type consulBuilder struct {
	opts Options
}

func (b *consulBuilder) Scheme() string {
	return "consul"
}

func (b *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	client, err := api.NewClient(&api.Config{
		Datacenter: b.opts.dc,
		Address:    b.opts.addr,
	})

	if err != nil {
		return nil, err
	}

	r := &consulResolver{
		cc:        cc,
		client:    client,
		dc:        b.opts.dc,
		mdConvert: b.opts.mdConvert,
		key:       target.Endpoint,
		done:      make(chan struct{}),
		backoff:   backoff.New(b.opts.maxDelay).Backoff,
	}

	go r.watch()

	return r, nil
}

func Init(addr string, opts ...Option) {
	_opts := Options{}

	for _, opt := range opts {
		opt(&_opts)
	}

	if _opts.dc == "" {
		_opts.dc = "dc1"
	}

	if _opts.maxDelay <= 0 {
		_opts.maxDelay = time.Second * 5
	}
	_opts.addr = addr

	resolver.Register(&consulBuilder{
		opts: _opts,
	})
}
