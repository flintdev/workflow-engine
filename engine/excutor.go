package engine

import "fmt"
import "workflow-engine/util"

func (wi *WorkflowInstance) ExecuteWorkflow(kubeconfig *string, namespace string, group string, version string, resource string) {
	//todo
	// 1. handle no condition match
	// 2. handle code execution exception
	currentStep := ""
	nextStep := ""
	fmt.Println("Executing workflow started")
	startAt := wi.Workflow.StartAt
	currentStep = startAt
	for {
		result := wi.StepFunc[currentStep]()
		nextSteps := wi.Workflow.Steps[currentStep].NextSteps
		for _, step := range nextSteps {
			value := step.Value
			operator := step.Operator
			stepLabels := step.StepLabels
			if operator == "=" {
				switch value {
				case "Always":
					nextStep = stepLabels[0]
					break
				default:
					if value == result {
						nextStep = stepLabels[0]
						break
					}
				}
			}
		}
		fmt.Println("current step: ", currentStep)
		fmt.Println("result: ", result)
		fmt.Println("next step: ", nextStep)
		util.UpdateSampleObj(kubeconfig, namespace, group, version, resource, currentStep, result, "complete")
		currentStep = nextStep
		if nextStep == "end" {
			break
		}
	}
	fmt.Println("Executing workflow complete")
}
