package cluster

import (
	"context"
	"fmt"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	pb "github.com/keikoproj/manager/pkg/proto/cluster"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type clusterService struct {
	k8sClient *k8s.Client
}

func New(sClient *k8s.Client) *clusterService {
	return &clusterService{
		k8sClient: sClient,
	}
}

//RegisterCluster handles registering cluster with the controller
func (c *clusterService) RegisterCluster(ctx context.Context, cl *pb.Cluster) (*pb.Cluster, error) {
	//Following are the list of actions to do
	// 1. Validate the Request -- This should be done at the end once i finalize the proto
	// 2. Create Namespace based on the cluster name if doesn't exists(or idempotent)
	// 3. Extract BearerToken and create a secret in respective namespace
	// 4. Copy the cluster request to controller cluster struct
	// 5. Create cluster custom resource in the respective namespace
	log := log.Logger(ctx, "server.cluster", "RegisterCluster")
	log.Info("Request received")
	log.Info("cluster name from the request", "name", cl.Name)
	name := utils.SanitizeName(cl.Name)
	log.V(1).Info("cluster name after sanitizing", "name", name)

	//Create the namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	err := c.k8sClient.CreateNamespace(ctx, ns)
	if err != nil {
		log.Error(err, "unable to create namespace", "name", name)
		return nil, err
	}

	// Create the secret
	s := make(map[string]string)

	s[fmt.Sprintf("%s_%s", name, "config")] = cl.Config.BearerToken

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", name, "secrets"),
			Namespace: name,
		},
		StringData: s,
	}

	err = c.k8sClient.CreateK8sSecret(ctx, secret, name)
	if err != nil {
		log.Error(err, "unable to create secret in the namespace", "name", name)
		return nil, err
	}

	//prepare cluster CR request
	cr := utils.PrepareClusterRequestFromClusterProto(cl)
	err = c.k8sClient.CreateOrUpdateClusterCR(ctx, cr)
	if err != nil {
		log.Error(err, "unable to create/update cluster CR in the namespace", "name", name)
		return nil, err
	}

	return cl, nil
}
