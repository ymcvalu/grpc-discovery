package main

import (
	"context"
	"github.com/ymcvalu/grpc-discovery/example/proto"
	"github.com/ymcvalu/grpc-discovery/pkg/resolver/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"log"
	"time"
)

func main() {
	consul.Init("http://127.0.0.1:8500", consul.WithDataCenter("dc1"))

	conn, err := grpc.Dial("consul:///dev/echo", grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
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
