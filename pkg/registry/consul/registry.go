package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
	"github.com/ymcvalu/grpc-discovery/pkg/registry"
	"log"
	"sync"
	"time"
)

type consulRegistry struct {
	mu       sync.Mutex
	insts    map[string]*instance.Instance
	doneOnce sync.Once
	done     chan struct{}
	client   *api.Client
	wg       sync.WaitGroup
}

func New(dc, addr string) (registry.Registry, error) {
	if dc == "" {
		dc = "dc1"
	}

	client, err := api.NewClient(&api.Config{
		WaitTime:   time.Second * 3,
		Datacenter: dc,
		Address:    addr,
	})
	if err != nil {
		return nil, err
	}

	r := &consulRegistry{
		insts:  make(map[string]*instance.Instance),
		done:   make(chan struct{}),
		client: client,
	}
	return r, nil
}

func (r *consulRegistry) Close() {
	r.doneOnce.Do(func() {
		close(r.done)
		r.wg.Wait()
	})
}

func (r *consulRegistry) Register(inst instance.Instance) <-chan error {
	errCh := make(chan error, 1)

	if dup := func() bool {
		r.mu.Lock()
		defer r.mu.Unlock()

		addr := fmt.Sprintf("%s:%d", inst.Addr, inst.Port)

		_, dup := r.insts[addr]
		if dup {
			return dup
		}

		r.insts[addr] = &inst
		return false
	}; dup() {
		errCh <- registry.ErrDupRegister
		return errCh
	}

	select {
	case <-r.done:
		errCh <- registry.ErrRegistryClosed
		return errCh
	default:
	}

	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		svcId := fmt.Sprintf("%s-%s-%s-%d", inst.Env, inst.AppID, inst.Addr, inst.Port)
		checkId := svcId

		err := r.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
			Kind:    api.ServiceKindTypical,
			ID:      svcId,
			Name:    fmt.Sprintf("%s/%s", inst.Env, inst.AppID),
			Address: inst.Addr,
			Port:    inst.Port,
			Meta:    inst.Metadata.ToMap(),
			Check: &api.AgentServiceCheck{
				CheckID: checkId,
				TTL:     "10s",
			},
		})

		if err != nil {
			errCh <- err
			return
		}

		tick := time.NewTicker(time.Second * 6)
		defer tick.Stop()
	loop:
		for {
			select {
			case <-r.done:
				r.client.Agent().ServiceDeregister(svcId)
				errCh <- registry.ErrRegistryClosed
				break loop
			case <-tick.C:
				err := r.client.Agent().UpdateTTL(checkId, "pass", "pass")
				if err != nil {
					log.Printf("[error] failed to upadte ttl, caused by: %s", err.Error())
				}
			}
		}
	}()

	return errCh
}
