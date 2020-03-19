package cluster

import (
	"context"
	"github.com/keikoproj/manager/pkg/log"
	pb "github.com/keikoproj/manager/pkg/proto/cluster"
)
type clusterServiceServer struct {
}

func New() *clusterServiceServer{
	return &clusterServiceServer{}
}

//RegisterCluster handles registering cluster with the controller
func (c *clusterServiceServer)RegisterCluster(ctx context.Context, cl *pb.Cluster) (*pb.Cluster, error) {
	//Following are the list of actions to do
	// 1. Validate the Request
	// 2. Create Namespace based on the cluster name if doesn't exists(or idempotent)
	// 3. Extract BearerToken and create a secret in respective namespace
	// 4. Copy the cluster request to controller cluster struct
	// 5. Create cluster custom resource in the respective namespace

	log := log.Logger(ctx, "server.cluster", "RegisterCluster")
	log.Info("Request received")
	log.Info("cluster name from the request", "name", cl.Name)
	return cl, nil
}