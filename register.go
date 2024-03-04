package rregist

import (
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// 服务注册
type ServiceRegistor struct {
	client  *clientv3.Client
	leaseID clientv3.LeaseID

	// 租约 keepalive 监听
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	weight        string
}

// NewServiceRegistor 创建一个注册服务
func NewServiceRegistor(endpoints []string, addr, weight string, Lease int64) (*ServiceRegistor, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	// 创建租约
	service := &ServiceRegistor{
		client: client,
		key:    "/" + schema + "/" + addr,
		weight: weight,
	}
	// 设置租约时间
	if err := service.setLease(Lease); err != nil {
		return nil, err
	}
	log.Println("service " + addr + " register success!")
	return service, nil
}

// 设置租约时间
func (s *ServiceRegistor) setLease(Lease int64) error {
	leaseResp, err := s.client.Grant(s.client.Ctx(), Lease)
	if err != nil {
		return err
	}
	_, err = s.client.Put(s.client.Ctx(), s.key, s.weight, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	// 设置续租
	keepAliveChan, err := s.client.KeepAlive(s.client.Ctx(), leaseResp.ID)
	if err != nil {
		return err
	}
	s.keepAliveChan = keepAliveChan
	s.leaseID = leaseResp.ID
	log.Println(s.leaseID)
	return nil
}

// 监听续租情况
func (s *ServiceRegistor) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		if leaseKeepResp == nil {
			log.Println("keep alive channel closed")
			return
		} else {
			log.Println("Recv reply from service: ", leaseKeepResp.ID)
		}
	}
}

// 关闭服务
func (s *ServiceRegistor) Close() error {
	// 释放租约
	if _, err := s.client.Revoke(s.client.Ctx(), s.leaseID); err != nil {
		return err
	}
	log.Println("service" + s.key + "close success!")
	return s.client.Close()
}
