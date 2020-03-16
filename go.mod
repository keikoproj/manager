module github.com/keikoproj/manager

go 1.13

require (
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/go-logr/logr v0.1.0
	github.com/golang/mock v1.2.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/prometheus/common v0.4.1
	github.com/spf13/cobra v0.0.6
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.1
	sigs.k8s.io/controller-tools v0.2.5 // indirect
)
