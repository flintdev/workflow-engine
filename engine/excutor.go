package engine

import (
	"fmt"
	"workflow-engine/handler"
	"workflow-engine/util"
)

func (wi *WorkflowInstance) ExecuteWorkflow(kubeconfig *string, objName string) {
	//todo
	// 1. handle no condition match
	// 2. handle code execution exception
	currentStep := ""
	nextStep := ""
	fmt.Println("Executing workflow started")
	startAt := wi.Workflow.StartAt
	currentStep = startAt
	for {
		util.AddStepToWorkflowObject(kubeconfig, currentStep, objName)
		wi.StepFunc[currentStep](kubeconfig, objName)
		nextSteps := wi.Workflow.Steps[currentStep].NextSteps
		for _, step := range nextSteps {
			emptyCondition := StepCondition{}
			nextStepName := step.Name
			condition := step.Condition
			if condition == emptyCondition {
				nextStep = nextStepName
				break
			}
			key := condition.Key
			value := condition.Value
			operator := condition.Operator
			flowDataResult := handler.GetFlowDataByPath(kubeconfig, objName, key)
			fmt.Println("successfully get value of ", key, flowDataResult)
			switch operator {
			case "=":
				if value == flowDataResult {
					nextStep = nextStepName
					break
				}
			}
		}
		fmt.Println("current step: ", currentStep)
		fmt.Println("next step: ", nextStep)
		util.SetWorkflowObjectStepToComplete(kubeconfig, objName, currentStep)
		currentStep = nextStep
		if nextStep == "end" {
			break
		}
	}
	fmt.Println("Executing workflow complete")
}
