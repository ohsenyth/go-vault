package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"vault"
	"vault/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/juju/ratelimit"
	ratelimitkit "github.com/go-kit/kit/ratelimit"
)

func main() {
	var (
		httpAddr = flag.String("http", ":8080", "http listen address")
		gRPCAddr = flag.String("grpc", ":8081", "gRPC listen address")
	)
	flag.Parse()
	ctx := context.Background()
	srv := vault.NewService()
	errChan := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	rlbucket := ratelimit.NewBucket(1*time.Second, 5)

	hashEndpoint := vault.MakeHashEndpoint(srv)
	{
		hashEndpoint = ratelimitkit.NewTokenBucketThrottler (rlbucket, time.Sleep) (hashEndpoint)
	}
	validateEndpoint := vault.MakeValidateEndpoint(srv)
	{
		validateEndpoint = ratelimitkit.NewTokenBucketThrottler (rlbucket, time.Sleep) (validateEndpoint)
	}
	endpoints := vault.Endpoints{
		HashEndPoint: hashEndpoint,
		ValidateEndpoint: validateEndpoint
	}

	// HTTP transport
	go func() {
		log.Println("http:", *httpAddr)
		handler := vault.NewHTTPServer(ctx, endpoints)
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()

	go func() {
		listener, err := net.Listen("tcp", *gRPCAddr)

		if err != nil {
			errChan <- errChan
			return
		}
		log.Println("grpc:", *gRPCAddr)
		handler := vault.NewGRPCServer(ctx, endpoints)
		gRPCServer := grpc.NewServer()
		pb.RegisterVaultServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(listener)
	}()

	log.Fatalln(<-errChan)
}

