package main

import (
	"context"
	"fmt"
	"log"
	"net"

	internal "rregist"
	pb "rregist/example/proto"

	"google.golang.org/grpc"
)

type SimpleServer struct {
}

var (
	endpoints = []string{"localhost:2379"}
	address   = "localhost:8080"
	network   = "tcp"
	serName   = "example"
)

func main() {
	listener, err := net.Listen(network, address)

	if err != nil {
		fmt.Println(err)
	}
	log.Println(address + " start")
	grpcSer := grpc.NewServer()
	pb.RegisterSimpleServer(grpcSer, &SimpleServer{})
	ser, err := internal.NewServiceRegistor(endpoints, serName+"/"+address, "4", 5)
	if err != nil {
		fmt.Println(err)
	}
	defer ser.Close()
	err = grpcSer.Serve(listener)
	if err != nil {
		fmt.Println(err)
	}

}

func (s *SimpleServer) GetKey(ctx context.Context, req *pb.SimpleRequest) (*pb.SimpleResponse, error) {
	log.Println(req.Key)
	return &pb.SimpleResponse{Val: req.Key}, nil
}
