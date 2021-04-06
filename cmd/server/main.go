package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/MobileStore-Grpc/image/pb"
	"github.com/MobileStore-Grpc/image/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func runRESTServer(reviewService pb.ImageServiceServer, listener net.Listener, grpcEndpoint string) error {
	mux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// in-process handler
	err := pb.RegisterImageServiceHandlerServer(ctx, mux, reviewService)
	if err != nil {
		return err
	}
	log.Printf("Start REST server at %s", listener.Addr().String())
	return http.Serve(listener, mux)
}

func runGRPCServer(reviewService pb.ImageServiceServer, listener net.Listener) error {
	grpcServer := grpc.NewServer()
	pb.RegisterImageServiceServer(grpcServer, reviewService)

	// Like we run server.ListenAnsServer(), similarly we do  grpcServer.serve()
	log.Printf("Start GRPC server at %s", listener.Addr().String())

	return grpcServer.Serve(listener)
}

func main() {
	port := flag.Int("port", 0, "the server port")
	serverType := flag.String("type", "grpc", "type of server (grpc/rest)")
	endPoint := flag.String("endpoint", "", "gRPC endpoint")
	flag.Parse()

	imageStore := service.NewDiskImageStore("/img")

	imageService := service.NewImageService(imageStore)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)

	if *serverType == "grpc" {
		err = runGRPCServer(imageService, listener)
	} else {
		err = runRESTServer(imageService, listener, *endPoint)
	}
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
