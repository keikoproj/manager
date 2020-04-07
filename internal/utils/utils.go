package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/manager/api/custom/v1alpha1"
	"github.com/keikoproj/manager/pkg/grpc/proto/cluster"
	"github.com/keikoproj/manager/pkg/log"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//StopIfError is a convenient function to stop processing if there is any error
func StopIfError(err error) {
	if err != nil {
		fmt.Printf("error %v", err)
		os.Exit(1)
	}
}

//SanitizeName sanitizes the string name based on K8s naming convention.
//
func SanitizeName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}

//PrepareClusterRequestFromClusterProto pretty much copies the value from cluster grpc struct to cluster controller struct
func PrepareClusterRequestFromClusterProto(cl *cluster.Cluster) *v1alpha1.Cluster {

	cr := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: SanitizeName(cl.Name),
			Name:      SanitizeName(cl.Name),
		},
		Spec: v1alpha1.ClusterSpec{
			Name:  SanitizeName(cl.Name),
			Cloud: cl.Cloud,
			Config: v1alpha1.Config{
				Host:              cl.Config.Host,
				BearerTokenSecret: fmt.Sprintf("%s-%s", SanitizeName(cl.Name), "secrets"),
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

//PrepareK8sRestConfigFromClusterCR
func PrepareK8sRestConfigFromClusterCR(ctx context.Context, cr *v1alpha1.Cluster, secret *v1.Secret) (*rest.Config, error) {
	log := log.Logger(ctx, "internal.utils", "PrepareK8sRestConfigFromClusterCR")
	token, ok := secret.Data[fmt.Sprintf("%s_%s", SanitizeName(cr.Spec.Name), "config")]
	if !ok {
		msg := "bearer token doesn't exist"
		err := errors.New(msg)
		log.Error(err, msg)
		return nil, err
	}

	conf := &rest.Config{
		Host:        cr.Spec.Config.Host,
		BearerToken: string(token),
		TLSClientConfig: rest.TLSClientConfig{
			CAData:     cr.Spec.Config.CAData,
			ServerName: cr.Spec.Config.ServerName,
			Insecure:   cr.Spec.Config.Insecure,
		},
	}
	return conf, nil
}

//ContainsString  Helper functions to check from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

//RemoveString Helper function to check remove string
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

//BoolValue converts string value to bool
func BoolValue(flag string) bool {
	if strings.EqualFold(flag, "true") {
		return true
	}

	if strings.EqualFold(flag, "false") {
		return false
	}
	return false
}
