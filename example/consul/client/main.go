package main

import (
	"context"
	"github.com/ymcvalu/grpc-discovery/example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"log"
	"time"
)

func main() {

	conn, err := grpc.Dial("consul://127.0.0.1:8500/dev/echo", grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name), grpc.WithBlock())
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
