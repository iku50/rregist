package rregist

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"rregist/balancer/weight"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
)

const schema = "etcdv3"

type ServiveDiscovery struct {
	client     *clientv3.Client
	clientConn resolver.ClientConn
	serverList sync.Map
	prefix     string
}

// NewServiceDiscovery 创建一个发现服务
func NewServiceDiscovery(endpoints []string) (*ServiveDiscovery, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &ServiveDiscovery{
		client: cli,
	}, nil
}

// Build 服务发现实现 resolver.Builder
func (s *ServiveDiscovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s.clientConn = cc
	s.prefix = "/" + target.Scheme + "/" + target.Endpoint + "/"
	// 获取所有服务
	resp, err := s.client.Get(s.client.Ctx(), s.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}
	s.clientConn.UpdateState(resolver.State{Addresses: s.GetServices()})
	// 监听前缀，修改服务列表
	go s.watcher()
	return s, nil
}

func (s *ServiveDiscovery) Scheme() string {
	return schema
}

func (s *ServiveDiscovery) ResolveNow(rn resolver.ResolveNowOptions) {
	log.Println("ResolveNow")
}

func (s *ServiveDiscovery) Close() {
	s.client.Close()
}

// 监听前缀
func (s *ServiveDiscovery) watcher() {
	rch := s.client.Watch(s.client.Ctx(), s.prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case clientv3.EventTypeDelete:
				s.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

// 设置服务列表
func (s *ServiveDiscovery) SetServiceList(key, val string) {
	addr := resolver.Address{Addr: strings.TrimPrefix(key, s.prefix)}
	weigh, err := strconv.Atoi(string(val))
	if err != nil {
		weigh = 1
	}
	addr = weight.SetAddrInfo(addr, weight.AddrInfo{Weight: weigh})
	s.serverList.Store(key, addr)
	s.clientConn.UpdateState(resolver.State{Addresses: s.GetServices()})
	log.Println("set data key:", key, "val:", val)
}

// 删除服务列表
func (s *ServiveDiscovery) DelServiceList(key string) {
	s.serverList.Delete(key)
	s.clientConn.UpdateState(resolver.State{Addresses: s.GetServices()})
	log.Println("del data key:", key)
}

// 获取服务列表
func (s *ServiveDiscovery) GetServices() []resolver.Address {
	addrs := make([]resolver.Address, 0, 10)
	s.serverList.Range(func(key, value interface{}) bool {
		addr := value.(resolver.Address)
		addrs = append(addrs, addr)
		return true
	})
	return addrs
}
