package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	pb "github.com/keikoproj/manager/pkg/proto/cluster"
	"github.com/keikoproj/manager/server/cluster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
	"net"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 10000, "The server port")
)

func main() {
	log.New()
	log := log.Logger(context.Background(), "main")

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Error(err, "failed to listen")
	}
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = testdata.Path("server1.pem")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Error(err, "Failed to generate credentials")
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)

	//Lets get k8s client here
	sClient := k8s.NewK8sSelfClientDoOrDie()
	pb.RegisterClusterServiceServer(grpcServer, cluster.New(sClient))
	log.Info("Server is up and running")
	grpcServer.Serve(lis)

}
