### grpc 用 etcd 做服务注册与发现
./rpc/server.go 中是服务注册, 主要是利用 etcd Grant 创建一个租约,在租约上去创建或者更新一个 key(客户端根据 key 的变化来更新服务地址),并且通过 keepalive 方法续租.
./rpc/client.go 中是客户端服务发现, 主要是利用 etcd 的 Watch 方法,去监测某一个 key 下的值的变化.