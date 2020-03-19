package engine

import (
	"encoding/json"
	"errors"
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

func executeStep(kubeconfig *string, wi *WorkflowInstance, handler handler.Handler, wfObjName string, currentStep string) (bool, string, error) {
	emptyCondition := StepCondition{}
	nextStep := ""
	port := get_python_executor_port()
	url := fmt.Sprintf("http://127.0.0.1:%s/execute?workflow=%s&step=%s&obj_name=%s", port, wi.Workflow.Name, currentStep, wfObjName)
	fmt.Printf("Sent GET request to %s\n", url)
	response, err := http.Get(url)
	if err != nil {
		message := fmt.Sprintf("The HTTP request failed with error %s\n", err)
		return false, "", errors.New(message)
	} else {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return false, "", err
		}
		r, err := ParseExecutorResponse(data)
		if err != nil {
			return false, "", err
		}
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
				flowDataResult, err := handler.FlowData.Get(key)
				if err != nil {
					return false, "", err
				}
				switch operator {
				case "=":
					if value == flowDataResult {
						nextStep = nextStepName
						break
					}
				}
			}
			err := util.SetWorkflowObjectStepToComplete(kubeconfig, wfObjName, currentStep)
			if err != nil {
				return false, "", err
			}
			if len(wi.Workflow.Steps[nextStep].NextSteps) == 0 {
				return true, "", nil
			} else {
				return false, nextStep, nil
			}
		} else {
			return false, "", errors.New(r.Message)
		}
	}
}

func (wi *WorkflowInstance) ExecuteWorkflow(kubeconfig *string, handler handler.Handler, wfObjName string) {
	currentStep := ""
	fmt.Printf("Start Executing workflow %s\n", wi.Workflow.Name)
	startAt := wi.Workflow.StartAt
	currentStep = startAt
	for {
		fmt.Printf("Start Running Step %s\n", currentStep)
		err := util.SetWorkflowObjectStep(kubeconfig, wfObjName, currentStep)
		if err != nil {
			//todo add message to wf object
		}
		if wi.Workflow.Steps[currentStep].Type == "manual" {
			err := util.SetWorkflowObjectStepToPending(kubeconfig, wfObjName, currentStep)
			if err != nil {
				//todo add message to wf object
			}
			stepTrigger := wi.Workflow.Steps[currentStep].StepTrigger
			stepTrigger.StepName = currentStep
			wi.StepTriggers = append(wi.StepTriggers, stepTrigger)
			fmt.Printf("Workflow is pending on manual step %s\n", currentStep)
			break
		}
		isComplete, nextStep, err := executeStep(kubeconfig, wi, handler, wfObjName, currentStep)
		currentStep = nextStep
		if err != nil {
			//todo add message to wf object
		}
		if isComplete {
			fmt.Printf("Workflow %s Complete\n", wi.Workflow.Name)
			break
		}
	}
}

func (wi *WorkflowInstance) ExecutePendingWorkflow(kubeconfig *string, handler handler.Handler, wfObjName string, currentStep string) {
	fmt.Printf("Start Executing workflow %s\n", wi.Workflow.Name)
	skipFirst := false
	for {
		fmt.Printf("Start Running Step %s\n", currentStep)
		err := util.SetWorkflowObjectStepToRunning(kubeconfig, wfObjName, currentStep)
		if err != nil {
			//todo add message to wf object
		}
		if skipFirst {
			err := util.SetWorkflowObjectStep(kubeconfig, wfObjName, currentStep)
			if err != nil {
				//todo add message to wf object
			}
			if wi.Workflow.Steps[currentStep].Type == "manual" {
				err := util.SetWorkflowObjectStepToPending(kubeconfig, wfObjName, currentStep)
				if err != nil {
					//todo add message to wf object
				}
				stepTrigger := wi.Workflow.Steps[currentStep].StepTrigger
				stepTrigger.StepName = currentStep
				wi.StepTriggers = append(wi.StepTriggers, stepTrigger)
				break
			}
		}
		skipFirst = true
		isComplete, nextStep, err := executeStep(kubeconfig, wi, handler, wfObjName, currentStep)
		if err != nil {
			//todo add message to wf object
		}
		currentStep = nextStep
		if isComplete {
			fmt.Printf("Workflow %s Complete\n", wi.Workflow.Name)
			break
		}

	}
}

func ParseExecutorResponse(body []byte) (ExecutorResponse, error) {
	var r ExecutorResponse
	err := json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func get_python_executor_port() string {
	data, _ := ioutil.ReadFile("/tmp/flint_python_executor_port")
	port := string(data)
	return port
}
