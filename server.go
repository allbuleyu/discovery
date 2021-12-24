package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/allbuleyu/discovery/diss"
	"github.com/allbuleyu/discovery/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	port = flag.Int("p", 2000, "port of server")
)

func main() {
	flag.Parse()
	addr := "127.0.0.1:" + fmt.Sprintf("%d", *port)
	etcdAddrs := "127.0.0.1:2379"

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	r := rpc.NewRegister(addr, etcdAddrs)
	r.Register()

	srv := grpc.NewServer()
	reflection.Register(srv)

	diss.RegisterFirstDiscoveryServiceServer(srv, r)

	fmt.Println(srv.Serve(listener))
}
