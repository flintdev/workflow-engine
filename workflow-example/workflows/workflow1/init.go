package workflow1

import (
	workflowFramework "github.com/flintdev/workflow-engine/engine"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/workflow-example/workflows/workflow1/steps/step1"
	"github.com/flintdev/workflow-engine/workflow-example/workflows/workflow1/steps/step2"
	"github.com/flintdev/workflow-engine/workflow-example/workflows/workflow1/steps/step3"
	"github.com/flintdev/workflow-engine/workflow-example/workflows/workflow1/steps/step4"
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
