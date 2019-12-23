package etcdv3

import (
	"context"
	"fmt"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"

	"log"
	"sync"
	"time"
)

type etcdResolver struct {
	done      chan struct{}
	doneOnce  sync.Once
	cc        resolver.ClientConn
	client    *clientv3.Client
	key       string
	mdConvert instance.MetadataConvert
	backoff   func(int) time.Duration
}

func (r *etcdResolver) ResolveNow(resolver.ResolveNowOption) {
}

func (r *etcdResolver) Close() {
	r.doneOnce.Do(func() {
		close(r.done)
		r.client.Close()
	})
}

func (r *etcdResolver) watch() {
	if r.hasClosed() {
		return
	}

	var (
		insts      = make(map[string]*instance.Instance)
		rev        int64
		retryTimes int
		watchCh    clientv3.WatchChan
	)

	for {
		cctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		resp, err := r.client.Get(cctx, r.key, clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Printf("[error]failed to resolve addr, caused by %s", err)
			delay := r.backoff(retryTimes)
			retryTimes++
			time.Sleep(delay)
			continue
		}

		retryTimes = 0
		rev = resp.Header.Revision

		for _, kv := range resp.Kvs {
			inst := instance.Instance{}
			inst.Decode(kv.Value)
			insts[string(kv.Key)] = &inst
		}

		addrs := r.insts2Addrs(insts)
		r.cc.UpdateState(resolver.State{
			Addresses: addrs,
		})
		break
	}

	if r.hasClosed() {
		return
	}

	watchCh = r.client.Watch(context.Background(), r.key, clientv3.WithPrefix(), clientv3.WithProgressNotify(), clientv3.WithRev(rev+1))

	for {
		select {
		case <-r.done:
			return

		case event := <-watchCh:
			if event.Canceled {
				if r.hasClosed() {
					watchCh = nil
				} else {
					delay := r.backoff(retryTimes)
					retryTimes++
					time.Sleep(delay)
					watchCh = r.client.Watch(context.Background(), r.key, clientv3.WithPrefix(), clientv3.WithProgressNotify(), clientv3.WithRev(rev+1))
				}
				continue
			}

			for _, ev := range event.Events {
				key := string(ev.Kv.Key)
				switch ev.Type {
				case mvccpb.PUT:
					inst := instance.Instance{}
					inst.Decode(ev.Kv.Value)
					insts[key ] = &inst
				case mvccpb.DELETE:
					delete(insts, key)
				}
			}
		}

		if retryTimes > 0 {
			retryTimes = 0
		}

		addrs := r.insts2Addrs(insts)

		r.cc.UpdateState(resolver.State{
			Addresses: addrs,
		})
	}
}

func (r *etcdResolver) insts2Addrs(insts map[string]*instance.Instance) []resolver.Address {
	addrs := make([]resolver.Address, 0, len(insts))
	for _, v := range insts {
		addr := resolver.Address{
			Addr:       fmt.Sprintf("%s:%d", v.Addr, v.Port),
			ServerName: v.AppID,
		}

		if r.mdConvert != nil && len(v.Metadata) > 0 {
			addr.Metadata = r.mdConvert(v.Metadata)
		}

		addrs = append(addrs, addr)
	}
	return addrs
}

func (r *etcdResolver) hasClosed() bool {
	select {
	case <-r.done:
		return true
	default:
	}
	return false
}
