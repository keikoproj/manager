package validation

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	"github.com/keikoproj/manager/pkg/log"
)

var (
	uniqueNameErr             = "resource Names must be unique across the template. %s repeated more than once"
	nonExistDependsOnValueErr = "%s resource DependsOn value referring to a value %s which doesn't exist"
	circularDependencyErr     = "circular dependency is not allowed for resource dependsOn property"
)

const (
	noDependency = "-151726"
)

//ValidateTemplate function validates following
//1. Resource Names must be unique
//2. DependsOn value belongs to one of the resource in the same template
//3. DependsOn Circular Dependency
func ValidateTemplate(ctx context.Context, resources *namespace.NamespaceResources) error {
	log := log.Logger(ctx, "pkg.validation", "ValidateDependsOn")

	dependsMap := make(map[string]string)
	for _, res := range resources.Resources {
		//Lets make sure the names are unique too
		if _, ok := dependsMap[res.Name]; ok {
			// This shouldn't happen because it supposed to be unique
			err := errors.New(fmt.Sprintf(uniqueNameErr, res.Name))
			log.Error(err, fmt.Sprintf(uniqueNameErr, res.Name))
			return err
		}
		if res.DependsOn != "" {
			dependsMap[res.Name] = res.DependsOn
		} else {
			dependsMap[res.Name] = noDependency
		}
	}

	//Check the namespace name to be unique
	if _, ok := dependsMap[resources.Namespace.Name]; ok {
		err := errors.New(fmt.Sprintf(uniqueNameErr, resources.Namespace.Name))
		log.Error(err, fmt.Sprintf(uniqueNameErr, resources.Namespace.Name))
		return err
	}

	//DependsOn value belongs to one of the resource in the same template
	for k, v := range dependsMap {
		//No dependency case
		if v == noDependency {
			continue
		}
		//It must be present in the map
		if _, ok := dependsMap[v]; !ok {
			// Throw error
			err := errors.New(fmt.Sprintf(nonExistDependsOnValueErr, k, v))
			log.Error(err, fmt.Sprintf(nonExistDependsOnValueErr, k, v))
			return err
		}
		//Circular dependency
		//May be map or set
		visited := make(map[string]bool)
		x := k
		for i := 0; i <= len(dependsMap); i++ {
			if x == noDependency {
				break
			}
			if visited[x] {
				//I got the circular dependency
				// Throw error
				err := errors.New(circularDependencyErr)
				log.Error(err, circularDependencyErr)
				return err
			} else {
				visited[x] = true
				x = dependsMap[x]
			}
		}
	}
	return nil
}
