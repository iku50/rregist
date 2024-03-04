package main

import (
	"context"
	"fmt"
	"log"
	internal "rregist"
	pb "rregist/example/proto"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

var (
	// ServiceName is the name of the service
	ServiceName = "example"
	grpcClient  pb.SimpleClient
)

func main() {
	r, err := internal.NewServiceDiscovery([]string{"localhost:2379"})

	if err != nil {
		log.Fatal(err)
	}
	resolver.Register(r)
	conn, err := grpc.Dial(fmt.Sprintf("%s:///%s", r.Scheme(), ServiceName), grpc.WithInsecure(), grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"weight"}`))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	grpcClient = pb.NewSimpleClient(conn)
	for i := 0; i < 100; i++ {
		route(i)
		time.Sleep(1 * time.Second)
	}
}

func route(i int) {
	req := pb.SimpleRequest{
		Key: "grpc " + strconv.Itoa(i),
	}
	resp, err := grpcClient.GetKey(context.Background(), &req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp)
}
