package config

import (
	"context"
	"fmt"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"os"
)

var (
	Props *Properties
)

type Properties struct {
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

	//res := k8s.NewK8sClientDoOrDie().GetConfigMap(context.Background(), "", "")
	//
	//// load properties into a global variable
	//err := LoadProperties("", res)
	//if err != nil {
	//	log.Error(err, "failed to load properties")
	//	panic(err)
	//}
	log.Info("Loaded properties in init func")
}

func LoadProperties(env string, cm ...*v1.ConfigMap) error {
	log := log.Logger(context.Background(), "internal.config.properties", "LoadProperties")

	// for local testing
	if env != "" {
		Props = &Properties{}
		return nil
	}

	if len(cm) == 0 || cm[0] == nil {
		log.Error(fmt.Errorf("config map cannot be nil"), "config map cannot be nil")
		return fmt.Errorf("config map cannot be nil")
	}

	return nil
}

func RunConfigMapInformer(ctx context.Context) {
	log := log.Logger(context.Background(), "internal.config.properties", "RunConfigMapInformer")
	cmInformer := k8s.GetConfigMapInformer(ctx, "", "")
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
