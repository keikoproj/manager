package template

import (
	"context"
	"encoding/json"
	managerv1alpha1 "github.com/keikoproj/manager/api/custom/v1alpha1"
	"github.com/keikoproj/manager/pkg/log"
	"github.com/keikoproj/manager/pkg/validation"
	"strings"
)

//Supported types
//Validate types from exportedParams
//Based on type replace
//Decide either to convert the template to Go Template or write our own template

//ProcessTemplate function is an utility function to replace the template exported fields with dynamic values
func ProcessTemplate(ctx context.Context, template *managerv1alpha1.NamespaceTemplate, nsReq *managerv1alpha1.ManagedNamespace) error {
	log := log.Logger(ctx, "pkg.template", "template", "ExecuteTemplate")

	//Validate Namespace Template
	if err := validation.ValidateTemplate(ctx, template.Spec.NsResources); err != nil {
		return err
	}
	//Template comes up with exported params
	//For each exported param, replace it with the cluster config map
	//For each exported param, replace it with values from nsReq
	//

	//Marshal it template to a string
	tempBytes, err := json.Marshal(template.Spec.NsResources)
	if err != nil {
		return err
	}
	templateString := string(tempBytes)

	log.V(1).Info("Exported params", "count", len(template.Spec.ExportedParamName))
	//Replace it from the namespace request
	for _, param := range template.Spec.ExportedParamName {
		templateString = strings.ReplaceAll(templateString, "${"+param+"}", nsReq.Spec.Params[param])
	}

	//log.V(1).Info("template ", "temp", templateString)

	//Unmarshal it back
	err = json.Unmarshal([]byte(templateString), template.Spec.NsResources)
	if err != nil {
		return err
	}

	//This case is when no additional resources are included in the namespace request
	if nsReq.Spec.NsResources == nil {
		nsReq.Spec.NsResources = template.Spec.NsResources
	}

	//TODO: should handle the use case where additional resources are included

	//Validate Namespace Template
	if err := validation.ValidateTemplate(ctx, nsReq.Spec.NsResources); err != nil {
		return err
	}
	return nil
}
