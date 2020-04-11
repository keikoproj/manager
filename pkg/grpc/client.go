package grpc

import (
	"flag"
	"fmt"
	pb "github.com/keikoproj/manager/pkg/grpc/proto/apis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
	"log"
)

var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

type grpcClient struct {
	conn *grpc.ClientConn
}

//NewConnectionOrDie function gets the new grpc client
func NewConnectionOrDie() *grpcClient {
	fmt.Println("Request received successfully")
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v \n", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to create grpc connection %v \n", err)
	}
	fmt.Println("Connection Established successfully")
	return &grpcClient{conn: conn}
}

//NewClusterClientOrDie function returns cluster client
func (client *grpcClient) NewClusterClientOrDie() pb.ClusterServiceClient {
	fmt.Println("Cluster client created successfully")
	return pb.NewClusterServiceClient(client.conn)
}
