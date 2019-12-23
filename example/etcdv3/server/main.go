package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ymcvalu/grpc-discovery/example/proto"
	"github.com/ymcvalu/grpc-discovery/pkg/instance"
	"github.com/ymcvalu/grpc-discovery/pkg/registry/etcdv3"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var InstanceID = os.Getenv("INSTANCE_ID")

type EchoServer struct{}

func (EchoServer) Echo(ctx context.Context, req *proto.EchoReq) (resp *proto.EchoResp, err error) {
	log.Printf("handle grpc req...")
	return &proto.EchoResp{
		Msg: InstanceID,
	}, err
}

func main() {
	port := flag.Int("port", 6060, "port")
	flag.Parse()

	r, err := etcdv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.50.12:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	proto.RegisterEchoSvcServer(s, &EchoServer{})

	errCh := r.Register(instance.Instance{
		Env:   "dev",
		AppID: "echo",
		Addr:  "127.0.0.1",
		Port:  *port,
	})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		select {
		case <-sig:
			r.Close()
			s.GracefulStop()
			os.Exit(0)
		case err := <-errCh:
			log.Fatalf("failed to register: %s", err.Error())
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
