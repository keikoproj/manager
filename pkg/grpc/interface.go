package grpc

import (
	pb "github.com/keikoproj/manager/pkg/grpc/proto/cluster"
	"io"
)

type Interface interface {
	NewClusterClientOrDie() (io.Closer, pb.ClusterServiceClient)
}
