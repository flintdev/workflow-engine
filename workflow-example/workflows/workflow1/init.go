package workflow1

import (
	workflowFramework "workflow-engine/engine"
	"workflow-engine/handler"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step1"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step2"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step3"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step4"
)

func Definition() workflowFramework.Workflow {
	w := ParseDefinition()
	return w
}

func Steps() map[string]func(handler handler.Handler) {
	StepFuncMap := map[string]func(handler handler.Handler){
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

type FilePath string
