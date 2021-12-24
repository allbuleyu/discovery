package rpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

type myClient struct {
	etcdAddrs   string
	scheme      string
	serviceName string

	// grpc 服务器地址集合
	serverAddrs map[string]resolver.Address

	// etcd 客户端
	cli *clientv3.Client
	cc  resolver.ClientConn
}

func NewClient(etcdAddrs, scheme, serviceName string) *myClient {
	return &myClient{
		etcdAddrs:   etcdAddrs,
		scheme:      scheme,
		serviceName: serviceName,
		serverAddrs: make(map[string]resolver.Address),
	}
}

func (c *myClient) Scheme() string {
	return c.scheme
}

func (c *myClient) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(c.etcdAddrs, ";"),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	c.cli = cli
	c.cc = cc

	// fmt.Printf("scheme: %s, endpoint: %s, endpoint1: %s \n", target.URL.Scheme, target.URL.Path, target.URL.Opaque)

	go c.watch()

	return c, nil
}

func (c *myClient) ResolveNow(o resolver.ResolveNowOptions) {
	fmt.Println("resolver now")
}

func (c *myClient) Close() {
	fmt.Println("close now")
}

func (c *myClient) watch() {
	resp, err := c.cli.Get(context.Background(), c.getPrefix(), clientv3.WithPrefix())
	if err != nil {
		fmt.Println("get prefix err:", err)
		return
	}

	if len(resp.Kvs) == 0 {
		fmt.Println("there is no server online")
		return
	}

	addrs := make([]resolver.Address, len(resp.Kvs))
	for i := 0; i < len(resp.Kvs); i++ {
		addrs[i] = resolver.Address{Addr: string(resp.Kvs[i].Value)}
		c.serverAddrs[string(resp.Kvs[i].Key)] = addrs[i]
	}

	err = c.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		fmt.Printf("UpdateState err:%v \n", err)
		return
	}

	// 监听给定前缀key的变化
	watchCh := c.cli.Watch(context.Background(), c.getPrefix(), clientv3.WithPrefix())
	for watchResp := range watchCh {
		// 处理key 变化事件
		for _, event := range watchResp.Events {
			key := string(event.Kv.Key)
			switch event.Type {
			case mvccpb.PUT:
				if _, ok := c.serverAddrs[key]; !ok {
					c.serverAddrs[key] = resolver.Address{
						Addr: string(event.Kv.Value),
					}
				}
			case mvccpb.DELETE:
				if _, ok := c.serverAddrs[key]; ok {
					delete(c.serverAddrs, key)
				}
			}
		}

		c.cc.UpdateState(resolver.State{Addresses: c.getServerAddrs()})
	}
}

func (c *myClient) getPrefix() string {
	return "/" + c.scheme + "/" + c.serviceName + "/"
}

func (c *myClient) getServerAddrs() []resolver.Address {
	addrs := make([]resolver.Address, 0, len(c.serverAddrs))
	for _, addr := range c.serverAddrs {
		addrs = append(addrs, addr)
	}
	return addrs
}
