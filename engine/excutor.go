package engine

import (
	"encoding/json"
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/util"
	"io/ioutil"
	"net/http"
)

type ExecutorResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (wi *WorkflowInstance) ExecuteWorkflow(handler handler.Handler, wfObjName string) {
	kubeconfig := wi.Kubeconfig
	currentStep := ""
	nextStep := ""
	fmt.Printf("Start Executing workflow %s\n", wi.Workflow.Name)
	startAt := wi.Workflow.StartAt
	currentStep = startAt
	emptyCondition := StepCondition{}
	for {
		fmt.Printf("Start Running Step %s\n", currentStep)
		util.SetWorkflowObjectCurrentStep(kubeconfig, wfObjName, currentStep)
		util.SetWorkflowObjectCurrentStepLabel(kubeconfig, wfObjName, currentStep)
		util.SetStepToWorkflowObject(kubeconfig, currentStep, wfObjName)
		if wi.Workflow.Steps[currentStep].Type == "manual" {
			util.SetWorkflowObjectStepToPending(kubeconfig, wfObjName, currentStep)
			stepTrigger := wi.Workflow.Steps[currentStep].StepTrigger
			stepTrigger.StepName = currentStep
			wi.StepTriggers = append(wi.StepTriggers, stepTrigger)
			fmt.Println("current step triggers", wi.StepTriggers)
			break
		}
		port := get_python_executor_port()
		url := fmt.Sprintf("http://127.0.0.1:%s/execute?workflow=%s&step=%s&obj_name=%s", port, wi.Workflow.Name, currentStep, wfObjName)
		fmt.Println(url)
		response, err := http.Get(url)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
			break
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			r := ParseExecutorResponse(data)
			if r.Status == "success" {
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
				util.SetWorkflowObjectStepToComplete(kubeconfig, wfObjName, currentStep)
				currentStep = nextStep
				if len(wi.Workflow.Steps[currentStep].NextSteps) == 0 {
					break
				}
			} else {
				break
			}
		}

	}
	fmt.Printf("Executing workflow complete %s\n", wi.Workflow.Name)

}

func (wi *WorkflowInstance) ExecutePendingWorkflow(handler handler.Handler, wfObjName string, currentStep string) {
	kubeconfig := wi.Kubeconfig
	nextStep := ""
	fmt.Printf("Start Executing workflow %s\n", wi.Workflow.Name)
	emptyCondition := StepCondition{}
	addStep := false
	for {
		fmt.Println(currentStep)
		fmt.Println(nextStep)
		fmt.Printf("Start Running Step %s\n", currentStep)
		util.SetWorkflowObjectCurrentStep(kubeconfig, wfObjName, currentStep)
		util.SetWorkflowObjectStepToRunning(kubeconfig, wfObjName, currentStep)
		util.SetWorkflowObjectCurrentStepLabel(kubeconfig, wfObjName, currentStep)
		if addStep {
			util.SetStepToWorkflowObject(kubeconfig, currentStep, wfObjName)
		}
		addStep = true
		port := get_python_executor_port()
		url := fmt.Sprintf("http://127.0.0.1:%s/execute?workflow=%s&step=%s&obj_name=%s", port, wi.Workflow.Name, currentStep, wfObjName)
		fmt.Println(url)
		response, err := http.Get(url)
		if err != nil {
			fmt.Printf("The HTTP request failed with error %s\n", err)
			break
		} else {
			data, _ := ioutil.ReadAll(response.Body)
			r := ParseExecutorResponse(data)
			if r.Status == "success" {
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
				util.SetWorkflowObjectStepToComplete(kubeconfig, wfObjName, currentStep)
				currentStep = nextStep
				if len(wi.Workflow.Steps[currentStep].NextSteps) == 0 {
					break
				}
			} else {
				break
			}
		}

	}
	fmt.Printf("Executing workflow complete %s\n", wi.Workflow.Name)

}

//func (wi *WorkflowInstance) ExecuteWorkflowWithBreakPoint(handler handler.Handler, currentStep string) {
//	kubeconfig := wi.Kubeconfig
//	currentStep := ""
//	nextStep := ""
//	fmt.Printf("Start Executing workflow %s\n", wi.Workflow.Name)
//	startAt := wi.Workflow.StartAt
//	currentStep = startAt
//	emptyCondition := StepCondition{}
//	for {
//		fmt.Printf("Start Running Step %s\n", currentStep)
//		util.SetWorkflowObjectCurrentStep(kubeconfig, objName, currentStep)
//		util.SetStepToWorkflowObject(kubeconfig, currentStep, objName)
//		if wi.Workflow.Steps[currentStep].Type == "manual" {
//			util.SetWorkflowObjectStepToPending(kubeconfig, objName, currentStep)
//			wi.StepTriggers = append(wi.StepTriggers, wi.Workflow.Steps[currentStep].StepTrigger)
//			break
//		}
//		port := get_python_executor_port()
//		url := fmt.Sprintf("http://127.0.0.1:%s/execute?workflow=%s&step=%s&obj_name=%s", port, wi.Workflow.Name, currentStep, objName)
//		fmt.Println(url)
//		response, err := http.Get(url)
//		if err != nil {
//			fmt.Printf("The HTTP request failed with error %s\n", err)
//			break
//		} else {
//			data, _ := ioutil.ReadAll(response.Body)
//			r := ParseExecutorResponse(data)
//			if r.Status == "success" {
//				nextSteps := wi.Workflow.Steps[currentStep].NextSteps
//				for _, step := range nextSteps {
//					nextStepName := step.Name
//					condition := step.Condition
//					if condition == emptyCondition {
//						nextStep = nextStepName
//						break
//					}
//					key := condition.Key
//					value := condition.Value
//					operator := condition.Operator
//					flowDataResult := handler.FlowData.Get(key)
//					switch operator {
//					case "=":
//						if value == flowDataResult {
//							nextStep = nextStepName
//							break
//						}
//					}
//				}
//				util.SetWorkflowObjectStepToComplete(kubeconfig, objName, currentStep)
//				currentStep = nextStep
//				if len(wi.Workflow.Steps[currentStep].NextSteps) == 0 {
//					break
//				}
//			} else {
//				break
//			}
//		}
//
//	}
//	util.SetWorkflowObjectCurrentStep(kubeconfig, objName, "end")
//	fmt.Printf("Executing workflow complete %s\n", wi.Workflow.Name)
//
//}

func ParseExecutorResponse(body []byte) ExecutorResponse {
	var r ExecutorResponse
	err := json.Unmarshal(body, &r)

	if err != nil {
		fmt.Println(err.Error())
	}
	return r
}

func get_python_executor_port() string {
	data, _ := ioutil.ReadFile("/tmp/flint_python_executor_port")
	port := string(data)
	return port
}
