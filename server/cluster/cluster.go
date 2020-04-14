package cluster

import (
	"context"
	"fmt"
	"github.com/keikoproj/manager/api/v1alpha1"
	"github.com/keikoproj/manager/internal/config/common"
	"github.com/keikoproj/manager/internal/utils"
	apis "github.com/keikoproj/manager/pkg/grpc/proto/apis"
	pb "github.com/keikoproj/manager/pkg/grpc/proto/cluster"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
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

	// Create the secret
	s := make(map[string]string)

	s[fmt.Sprintf("%s_%s", name, "config")] = cl.Config.BearerToken
	secretName := fmt.Sprintf("%s-%s", name, "secrets")
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: common.ManagerDeployedNamespace,
		},
		StringData: s,
	}

	err := c.k8sClient.CreateOrUpdateK8sSecret(ctx, secret, common.ManagerDeployedNamespace)
	if err != nil {
		log.Error(err, "unable to create/update secret in the namespace", "name", name)
		return nil, err
	}

	//prepare cluster CR request i.e, just remove cl.Config.BearerToken and set cl.config.BearerTokenSecret
	cl.Config.BearerTokenSecret = secretName
	cl.Config.BearerToken = ""
	cr := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.SanitizeName(cl.Name),
			Namespace: common.ManagerDeployedNamespace,
		},
		Spec: v1alpha1.ClusterSpec{
			Cluster: *cl,
		},
	}
	err = c.k8sClient.CreateOrUpdateManagedCluster(ctx, cr, common.ManagerDeployedNamespace)
	if err != nil {
		log.Error(err, "unable to create/update cluster CR in the namespace", "name", name)
		return nil, err
	}

	return cl, nil
}

//UnregisterCluster unregisters the cluster with the server
func (c *clusterService) UnregisterCluster(ctx context.Context, req *apis.UnregisterClusterRequest) (*apis.UnregisterClusterResponse, error) {
	//Good thing is, we can just delete the respective namespace for that cluster and all the resources should be deleted
	//This should send the event to cluster controller implicitly and doesn't need to delete the cluster CR

	log := log.Logger(ctx, "server.cluster", "UnregisterCluster")
	log.Info("cluster name from the request", "name", req.ClusterName)
	name := utils.SanitizeName(req.ClusterName)
	log.V(1).Info("cluster name after sanitizing", "name", name)

	//Delete cluster CR
	cr := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.SanitizeName(req.ClusterName),
			Namespace: common.ManagerDeployedNamespace,
		},
	}
	err := c.k8sClient.DeleteManagedCluster(ctx, cr, common.ManagerDeployedNamespace)
	if err != nil {
		log.Error(err, "unable to delete the namespace", "name", name)
		return nil, err
	}
	return &apis.UnregisterClusterResponse{}, nil
}
