package common

// Global constants
const (
	// systemNameSpace is the default k8s namespace i.e, kube-system
	SystemNameSpace = "kube-system"

	ManagerServiceAccountName = "keiko-manager-sa"

	ManagerClusterRole = "keiko-manager-cluster-role"

	ManagerClusterRoleBinding = "keiko-manager-cluster-role-binding"

	RBACApiVersion = "rbac.authorization.k8s.io/v1"

	ServiceAccountKind = "ServiceAccount"

	ClusterRoleKind = "ClusterRole"

	ClusterRoleBindingKind = "ClusterRoleBinding"

	InClusterAPIServerAddr = "https://kubernetes.default.svc"

	RoleKind = "Role"

	RoleBindingKind = "RoleBinding"

	ResourceQuotaKind = "ResourceQuota"

	CustomResourceKind = "CustomResource"

	ManagerDeployedNamespace = "manager-system"
)

const (
	PropertyClusterValidationFrequency = "cluster.validation.frequency"

	// ManagerNamespaceName is the namespace name where manager controllers are running
	ManagerNamespaceName = "manager-system"

	// ManagerConfigMapName is the config map name for manager namespace
	ManagerConfigMapName = "manager-v1alpha1-configmap"
)
