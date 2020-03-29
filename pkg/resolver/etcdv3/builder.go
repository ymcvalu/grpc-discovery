package etcdv3

import (
	"github.com/ymcvalu/grpc-discovery/pkg/backoff"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
	"path"
	"time"
)

func WithPrefix(prefix string) Option {
	return func(opts *Options) {
		opts.prefix = prefix
	}
}

func WithMdConvert(mc instance.MetadataConvert) Option {
	return func(opts *Options) {
		opts.mdConvert = mc
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
	config    clientv3.Config
	prefix    string
	maxDelay  time.Duration
}

func Init(cfg clientv3.Config, opts ...Option) resolver.Builder {
	builder := new(etcdBuilder)
	builder.opts = new(Options)
	for _, opt := range opts {
		opt(builder.opts)
	}

	if builder.opts.prefix == "" {
		builder.opts.prefix = "/grpc-discovery"
	}

	if builder.opts.maxDelay <= 0 {
		builder.opts.maxDelay = time.Second * 5
	}

	builder.opts.config = cfg
	builder.register()
	return builder
}

type etcdBuilder struct {
	opts *Options
}

func (b *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	key := path.Join(b.opts.prefix, target.Endpoint)
	r := &etcdResolver{
		cc:        cc,
		key:       key,
		done:      make(chan struct{}),
		mdConvert: b.opts.mdConvert,
		backoff:   backoff.New(b.opts.maxDelay).Backoff,
	}

	client, err := clientv3.New(b.opts.config)
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

func (b *etcdBuilder) register() {
	resolver.Register(b)
}
