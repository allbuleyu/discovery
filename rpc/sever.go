package rpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/allbuleyu/discovery/diss"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	scheme      = "etcd"
	serviceName = "discovery"
)

type register struct {
	addr        string
	etcdAddrs   string
	scheme      string
	serviceName string

	cli     *clientv3.Client
	leaseId clientv3.LeaseID

	diss.UnimplementedFirstDiscoveryServiceServer
}

func NewRegister(addr, etcdAddrs string) *register {
	return &register{
		addr:        addr,
		etcdAddrs:   etcdAddrs,
		scheme:      scheme,
		serviceName: serviceName,
	}
}

func (s *register) Hello(ctx context.Context, in *diss.HelloRequest) (*diss.HelloResponse, error) {
	return &diss.HelloResponse{
		Addr: s.addr,
	}, nil
}

func (s *register) FallInLove(ctx context.Context, in *diss.FallInLoveRequest) (*diss.FallInLoveResponse, error) {
	return &diss.FallInLoveResponse{}, nil
}

func (s *register) etcdKey() string {
	return "/" + s.scheme + "/" + s.serviceName + "/" + s.addr
}

func (s *register) Register() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(s.etcdAddrs, ";"),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("etcd connection err:", err)
		return
	}

	s.cli = cli

	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	resp, err := cli.Grant(ctx, 60)
	if err != nil {
		fmt.Println("grant lease err:", err)
		return
	}
	s.leaseId = resp.ID

	_, err = cli.Put(ctx, s.etcdKey(), s.addr, clientv3.WithLease(s.leaseId))
	if err != nil {
		fmt.Println("put key err:", err)
		return
	}

	go s.keepAlive()
}

func (s *register) keepAlive() {
	respCh, err := s.cli.KeepAlive(context.Background(), s.leaseId)
	if err != nil {
		fmt.Println("keep alive err:", err)
		return
	}

	startTime := time.Now()
	for range respCh {
		keepAliveTime := time.Now()
		fmt.Printf("%s leaseId: %v time diff: %v \n", s.etcdKey(), s.leaseId, keepAliveTime.Sub(startTime))
		startTime = keepAliveTime
	}
}
