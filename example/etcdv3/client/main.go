package main

import (
	"context"
	"github.com/ymcvalu/grpc-discovery/example/proto"
	"github.com/ymcvalu/grpc-discovery/pkg/resolver/etcdv3"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"log"
	"time"
)

func main() {
	etcdv3.Init(clientv3.Config{
		Endpoints:   []string{"192.168.50.12:2379"},
		DialTimeout: time.Second * 5,
	})

	conn, err := grpc.Dial("etcd:///dev/echo", grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}

	client := proto.NewEchoSvcClient(conn)
	for {
		time.Sleep(time.Second)
		resp, err := client.Echo(context.Background(), &proto.EchoReq{})
		if err != nil {
			log.Println(err)
		} else {
			log.Println(resp.Msg)
		}
	}
}
