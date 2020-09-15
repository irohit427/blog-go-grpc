package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/irohit427/go_grpc/blog/blog_pb"
	"google.golang.org/grpc"
)

type server struct {
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Staring Blog Server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blog_pb.RegisterBlogServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Ctrl+C to stop the server
	ch := make(chan os.Signal, 1)

	// Block until a signal is received
	<-ch
	fmt.Println("Stopping server")
	s.Stop()
	fmt.Println("Closing Listener")
	lis.Close()
	fmt.Println("Exit")
}
