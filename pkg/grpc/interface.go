package grpc

import (
	"io"
	pb "github.com/keikoproj/manager/pkg/proto/cluster"
)


type Interface interface {
	NewClusterClientOrDie()(io.Closer, pb.ClusterServiceClient)
}
