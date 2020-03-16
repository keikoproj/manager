package commands

import (
	"context"
	"fmt"
	"github.com/keikoproj/manager/internal/config/common"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

// NewClusterCommand returns a new instance of an `manager cluster` command
func NewClusterCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "cluster",
		Short: "Manage cluster operations",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
			os.Exit(1)
		},
		Example: `  # Register a cluster with keiko-manager. The context must exist in your kubectl config:
  manager cluster register -ctx admins@iksm-ppd-usw2-k8s
  #	Remove managed cluster from manager
  manager cluster unregister -c admins@iksm-ppd-usw2-k8s
`,
	}

	command.AddCommand(NewClusterRegisterCommand())
	command.AddCommand(NewClusterUnregisterCommand())
	return command
}

//NewClusterRegisterCommand registers the target cluster with the manager.
//Target cluster can be the same cluster where manager resides (provide --self true)
//If service account is not provided manager will create service account, cluster role and role binding
//Service Account must exists in "kube-system" namespace
func NewClusterRegisterCommand() *cobra.Command {
	var (
		self           bool
		serviceAccount string
		configContext  string
	)

	var command = &cobra.Command{
		Use:     "register",
		Short:   fmt.Sprintf("%s cluster register", "manager"),
		Long:    "Add/register managed cluster with manager",
		Example: "manager cluster register -c admins@iksm-ppd-usw2-k8s",
		Run: func(c *cobra.Command, args []string) {
			ctx := context.Background()
			clientSet := getManagedClusterKubeConfig(configContext)
			managedClusterClient := k8s.NewK8sManagedClusterClientDoOrDie(clientSet)
			if serviceAccount == "" {
				createRBACInManagedCluster(ctx, managedClusterClient)
				serviceAccount = common.ManagerServiceAccountName
			}
			_, err := managedClusterClient.GetServiceAccountTokenSecret(ctx, serviceAccount, common.SystemNameSpace)
			utils.StopIfError(err)
			fmt.Printf("token received successfully\n")
		},
	}

	command.Flags().BoolVarP(&self, "self", "i", false, "To self manage keiko manager cluster itself. Default = false")
	command.Flags().StringVarP(&serviceAccount, "service-account", "s", "", fmt.Sprintf("System namespace service account to use for kubernetes resource management. If not set then default \"%s\" SA will be created", common.ManagerServiceAccountName))
	command.Flags().StringVarP(&configContext, "use-context", "c", "", "context to be used from user kubeconfig file. This kubeconfig context must have cluster admin access to create required RBAC in the target cluster if service account is not provided")

	return command
}

//NewClusterUnregisterCommand unregisters the target cluster from manager
//For now, lets make sure user who is unregistering the cluster does have cluster admin access on target cluster.
//This should be updated once we have RBAC on manager itself to see if user is authorized unregister a particular cluster
//At the moment, we concentrate only on removing the created rbac resources in the target clusters.
func NewClusterUnregisterCommand() *cobra.Command {
	var (
		configContext string
	)

	var command = &cobra.Command{
		Use:     "unregister",
		Short:   fmt.Sprintf("%s cluster unregister", "manager"),
		Long:    "Remove/unregister managed cluster from manager",
		Example: "manager cluster unregister -c admins@iksm-ppd-usw2-k8s",
		Run: func(c *cobra.Command, args []string) {

			ctx := context.Background()
			clientSet := getManagedClusterKubeConfig(configContext)
			managedClusterClient := k8s.NewK8sManagedClusterClientDoOrDie(clientSet)
			removeRBACInManagedCluster(ctx, managedClusterClient)
		},
	}

	command.Flags().StringVarP(&configContext, "use-context", "c", "", "context to be used from user kubeconfig file. This kubeconfig context must have cluster admin access to create required RBAC in the target cluster if service account is not provided")

	return command
}

func getManagedClusterKubeConfig(contextName string) *kubernetes.Clientset {

	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	utils.StopIfError(err)
	clstContext := config.Contexts[contextName]
	if clstContext == nil {
		log.Fatalf("Context %s does not exist in kubeconfig", contextName)
	}

	overrides := clientcmd.ConfigOverrides{
		Context: *clstContext,
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &overrides)
	conf, err := clientConfig.ClientConfig()
	utils.StopIfError(err)

	clientset, err := kubernetes.NewForConfig(conf)
	utils.StopIfError(err)
	return clientset
}

func createRBACInManagedCluster(ctx context.Context, client *k8s.Client) {
	//Create ServiceAccount
	err := client.CreateServiceAccount(ctx, common.ManagerServiceAccountName, common.SystemNameSpace)
	utils.StopIfError(err)

	//Create Cluster Role
	err = client.CreateOrUpdateClusterRole(ctx, common.ManagerClusterRole)
	utils.StopIfError(err)

	sub := v1.Subject{
		Kind:      common.ServiceAccountKind,
		Name:      common.ManagerServiceAccountName,
		Namespace: common.SystemNameSpace,
	}
	//Create Cluster RoleBinding
	err = client.CreateOrUpdateClusterRoleBinding(ctx, common.ManagerClusterRoleBinding, common.ManagerClusterRole, sub)
	utils.StopIfError(err)
}

func removeRBACInManagedCluster(ctx context.Context, client *k8s.Client) {
	//Delete Cluster RoleBinding
	err := client.DeleteClusterRoleBinding(ctx, common.ManagerClusterRoleBinding)
	utils.StopIfError(err)

	//Delete Cluster Role
	err = client.DeleteClusterRole(ctx, common.ManagerClusterRole)
	utils.StopIfError(err)

	//Delete Service Account
	err = client.DeleteServiceAccount(ctx, common.ManagerServiceAccountName, common.SystemNameSpace)
	utils.StopIfError(err)
}
