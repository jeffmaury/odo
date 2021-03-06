package completion

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/service"
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/url"
	"github.com/openshift/odo/pkg/util"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// ServiceCompletionHandler provides service name completion for the current project and application
var ServiceCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	services, err := service.List(context.Client, context.Application)
	if err != nil {
		return completions
	}

	for _, class := range services.Items {
		if args.commands[class.ObjectMeta.Name] {
			return nil
		}
		completions = append(completions, class.ObjectMeta.Name)
	}

	return
}

// ServiceClassCompletionHandler provides catalog service class name completion
var ServiceClassCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	services, err := context.Client.GetKubeClient().ListClusterServiceClasses()
	if err != nil {
		complete.Log("error retrieving services")
		return completions
	}

	complete.Log(fmt.Sprintf("found %d services", len(services)))
	for _, class := range services {
		if args.commands[class.Spec.ExternalName] {
			return nil
		}
		completions = append(completions, class.Spec.ExternalName)
	}

	return
}

// ServicePlanCompletionHandler provides completion for the the plan of a selected service
var ServicePlanCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	// if we have less than two arguments, it means the user didn't specify the name of the service
	// meaning that there is no point in providing suggestions
	if len(args.original.Completed) < 2 {
		complete.Log("Couldn't extract the service name")
		return completions
	}

	inputServiceName := args.original.Completed[1]

	complete.Log(fmt.Sprintf("Using input: serviceName = %s", inputServiceName))

	clusterServiceClass, err := context.Client.GetKubeClient().GetClusterServiceClass(inputServiceName)
	if err != nil {
		complete.Log("Error retrieving details of service")
		return completions
	}

	servicePlans, err := context.Client.GetKubeClient().ListClusterServicePlansByServiceName(clusterServiceClass.Name)
	if err != nil {
		complete.Log("Error retrieving details of plans of service")
		return completions
	}

	for _, servicePlan := range servicePlans {
		completions = append(completions, servicePlan.Spec.ExternalName)
	}

	return completions
}

// ServiceParameterCompletionHandler provides completion for the parameter names of a selected service and plan
var ServiceParameterCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	if len(args.original.Completed) < 2 {
		complete.Log("Couldn't extract the service name")
		return completions
	}

	inputServiceName := args.original.Completed[1]
	inputPlanName := args.flagValues["plan"]

	complete.Log(fmt.Sprintf("Using input: serviceName = %s, servicePlan = %s", inputServiceName, inputPlanName))

	_, servicePlans, err := service.GetServiceClassAndPlans(context.Client, inputServiceName)
	if err != nil {
		complete.Log("Error retrieving details of service")
		return completions
	}

	var matchingServicePlan *service.ServicePlan = nil
	if len(servicePlans) == 0 {
		complete.Log("Service has no plans so no parameters can be found")
		return completions
	} else if len(servicePlans) == 1 && inputPlanName == "" {
		matchingServicePlan = &servicePlans[0]
	} else {
		for _, sp := range servicePlans {
			servicePlan := sp
			if servicePlan.Name == inputPlanName {
				matchingServicePlan = &servicePlan
				break
			}
		}
		if matchingServicePlan == nil {
			complete.Log("No service plan for the service matched the supplied plan name")
			return completions
		}
	}

	alreadyAddedParameters := args.flagValues["parameters"]
	for _, servicePlanParameter := range matchingServicePlan.Parameters {
		// don't add the parameter if it's already on the command line
		if !strings.Contains(alreadyAddedParameters, servicePlanParameter.Name) {
			completions = append(completions, servicePlanParameter.Name)
		}
	}

	return completions
}

// AppCompletionHandler provides completion for the app commands
var AppCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	applications, err := application.List(context.Client)
	if err != nil {
		return completions
	}

	for _, app := range applications {
		if args.commands[app] {
			return nil
		}
		completions = append(completions, app)
	}
	return
}

// FileCompletionHandler provides suggestions for files and directories
var FileCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = append(completions, complete.PredictFiles("*").Predict(args.original)...)
	return
}

// ProjectNameCompletionHandler provides project name completion
var ProjectNameCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	projects, err := context.Client.ListProjectNames()
	if err != nil {
		return completions
	}

	for _, project := range projects {
		// we found the project name in the list which means
		// that the project name has been already selected by the user so no need to suggest more
		if args.commands[project] {
			return nil
		}
		completions = append(completions, project)
	}
	return completions
}

// URLCompletionHandler provides completion for the url commands
var URLCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	urls, err := url.ListPushed(context.Client, context.Component(), context.Application)
	if err != nil {
		return completions
	}

	for _, url := range urls.Items {
		// we found the url in the list which means
		// that the url name has been already selected by the user so no need to suggest more
		if args.commands[url.Name] {
			return nil
		}
		completions = append(completions, url.Name)
	}
	return
}

// StorageDeleteCompletionHandler provides storage name completion for storage delete
var StorageDeleteCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	localConfig, err := config.New()
	if err != nil {
		return completions
	}

	storageList, err := localConfig.StorageList()
	if err != nil {
		return completions
	}

	for _, storage := range storageList {
		// we found the storage name in the list which means
		// that the storage name has been already selected by the user so no need to suggest more
		if args.commands[storage.Name] {
			return nil
		}
		completions = append(completions, storage.Name)
	}
	return completions
}

// StorageMountCompletionHandler provides storage name completion for storage mount
var StorageMountCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	storages, err := storage.ListUnmounted(context.Client, context.Application)
	if err != nil {
		return completions
	}

	for _, storage := range storages.Items {
		// we found the storage name in the list which means
		// that the storage name has been already selected by the user so no need to suggest more
		if args.commands[storage.Name] {
			return nil
		}
		completions = append(completions, storage.Name)
	}
	return completions
}

// StorageUnMountCompletionHandler provides storage name completion for storage unmount
var StorageUnMountCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	storageList, err := storage.ListMounted(context.Client, context.Component(), context.Application)
	if err != nil {
		return completions
	}

	for _, storage := range storageList.Items {
		// we found the storage name in the list which means
		// that the storage name has been already selected by the user so no need to suggest more
		if args.commands[storage.Name] {
			return nil
		}
		completions = append(completions, storage.Name)
	}
	return completions
}

// CreateCompletionHandler provides component type completion in odo create command
var CreateCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	comps := &completions
	found := false

	tasks := util.NewConcurrentTasks(2)
	tasks.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
		catalogList, _ := catalog.ListComponents(context.Client)
		for _, builder := range catalogList.Items {
			if args.commands[builder.Name] {
				found = true
				return
			}
			if len(builder.Spec.NonHiddenTags) > 0 {
				*comps = append(*comps, builder.Name)
			}
		}
	}})
	tasks.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
		components, _ := catalog.ListDevfileComponents("")
		for _, devfile := range components.Items {
			if args.commands[devfile.Name] {
				found = true
				return
			}
			*comps = append(*comps, devfile.Name)
		}
	}})

	_ = tasks.Run()
	if found {
		return nil
	}
	return completions
}

// LinkCompletionHandler provides completion for the odo link command
// The function returns both components and services
var LinkCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	components, err := component.GetComponentNames(context.Client, context.Application)
	if err != nil {
		return completions
	}

	services, err := service.List(context.Client, context.Application)
	if err != nil {
		return completions
	}

	completions = make([]string, 0, len(components)+len(services.Items))
	for _, component := range components {
		// we found the name in the list which means
		// that the name has been already selected by the user so no need to suggest more
		if val, ok := args.commands[component]; ok && val {
			return nil
		}
		// we don't want to show the selected component as a target for linking, so we remove it from the suggestions
		if component != context.Component() {
			completions = append(completions, component)
		}
	}

	for _, service := range services.Items {
		// we found the name in the list which means
		// that the name has been already selected by the user so no need to suggest more
		if val, ok := args.commands[service.ObjectMeta.Name]; ok && val {
			return nil
		}
		completions = append(completions, service.ObjectMeta.Name)
	}

	return completions
}

// LinkCompletionHandler provides completion for the odo unlink command
// The function returns both components and services
var UnlinkCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	// first we need to retrieve the current component
	comp, err := component.GetPushedComponent(context.Client, context.Component(), context.Application)
	if err != nil {
		return completions
	}

	components, err := component.GetComponentNames(context.Client, context.Application)
	if err != nil {
		return completions
	}

	services, err := service.List(context.Client, context.Application)
	if err != nil {
		return completions
	}

	completions = make([]string, 0, len(components)+len(services.Items))
	secretNames := comp.GetLinkedSecretNames()
	for _, component := range components {
		// we found the name in the list which means
		// that the name has been already selected by the user so no need to suggest more
		if val, ok := args.commands[component]; ok && val {
			return nil
		}
		// we don't want to show the selected component as a target for linking, so we remove it from the suggestions
		if component != context.Component() {
			// we also need to make sure that this component has been linked to the current component
			for _, secret := range secretNames {
				if strings.Contains(secret, component) {
					completions = append(completions, component)
				}
			}
		}
	}

	for _, service := range services.Items {
		// we found the name in the list which means
		// that the name has been already selected by the user so no need to suggest more
		if val, ok := args.commands[service.Name]; ok && val {
			return nil
		}
		// we also need to make sure that this component has been linked to the current component
		for _, secret := range secretNames {
			if strings.Contains(secret, service.Name) {
				completions = append(completions, service.Name)
			}
		}
	}

	return completions
}

// ComponentNameCompletionHandler provides component name completion
var ComponentNameCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	components, err := component.List(context.Client, context.Application, nil)

	if err != nil {
		return completions
	}

	for _, component := range components.Items {
		// we found the component name in the list which means
		// that the component name has been already selected by the user so no need to suggest more
		if args.commands[component.Name] {
			return nil
		}
		completions = append(completions, component.Name)
	}
	return completions
}
