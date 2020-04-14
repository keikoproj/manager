package grpc

import (
	pb "github.com/keikoproj/manager/pkg/grpc/proto/apis"
	"io"
)

type Interface interface {
	NewClusterClientOrDie() (io.Closer, pb.ClusterServiceClient)
}
