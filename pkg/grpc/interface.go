package grpc

import (
	pb "github.com/keikoproj/manager/pkg/proto/cluster"
	"io"
)

type Interface interface {
	NewClusterClientOrDie() (io.Closer, pb.ClusterServiceClient)
}
