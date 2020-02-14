package workflow1

import (
	"workflow-engine/engine"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step1"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step2"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step3"
	"workflow-engine/workflow-example/workflows/workflow1/steps/step4"
)

func Definition() engine.Workflow {
	w := engine.ParseDefinition("/Users/gaoxindai/go/src/workflow-engine/workflow-example/workflows/workflow1/definition.json")
	return w
}

func Steps() map[string]func() string {
	StepFuncMap := map[string]func() string{
		"step1": step1.Execute,
		"step2": step2.Execute,
		"step3": step3.Execute,
		"step4": step4.Execute,
	}
	return StepFuncMap
}

//todo Trigger
func Trigger() string {
	return "test"
}
