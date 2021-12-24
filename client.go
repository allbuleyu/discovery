package main

import (
	"context"
	"fmt"
	"time"

	"github.com/allbuleyu/discovery/diss"
	"github.com/allbuleyu/discovery/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

func main() {
	scheme := "etcd"
	serviceName := "discovery"
	etcdAddrs := "127.0.0.1:2379"
	b := rpc.NewClient(etcdAddrs, scheme, serviceName)
	resolver.Register(b)

	target := scheme + ":///" + serviceName
	cc, err := grpc.Dial(target, grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if err != nil {
		fmt.Println("grpc connection err:", err)
		return
	}
	defer cc.Close()

	c := diss.NewFirstDiscoveryServiceClient(cc)
	ticker := time.NewTicker(time.Second * 5)

	for tt := range ticker.C {
		ctx, cancle := context.WithTimeout(context.Background(), time.Second*3)
		fmt.Println("hello 调用")

		resp, err := c.Hello(ctx, &diss.HelloRequest{
			HelloWord: "hello",
			Custom:    make(map[string]string),
		})
		if err != nil {
			cancle()
			fmt.Println("response err:", err)
			return
		}

		fmt.Println(resp.Addr, tt)
	}
}
