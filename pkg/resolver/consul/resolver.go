package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
	"google.golang.org/grpc/resolver"
	"log"
	"sync"
	"time"
)

type consulResolver struct {
	cc        resolver.ClientConn
	client    *api.Client
	dc        string
	key       string
	mdConvert instance.MetadataConvert
	done      chan struct{}
	doneOnce  sync.Once
	backoff   func(int) time.Duration
}

func (r *consulResolver) watch() {
	qo := &api.QueryOptions{
		Datacenter: r.dc,
		WaitTime:   time.Second * 10,
	}

	retryTimes := 0

	for {
		addrs, qm, err := r.client.Health().Service(r.key, "", true, qo)
		if err != nil {
			log.Printf("[error]failed to resolve addr, caused by %s", err)
			delay := r.backoff(retryTimes)
			retryTimes++
			time.Sleep(delay)
			continue
		}

		if r.hasClosed() {
			break
		}

		qo.WaitIndex = qm.LastIndex

		addresses := make([]resolver.Address, len(addrs))

		for i := range addrs {
			svc := addrs[i].Service
			addresses[i] = resolver.Address{
				Addr:       fmt.Sprintf("%s:%d", svc.Address, svc.Port),
				ServerName: svc.Service,
			}
			if r.mdConvert != nil && len(svc.Meta) > 0 {
				addresses[i].Metadata = r.mdConvert(svc.Meta)
			} else {
				addresses[i].Metadata = &svc.Meta
			}
		}

		r.cc.UpdateState(resolver.State{
			Addresses: addresses,
		})

		if r.hasClosed() {
			break
		}

		retryTimes = 0
	}
}

func (r *consulResolver) hasClosed() bool {
	select {
	case <-r.done:
		return true
	default:
	}
	return false
}

func (r *consulResolver) ResolveNow(opts resolver.ResolveNowOption) {}

func (r *consulResolver) Close() {
	r.doneOnce.Do(func() {
		close(r.done)
	})
}
