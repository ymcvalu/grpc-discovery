package smooth_weighted

import (
	"encoding/json"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"strconv"
	"sync"
)

const Name = "smooth_weighted"
const WeightTag = "weight"
const defaultWeight = 0

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilderV2(Name, &smoothWeightPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type smoothWeightPickerBuilder struct {
}

func (*smoothWeightPickerBuilder) Build(info base.PickerBuildInfo) balancer.V2Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPickerV2(balancer.ErrNoSubConnAvailable)
	}

	p := smoothWeightPicker{}

	p.weightPeers = make([]weightPeer, 0, len(info.ReadySCs))
	for sc, info := range info.ReadySCs {
		w := defaultWeight

		switch md := info.Address.Metadata.(type) {
		case *map[string]string:
			w = weight(*md)
		case *instance.Metadata:
			w = weight(map[string]string(*md ))
		case string:
			w = weight(parseMetadata(md))
		}

		p.weightPeers = append(p.weightPeers, weightPeer{
			subConn: sc,
			weight:  w,
		})
	}

	return &p
}

type weightPeer struct {
	subConn         balancer.SubConn
	weight          int
	effectiveWeight int
	currentWeight   int
}

type smoothWeightPicker struct {
	weightPeers []weightPeer
	mu          sync.Mutex
}

func (p *smoothWeightPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.weightPeers) == 1 {
		return balancer.PickResult{SubConn: p.weightPeers[0].subConn}, nil
	}

	best := -1
	total := 0
	for i := 0; i < len(p.weightPeers); i++ {
		wp := &p.weightPeers[i]

		wp.currentWeight += wp.effectiveWeight
		total += wp.effectiveWeight
		if wp.effectiveWeight < wp.weight {
			wp.effectiveWeight++
		}
		if best == -1 || wp.currentWeight > p.weightPeers[best].currentWeight {
			best = i
		}
	}
	p.weightPeers[best].currentWeight -= total

	return balancer.PickResult{SubConn: p.weightPeers[best].subConn}, nil
}

func weight(md map[string]string) int {
	w := defaultWeight
	_w, err := strconv.Atoi(md[WeightTag])
	if err == nil {
		w = _w
	}
	return w
}

func parseMetadata(s string) map[string]string {
	md := make(map[string]string)
	json.Unmarshal([]byte(s), &md)
	return md
}
