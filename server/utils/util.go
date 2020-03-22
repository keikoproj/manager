package utils

import (
	"fmt"
	"github.com/keikoproj/manager/api/custom/v1alpha1"
	"github.com/keikoproj/manager/pkg/proto/cluster"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//SantitizeName sanitizes the string name based on K8s naming convention.
//
func SantitizeName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}

//PrepareClusterRequestFromClusterProto pretty much copies the value from cluster grpc struct to cluster controller struct
func PrepareClusterRequestFromClusterProto(cl *cluster.Cluster) *v1alpha1.Cluster {

	cr := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: SantitizeName(cl.Name),
			Name:      SantitizeName(cl.Name),
		},
		Spec: v1alpha1.ClusterSpec{
			Name:  SantitizeName(cl.Name),
			Cloud: cl.Cloud,
			Config: v1alpha1.Config{
				Host:              cl.Config.Host,
				BearerTokenSecret: fmt.Sprintf("%s-%s", SantitizeName(cl.Name), "secrets"),
				TLSClientConfig: v1alpha1.TLSClientConfig{
					Insecure:   cl.Config.TlsClientConfig.InSecure,
					ServerName: cl.Config.TlsClientConfig.ServerName,
					CAData:     cl.Config.TlsClientConfig.CaData,
				},
			},
		},
	}

	return cr
}
