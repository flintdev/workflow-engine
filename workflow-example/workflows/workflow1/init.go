package workflow1

import (
	workflowFramework "workflow-engine/engine"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step1"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step2"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step3"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step4"
)

func Definition() workflowFramework.Workflow {

	//todo change to relative path
	w := workflowFramework.ParseDefinition("/Users/gaoxindai/go/src/workflow-engine/workflow-example/workflows/workflow1/definition.json")
	return w
}

func Steps() map[string]func(kubeconfig *string, objName string) {
	StepFuncMap := map[string]func(kubeconfig *string, objName string){
		"step1": step1.Execute,
		"step2": step2.Execute,
		"step3": step3.Execute,
		"step4": step4.Execute,
	}
	return StepFuncMap
}

func Trigger(event workflowFramework.Event) bool {
	return TriggerCondition(event)
}
