package engine

import (
	"encoding/json"
	"fmt"
	"github.com/flintdev/workflow-engine/handler"
	"github.com/flintdev/workflow-engine/util"
	"io/ioutil"
	"net/http"
)

const executorEndpoint = "http://127.0.0.1:8080/execute?"

type ExecutorResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

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
		executorEndpointParams := fmt.Sprintf("workflow=%s&step=%s&obj_name=%s", wi.Workflow.Name, currentStep, objName)
		url := executorEndpoint + executorEndpointParams
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
				util.SetWorkflowObjectStepToComplete(kubeconfig, objName, currentStep)
				currentStep = nextStep
				if len(wi.Workflow.Steps[currentStep].NextSteps) == 0 {
					break
				}
			} else {
				break
			}
		}

	}
	fmt.Printf("Executing workflow complete%s\n", wi.Workflow.Name)

}

func ParseExecutorResponse(body []byte) ExecutorResponse {
	var r ExecutorResponse
	err := json.Unmarshal(body, &r)

	if err != nil {
		fmt.Println(err.Error())
	}
	return r
}
