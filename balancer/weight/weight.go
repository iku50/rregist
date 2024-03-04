package weight

import (
	"sync"

	"math/rand"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const Name = "weight"

var (
	minWeight = 1
	maxWeight = 100
)

// attributeKey
type attributeKey struct{}

// AddrInfo
type AddrInfo struct {
	Weight int
}

// SetAddrInfo
func SetAddrInfo(addr resolver.Address, info AddrInfo) resolver.Address {
	addr.Attributes = attributes.New()
	addr.Attributes = addr.Attributes.WithValues(attributeKey{}, info)
	return addr
}

// GetAddrInfo
func GetAddrInfo(addr resolver.Address) (info AddrInfo, ok bool) {
	info, ok = addr.Attributes.Value(attributeKey{}).(AddrInfo)
	return
}

// newBuilder
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilderV2(Name, &weightPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type weightPickerBuilder struct{}

func (*weightPickerBuilder) Build(info base.PickerBuildInfo) balancer.V2Picker {
	var scs []balancer.SubConn
	for sc, addr := range info.ReadySCs {
		node, ok := GetAddrInfo(addr.Address)
		if !ok {
			node.Weight = 1
		}
		if node.Weight < minWeight {
			node.Weight = minWeight
		}
		if node.Weight > maxWeight {
			node.Weight = maxWeight
		}
		for i := 0; i < node.Weight; i++ {
			scs = append(scs, sc)
		}
	}
	return &weightPicker{
		subConns: scs,
	}
}

type weightPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
}

func (p *weightPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	index := rand.Intn(len(p.subConns))
	sc := p.subConns[index]
	p.mu.Unlock()
	return balancer.PickResult{SubConn: sc}, nil
}
