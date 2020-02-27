package engine

import (
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/util"
)

func (wi *WorkflowInstance) ExecuteWorkflow(handler handler.Handler) {
	kubeconfig := wi.Kubeconfig
	objName := wi.WFObjName
	currentStep := ""
	nextStep := ""
	fmt.Printf("Start Executing workflow %s\n", wi.Workflow.Name)
	startAt := wi.Workflow.StartAt
	currentStep = startAt
	emptyCondition := StepCondition{}
	for {
		util.AddStepToWorkflowObject(kubeconfig, currentStep, objName)
		fmt.Printf("Start Running Step %s\n", currentStep)
		wi.StepFunc[currentStep](handler)
		nextSteps := wi.Workflow.Steps[currentStep].NextSteps
		for _, step := range nextSteps {
			nextStepName := step.Name
			condition := step.Condition
			if condition == emptyCondition {
				nextStep = nextStepName
				break
			}
			key := condition.Key
			value := condition.Value
			operator := condition.Operator
			flowDataResult := handler.FlowData.Get(key)
			switch operator {
			case "=":
				if value == flowDataResult {
					nextStep = nextStepName
					break
				}
			}
		}
		util.SetWorkflowObjectStepToComplete(kubeconfig, objName, currentStep)
		currentStep = nextStep
		if len(wi.Workflow.Steps[currentStep].NextSteps) == 0 {
			break
		}
	}
	fmt.Printf("Executing workflow complete%s\n", wi.Workflow.Name)

}
