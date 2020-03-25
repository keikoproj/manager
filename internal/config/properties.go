package config

import (
	"context"
	"fmt"
	"github.com/keikoproj/manager/internal/config/common"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"os"
	"strconv"
)

var (
	Props *Properties
)

type Properties struct {
	clusterValidationFrequency int
}

func init() {
	log := log.Logger(context.Background(), "internal.config.properties", "init")

	if os.Getenv("LOCAL") != "" {
		err := LoadProperties("LOCAL")
		if err != nil {
			log.Error(err, "failed to load properties for local environment")
			return
		}
		log.Info("Loaded properties in init func for tests")
		return
	}

	res := k8s.NewK8sSelfClientDoOrDie().GetConfigMap(context.Background(), common.ManagerNamespaceName, common.ManagerConfigMapName)

	// load properties into a global variable
	err := LoadProperties("", res)
	if err != nil {
		log.Error(err, "failed to load properties")
		panic(err)
	}
	log.Info("Loaded properties in init func")
}

func LoadProperties(env string, cm ...*v1.ConfigMap) error {
	log := log.Logger(context.Background(), "internal.config.properties", "LoadProperties")
	Props = &Properties{}
	// for local testing
	if env != "" {

		return nil
	}

	if len(cm) == 0 || cm[0] == nil {
		log.Error(fmt.Errorf("config map cannot be nil"), "config map cannot be nil")
		return fmt.Errorf("config map cannot be nil")
	}

	ClusterValidationFrequency := cm[0].Data[common.PropertyClusterValidationFrequency]
	if ClusterValidationFrequency != "" {
		ClusterValidationFrequency, err := strconv.Atoi(ClusterValidationFrequency)
		if err != nil {
			return err
		}
		Props.clusterValidationFrequency = ClusterValidationFrequency
	} else {
		Props.clusterValidationFrequency = 1800
	}

	return nil
}

func (p *Properties) ClusterValidationFrequency() int {
	return p.clusterValidationFrequency
}

func RunConfigMapInformer(ctx context.Context) {
	log := log.Logger(context.Background(), "internal.config.properties", "RunConfigMapInformer")
	cmInformer := k8s.GetConfigMapInformer(ctx, common.ManagerNamespaceName, common.ManagerConfigMapName)
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: updateProperties,
	},
	)
	log.Info("Starting config map informer")
	cmInformer.Run(ctx.Done())
	log.Info("Cancelling config map informer")
}

func updateProperties(old, new interface{}) {
	log := log.Logger(context.Background(), "internal.config.properties", "updateProperties")
	oldCM := old.(*v1.ConfigMap)
	newCM := new.(*v1.ConfigMap)
	if oldCM.ResourceVersion == newCM.ResourceVersion {
		return
	}
	log.Info("Updating config map", "new revision ", newCM.ResourceVersion)
	err := LoadProperties("", newCM)
	if err != nil {
		log.Error(err, "failed to update config map")
	}
}
