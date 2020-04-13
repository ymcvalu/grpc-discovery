package main

import (
	"context"
	"github.com/ymcvalu/grpc-discovery/example/proto"
	"github.com/ymcvalu/grpc-discovery/pkg/balancer/smooth_weighted"
	_ "github.com/ymcvalu/grpc-discovery/pkg/resolver/etcdv3"
	"google.golang.org/grpc"
	"log"
)

func main() {
	log.Println("client begin...")
	// etcd://username:password@etcd1,etcd2,etcd3?timeout=5s/dev/app
	conn, err := grpc.Dial("etcd://192.168.50.10:2379?timeout=5s/dev/echo", grpc.WithInsecure(), grpc.WithBalancerName(smooth_weighted.Name), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}

	client := proto.NewEchoSvcClient(conn)
	for {
		resp, err := client.Echo(context.Background(), &proto.EchoReq{})
		if err != nil {
			log.Println(err)
		} else {
			log.Println(resp.Msg)
		}
	}
}
